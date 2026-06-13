## Context

The `notification-publisher-hooks` change (commit `73c7c6f`) shipped a generic
`HookRegistrar` with one `Fire(ctx, msg)` call placed once before fan-out in
`handleBusMessage` (`publisher.go:341`). In practice the publisher needs *specific*
interception points, not one general one. The relevant points cluster around the per-chat
delivery loop (`publisher.go:355-375`):

- the duplicate check (`n.dupeCache.Has(dupKey)`, line 361) — keyed per `chatId`,
- the send (`n.tgbot.Send(...)`, line 368).

The driving use case is expensive shared enrichment: run a raw weather alert through an LLM
to summarize/rewrite it before delivery. Today the weather poller renders the message and
the dedup decision is made later, *per recipient*, on a raw-content key
(`getDedupeKey` → `location_event_end`), independent of the rendered/LLM text. That ordering
is fortunate: we can dedup on the cheap raw key first and only spend LLM tokens on content
that survives. But the existing loop interleaves dedup-check and send per recipient, so
there is no point that is both "after we know it's fresh" **and** "once for all
recipients." A per-recipient hook would multiply LLM cost by the subscriber count; a
pre-fan-out hook would spend tokens on pure duplicates.

This change reworks the existing hook code (it is not net-new) to expose three named points
and reshapes `handleBusMessage` into phases so a once-per-message, post-dedup enrichment
point exists.

## Goals / Non-Goals

**Goals:**
- Three named hook points — Predupe, AfterDedupe, OnSend — registered by topic pattern.
- Predupe/OnSend fire per recipient with `chatId`; AfterDedupe fires once per message with
  a mutable `*Message`.
- AfterDedupe runs only when at least one recipient is fresh, enabling expensive shared
  enrichment (LLM summarization) that is skipped entirely for pure duplicates.
- Cancellation of a single recipient via `ErrCancelNotification` (Predupe/OnSend); whole-
  message cancellation via the same sentinel from AfterDedupe.
- Unknown hook errors are non-fatal: Predupe/OnSend log and skip the recipient; AfterDedupe
  logs and fails open (delivers the unenriched message).

**Non-Goals:**
- Predupe participating in the dedupe decision (it is observe/cancel only).
- A generic pre-fan-out (per-message-before-dedup) firing point — removed.
- Hook priorities/ordering guarantees beyond registration order, or async hooks.
- Changing subscriber matching in `getChatIdsForTopic` or the dedup key scheme.
- Memoizing AfterDedupe output across separate message-handling invocations (so a late
  subscriber re-triggers enrichment) — a possible future optimization, not built here.
- The specifics of the LLM client/provider — that is the weather hook consumer's concern;
  this change defines the AfterDedupe contract it plugs into.

## Decisions

### Three named registrations on one registrar

Rework the interface to hold three independent hook lists rather than one:

```go
// Per-recipient hooks: observe/cancel, non-mutating.
type Hook func(ctx context.Context, msg Message, chatId int64) error

// Once-per-message hook: may mutate the shared message.
type EnrichHook func(ctx context.Context, msg *Message) error

type HookRegistrar interface {
    RegisterPreDupe(topicPattern string, hook Hook)
    RegisterAfterDedupe(topicPattern string, hook EnrichHook)
    RegisterOnSend(topicPattern string, hook Hook)
    FirePreDupe(ctx context.Context, msg Message, chatId int64) error
    FireAfterDedupe(ctx context.Context, msg *Message) error
    FireOnSend(ctx context.Context, msg Message, chatId int64) error
}
```

The concrete `hookRegistrar` keeps three `[]hookEntry`-style slices (`preDupe`,
`afterDedupe`, `onSend`) guarded by the existing `sync.RWMutex`, preserving registration
order and allowing multiple hooks per pattern. Each `Fire*` copies its slice under the read
lock (the deadlock-safe pattern already established), then invokes matching hooks in order,
returning the first non-nil error.

**Alternative considered:** separate registrar interfaces/fields per point. Rejected — one
registrar with register/fire triples keeps wiring (`WithHookRegistrar`, the single `hooks`
field) unchanged and mirrors the existing structure.

### Two signatures: per-recipient `Hook` and once-per-message `EnrichHook`

The old `Hook` was `func(ctx, msg) error` because it fired once per message. Predupe and
OnSend are per-recipient, so they take `chatId`. AfterDedupe is fundamentally different:
it fires **once** for the message and its job is to **rewrite the shared body**, so it
takes `*Message` and no `chatId`. Mixing these into one signature would force AfterDedupe
to either lie about granularity or be unable to mutate — hence the separate `EnrichHook`
type. This per-point signature divergence is intentional and must not be collapsed by
reflex.

### Keep anchored topic matching

Reuse the existing `compileHookPattern` helper (tokenize → `regexp.QuoteMeta` → detokenize,
wrapped `^...$`). The anchored behavior was a deliberate, tested decision in the prior
change and there is no reason to change it.

### Three-phase `handleBusMessage`: partition → enrich → send

The single interleaved per-chat loop is split into three phases so a once-per-message,
post-dedup point can exist. The per-recipient dupe key is computed once in the partition
phase and reused in the send phase (so enriching `msg.Msg` cannot shift the key):

```go
// Phase 1 — partition: build the fresh recipient set.
type target struct{ chatId int64; dupKey string }
fresh := make([]target, 0, len(chatIds))
for _, chatId := range chatIds {
    if err := n.hooks.FirePreDupe(ctx, msg, chatId); err != nil {
        if errors.Is(err, ErrCancelNotification) { continue }
        logger.Error().Err(err).Msg("predupe hook error; skipping recipient")
        continue
    }
    dupKey := fmt.Sprintf("%d-%s", chatId, msg.DuplicationKey())
    if n.dupeCache.Has(dupKey) { continue }
    fresh = append(fresh, target{chatId, dupKey})
}

// Phase 2 — enrich: once, only if at least one recipient is fresh.
if len(fresh) > 0 {
    if err := n.hooks.FireAfterDedupe(ctx, &msg); err != nil {
        if errors.Is(err, ErrCancelNotification) {
            return nil // whole message cancelled
        }
        // fail open: log and deliver the original unenriched message
        logger.Error().Err(err).Msg("afterdedupe hook error; sending unenriched")
    }
}

// Phase 3 — send: per fresh recipient, on the (possibly enriched) message.
for _, t := range fresh {
    if err := n.hooks.FireOnSend(ctx, msg, t.chatId); err != nil {
        if errors.Is(err, ErrCancelNotification) { continue }
        logger.Error().Err(err).Msg("onsend hook error; skipping recipient")
        continue
    }
    // existing send + n.dupeCache.Set(t.dupKey, ...) ...
}
```

Cancel/error semantics per point:
- **Predupe** — `ErrCancelNotification` or unknown error → skip that recipient (it never
  enters the fresh set, so it is neither enriched-for nor sent).
- **AfterDedupe** — `ErrCancelNotification` → drop the whole message (return). Unknown
  error → **fail open**: log and continue with the original `msg`.
- **OnSend** — `ErrCancelNotification` or unknown error → skip the send for that recipient.

The single pre-fan-out `Fire` call at line 341 is removed.

**Edge case (accepted):** if AfterDedupe runs (fresh set non-empty) but every fresh
recipient is then cancelled at OnSend, the enrichment work is wasted. This is rare and not
worth guarding against.

### Fail-open for AfterDedupe; non-fatal for all (changed from the prior design)

The prior change used `logger.Fatal()` for unknown hook errors, which AGENTS.md flags as
init-only and which would crash the bot mid-delivery. Predupe/OnSend unknown errors now log
at Error and skip the recipient. AfterDedupe is **fail-open** specifically because the
driving use case is *safety alerts*: an LLM timeout or error must not suppress a weather
warning, so the publisher delivers the original rendered alert instead. This is a deliberate
behavior change from the superseded change.

## Risks / Trade-offs

- **Per-recipient firing multiplies Predupe/OnSend invocations** (N recipients → N calls)
  → Acceptable; those hooks are lightweight and topic-scoped. AfterDedupe deliberately
  fires once to avoid multiplying expensive enrichment.
- **Synchronous hooks block the single bus goroutine** → A slow AfterDedupe (e.g. a 2–5s
  LLM call) serializes all delivery while it runs, since `handleBusMessage` processes one
  message at a time. Acceptable at weather-bot volume; async remains a non-goal.
- **Late-subscriber re-enrichment** → Because dedup is per-recipient and AfterDedupe output
  is not memoized across invocations, a subscriber who joins after an alert already went out
  triggers one more LLM call (the others are duplicates). Acceptable; memoization is a noted
  future optimization, not built here.
- **Supersedes a 15/16 change whose code is live** → This change rewrites that code in
  place; the old `notification-publisher-hooks` change should be archived/abandoned
  afterward so the two are not read as parallel features. Non-blocking for this proposal.
