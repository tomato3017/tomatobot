## Context

The notification publisher (`pkg/notifications/publisher.go`) exposes
`Publish(msg Message)` where `Message{Topic, Msg, DupeKey, DupeTTL}`. Messages are produced
only by the weather and birthday pollers today, so there is no manual way to drive the
delivery pipeline. The `/topic` command (`pkg/modules/topic/`) already holds a
`notifications.Publisher`, is admin-gated, and follows a clear subcommand pattern
(`sub`, `unsub`, `list`). Adding a `publish` subcommand is the smallest change that makes the
pipeline (fan-out → per-recipient dedup → lifecycle hooks → send) exercisable on demand.

## Goals / Non-Goals

**Goals:**
- An admin-only `/topic publish <topic> "<message>" [dedupeKey]` subcommand.
- Inject through the real `Publisher.Publish` so the full delivery path (including hooks and
  dedup) runs, not a synthetic shortcut.
- Optional dedupe key to deliberately force or break duplicate detection while testing.

**Non-Goals:**
- Changing the `Publisher`/`Message` API.
- Exposing `DupeTTL` as an argument (the cache default suffices for testing).
- Delivering the test message directly to the caller (it routes by subscription, like any
  other message).
- A general scripting/replay facility — this is a single manual inject.

## Decisions

### Live as a `/topic publish` subcommand

Reuse the existing topic-module wiring: the module already injects the `Publisher`, and the
parent `/topic` command carries `WithAdminPermission()`, which the subcommand inherits.
Mirror `sub.go`: a `newTopicPublishCmd(publisher, botProxy, logger)` returning a
`BaseCommand` with `WithMinArgs(2)`, registered in `command.go`.

**Alternative considered:** a separate debug module. Rejected — the topic module already has
the publisher and the admin gate; a new module is pure overhead.

### Argument shape: `<topic> "<message>" [dedupeKey]`

The shared argument parser (`quotedCommandsRe = "([^"]*)"|(\S+)`) already supports quoted
strings, so the free-text message is a single quoted token while `topic` and `dedupeKey`
stay bare. After subcommand dispatch strips `publish`, `params.Args` is
`[topic, message, dedupeKey?]`:

```
/topic publish weather.warning "big storm incoming" key1
   Args (post-dispatch): ["weather.warning", "big storm incoming", "key1"]
```

`WithMinArgs(2)` enforces topic + message; `Args[2]` (dedupe key) is optional. Extra args
beyond index 2 are ignored.

### Dedupe key optional; TTL defaulted

`DupeKey` is set only when `Args[2]` is present. When omitted, the publisher's existing
behavior keys dedup on the message text (`checksum(Msg)`), which is fine. `DupeTTL` is left
zero so the publisher applies its configured default cache TTL. No TTL argument is exposed.

This makes the command a precise tool for the `notification-lifecycle-hooks` work: publishing
the same dedupe key twice should show the once-per-message `AfterDedupe` hook (and `OnSend`)
firing the first time and being skipped the second.

### Confirm, don't echo

`Publish` routes by subscription; the command does not send the message back to the caller.
The reply confirms publication (topic + whether a dedupe key was used). Testers subscribe
themselves first (`/topic sub <pattern>`) to receive the message through the normal path.

### Topic validation

Reuse the module's existing `topicRegex` to reject obviously malformed topics, consistent
with `sub`. A published topic is concrete (not a pattern), but the regex permits the same
shape and keeps behavior uniform; an empty topic is rejected.

## Risks / Trade-offs

- **Publishing to all subscribers is powerful** → Admin-only (inherited) is the right gate;
  no broader exposure.
- **"Nothing happens" if not subscribed** → Expected: delivery is subscription-routed.
  Mitigated by a clear confirmation reply and documented test ritual (sub, then publish).
- **`Publish` blocks on the unbuffered bus until the handler reads it** → A momentary,
  bounded handoff; acceptable for an interactive admin command.
