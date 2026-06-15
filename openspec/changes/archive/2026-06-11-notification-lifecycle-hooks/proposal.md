## Why

The just-landed hook system (commit `73c7c6f`, change `notification-publisher-hooks`)
exposes a single generic `HookRegistrar.Fire` that runs all topic-matched hooks once
before fan-out. That generality is the wrong starting point: the real need is specific,
well-defined interception points in the delivery pipeline. A generic catch-all hook gives
callers no contract about *when* in the lifecycle they run and forces every consumer to
re-derive intent.

A concrete driving use case makes the requirement sharp: we want to run a raw weather-API
alert through an **LLM to summarize/rewrite it before sending**, but only for messages that
are **not duplicates**, and **once** for all recipients rather than once per recipient.
Doing this in the weather module would spend LLM tokens on every poll regardless of dedup;
doing it per recipient would multiply the cost by the subscriber count. The pipeline needs
a point that fires *after* dedup is known and *once* per message, with a **mutable**
message so the enriched body is shared by all fresh recipients.

## What Changes

- **BREAKING** Replace the generic single-`Fire` hook model with **three named lifecycle
  hook points** on the notification publisher:
  - **Predupe** — fires *per recipient* immediately *before* the dedupe-cache lookup.
    Observe/cancel only: it may inspect the message and cancel delivery to that recipient,
    but does not alter dedup behavior.
  - **AfterDedupe** — fires **once per message**, after the publisher knows at least one
    recipient is fresh (not a duplicate) and before any send. Receives a **mutable**
    message so it can enrich/rewrite the body exactly once for all fresh recipients. This
    is the point for expensive, shared enrichment (e.g. LLM summarization of a weather
    alert) that must not run for pure duplicates and must not run per recipient.
  - **OnSend** — fires *per recipient* immediately *before* the Telegram send, on the
    already-enriched message.
- **Reshape `handleBusMessage`** from a single interleaved per-chat loop into three phases:
  1. *Partition* (per recipient): run Predupe and the dupe-cache check to build the set of
     fresh recipients.
  2. *Enrich* (once): if the fresh set is non-empty, run AfterDedupe on the message.
  3. *Send* (per recipient over the fresh set): run OnSend, send, and record the dupe cache.
- Predupe and OnSend fire **per recipient** and receive the resolved `chatId` in addition
  to the message; AfterDedupe fires **once** and receives a `*Message` (no `chatId`).
- All three hooks keep **topic-pattern registration** with the existing `*` wildcard
  matching, so a hook only runs for topics it registered for.
- Predupe and OnSend may **cancel** delivery to the current recipient by returning the
  existing `ErrCancelNotification` sentinel; the publisher then skips that recipient and
  continues. AfterDedupe cancellation drops the message for **all** recipients. Any other
  (non-sentinel) error from Predupe/OnSend is logged and that recipient is skipped; a
  non-sentinel error from AfterDedupe is **fail-open** — logged, with the original
  unenriched message delivered. No hook error crashes the bot.
- **Remove** the generic `Register`/`Fire` API, the per-message pre-fan-out firing point,
  and the anchored-vs-unanchored single matcher in favor of the three named registrations.

This supersedes the `notification-publisher-hooks` change; that change's generic API is
replaced wholesale rather than extended.

## Capabilities

### New Capabilities
- `notification-hooks`: Three named lifecycle hooks on the notification publisher,
  registered by topic pattern — two per-recipient (Predupe, OnSend) that can cancel
  delivery to a recipient via a sentinel error, and one once-per-message (AfterDedupe) that
  fires after dedup is known and may mutate the shared message body before send.

### Modified Capabilities
<!-- None: no capability spec has been archived to openspec/specs/ yet. -->

## Impact

- `pkg/notifications/hooks.go`: rework the `HookRegistrar` interface and concrete registrar
  to three named hook lists — `RegisterPreDupe`/`RegisterOnSend`/`RegisterAfterDedupe` and
  `FirePreDupe`/`FireOnSend`/`FireAfterDedupe`. Predupe/OnSend use a `chatId`-carrying
  `Hook` signature; AfterDedupe uses an `EnrichHook` signature taking `*Message`.
- `pkg/notifications/publisher.go`: remove the single pre-fan-out `Fire` call; reshape
  `handleBusMessage` into partition → enrich → send phases. `FirePreDupe` + dupe check
  build the fresh set; `FireAfterDedupe(ctx, &msg)` runs once when the fresh set is
  non-empty; `FireOnSend` + send + `dupeCache.Set` run per fresh recipient. The per-recipient
  dupe key is computed once in the partition phase and reused in the send phase.
- `pkg/notifications/options.go`: `WithHookRegistrar` stays (nil-guarded) but targets the
  reworked interface.
- `pkg/notifications/hookregistrar_mock.go`: regenerated via `make generate`.
- `pkg/notifications/hooks_test.go`: rewritten for the three named hooks.
- No database, config, or external API changes. (The weather LLM-summarization hook that
  motivates AfterDedupe is a separate follow-up change that *consumes* this contract.)
