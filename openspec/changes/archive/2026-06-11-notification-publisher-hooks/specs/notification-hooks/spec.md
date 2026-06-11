## ADDED Requirements

### Requirement: Hook registration by topic pattern

The system SHALL provide a `HookRegistrar` interface that owns all registered hooks and
their registration. A hook SHALL be registered against a topic pattern, where a pattern
may contain the `*` wildcard with the same semantics used for subscription matching.

#### Scenario: Register a hook for an exact topic

- **WHEN** a hook is registered for pattern `weather.warning`
- **THEN** the registrar stores the hook and associates it with that pattern

#### Scenario: Register a hook for a wildcard topic

- **WHEN** a hook is registered for pattern `weather.*`
- **THEN** the registrar stores the hook and associates it with that pattern

### Requirement: Anchored topic matching

The registrar SHALL match a fired topic against a registered pattern using anchored
matching, so a non-wildcard pattern matches only the identical topic.

#### Scenario: Wildcard pattern matches multiple topics

- **WHEN** a hook is registered for `weather.*`
- **AND** a message with topic `weather.warning` is fired
- **AND** a message with topic `weather.watch` is fired
- **THEN** the hook is invoked for both topics

#### Scenario: Exact pattern matches only the identical topic

- **WHEN** a hook is registered for `weather.warning`
- **AND** a message with topic `weather.warning` is fired
- **THEN** the hook is invoked
- **AND** WHEN a message with topic `weather.warning.severe` is fired
- **THEN** the hook is NOT invoked

#### Scenario: Non-matching topic does not invoke the hook

- **WHEN** a hook is registered for `weather.*`
- **AND** a message with topic `traffic.alert` is fired
- **THEN** the hook is NOT invoked

### Requirement: Publisher fires hooks before fan-out

The notification publisher SHALL call the registrar to fire all matching hooks exactly
once per message, before the message is fanned out to any subscriber.

#### Scenario: Hooks fire once per message

- **WHEN** a message is handled by the publisher
- **THEN** the registrar is asked to fire matching hooks one time, before the per-chat
  delivery loop runs

#### Scenario: No matching hooks

- **WHEN** a message is handled and no registered pattern matches its topic
- **THEN** no hook is invoked
- **AND** the message is delivered normally

### Requirement: Hook cancellation via sentinel error

The system SHALL define an exported sentinel error `ErrCancelNotification`. When a
matching hook returns an error that satisfies `errors.Is(err, ErrCancelNotification)`,
the publisher SHALL drop the message before any send. Because `Publish` is asynchronous
(it enqueues the message and returns no error), cancellation occurs inside message
handling and is not surfaced to the caller of `Publish`.

#### Scenario: A hook cancels the notification

- **WHEN** a matching hook returns `ErrCancelNotification`
- **THEN** the message is dropped and delivered to no subscriber

#### Scenario: Cancellation is not visible to the publisher caller

- **WHEN** a caller invokes `Publish` for a message that a hook later cancels
- **THEN** `Publish` still returns without an error to the caller

### Requirement: Fatal log on unknown hook error

When a matching hook returns an error that is NOT `ErrCancelNotification`, the publisher
SHALL emit a fatal log. In this codebase a fatal log terminates the process; this
behavior is intentional for unexpected hook failures.

#### Scenario: A hook returns an unexpected error

- **WHEN** a matching hook returns an error that is not `ErrCancelNotification`
- **THEN** the publisher emits a fatal log for that error
