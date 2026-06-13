## Why

There is no in-band way to inject a message into the notification publisher. Exercising the
delivery pipeline — fan-out, per-recipient dedup, and the lifecycle hooks
(`notification-lifecycle-hooks`) — currently requires waiting on a real producer (the
weather or birthday poller) to emit something. A small admin command that publishes an
arbitrary message to a topic makes the whole path testable on demand: subscribe yourself,
publish, and watch the message route through dedup and hooks to your chat.

## What Changes

- Add a `publish` subcommand to the existing `/topic` command:
  `/topic publish <topic> "<message>" [dedupeKey]`.
- It constructs a `notifications.Message{Topic, Msg, DupeKey}` from the arguments and calls
  the existing `publisher.Publish`. No publisher API changes.
- `topic` and `dedupeKey` are bare single tokens; `message` is a quoted string (the existing
  argument parser already supports `"..."`). `dedupeKey` is optional — when omitted, the
  publisher dedupes on the message text as it already does.
- `DupeTTL` is left at its zero value, so the publisher's default dedupe-cache TTL applies.
  No TTL argument is exposed.
- The subcommand requires at least two arguments (topic + message) via `WithMinArgs(2)` and
  inherits the admin-permission gate already on the parent `/topic` command.
- On success it replies with a confirmation that the message was published to the topic. It
  does **not** echo the message back: delivery happens through the normal subscription path,
  so the caller sees the message only if subscribed to a matching topic.

## Capabilities

### New Capabilities
- `topic-commands`: An admin-only `/topic publish` subcommand that injects a message into the
  notification publisher for a given topic, with an optional dedupe key, for testing the
  delivery pipeline.

### Modified Capabilities
<!-- None: no capability spec has been archived to openspec/specs/ yet. -->

## Impact

- `pkg/modules/topic/`: new `publish.go` implementing the subcommand (mirrors the existing
  `sub.go`/`unsub.go`/`list.go` pattern), registered in `command.go` via
  `RegisterSubcommand("publish", ...)`.
- No changes to `pkg/notifications/` (uses the existing `Publisher.Publish`).
- No database, config, or external API changes.
- Independent of `notification-lifecycle-hooks`: the `Publish` contract is unchanged, so this
  command works regardless of how many hooks exist.
