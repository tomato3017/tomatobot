## Why

The notification publisher fans messages out to subscribers, but there is no way to
intercept a message between publish and delivery. We need a way for code to react to
notifications by topic (e.g. enrich, audit, or suppress them) without modifying the
publisher's core fan-out logic.

## What Changes

- Add a `HookRegistrar` interface that owns hook storage and registration, keyed by
  topic pattern (with the same `*` wildcard semantics already used for subscriptions).
- Hooks are fired by the notification publisher once per message, before the message is
  fanned out to subscribers.
- A hook may cancel delivery by returning a sentinel error (`ErrCancelNotification`);
  the message is then dropped before any send.
- Any other (unknown) error returned by a hook causes a fatal log. **Note:** in this
  codebase `logger.Fatal()` exits the process, so an unexpected hook error crashes the
  bot from inside the message-handling goroutine. This is the requested behavior and is
  called out for confirmation.
- Topic matching for hooks is **anchored** (`^...$`) so that a non-wildcard pattern like
  `weather.warning` matches only `weather.warning`, while `weather.*` matches
  `weather.warning`, `weather.watch`, etc.

## Capabilities

### New Capabilities
- `notification-hooks`: Topic-pattern-based hooks that the notification publisher fires
  per message, able to cancel delivery via a sentinel error or crash on unknown errors.

### Modified Capabilities
<!-- None: there is no existing spec for the notification publisher to amend. -->

## Impact

- `pkg/notifications/`: new hook registrar type + interface (new file, e.g. `hooks.go`),
  a wiring point on `NotificationPublisher` (struct field + constructor option), and a
  `Fire` call inside `handleBusMessage`.
- Generated mocks: a new mock for the `HookRegistrar` interface via mockery (`make generate`).
- No database, config, or external API changes.
