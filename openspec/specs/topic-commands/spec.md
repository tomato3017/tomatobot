# Spec: Topic Commands

## Purpose

Defines the behaviour of the `/topic` slash command and its subcommands, including how admins interact with the notification publisher through the bot's command interface.

## Requirements

### Requirement: Publish subcommand injects a message into the publisher

The `/topic` command SHALL provide a `publish` subcommand that constructs a notification
message from its arguments and submits it to the notification publisher via the existing
publish entry point. The subcommand SHALL accept `<topic> "<message>" [dedupeKey]`, where
`topic` and `dedupeKey` are single tokens and `message` may be a quoted string containing
spaces.

#### Scenario: Publishing a message submits it to the publisher

- **WHEN** an admin runs `/topic publish weather.warning "big storm" key1`
- **THEN** the publisher receives a message with topic `weather.warning`, body `big storm`, and dedupe key `key1`

#### Scenario: Message may contain spaces

- **WHEN** an admin runs `/topic publish news.daily "hello there world"`
- **THEN** the publisher receives a message whose body is `hello there world`

### Requirement: Dedupe key is optional

The `publish` subcommand SHALL treat the dedupe key as optional. When the dedupe key argument
is omitted, the subcommand SHALL submit the message without an explicit dedupe key, leaving
the publisher's default duplicate handling in effect.

#### Scenario: Omitting the dedupe key

- **WHEN** an admin runs `/topic publish weather.warning "test"` with no dedupe key
- **THEN** the publisher receives a message with topic `weather.warning` and body `test` and no explicit dedupe key

### Requirement: Publish requires topic and message and is admin-only

The `publish` subcommand SHALL require at least a topic and a message argument and SHALL
reject invocations with fewer arguments. The subcommand SHALL be restricted to admins via the
admin permission already enforced on the parent `/topic` command.

#### Scenario: Missing arguments are rejected

- **WHEN** an admin runs `/topic publish weather.warning` with no message
- **THEN** the command returns an argument error and does not publish

#### Scenario: Non-admin cannot publish

- **WHEN** a non-admin user attempts `/topic publish weather.warning "test"`
- **THEN** the command is rejected by the admin permission gate and does not publish

### Requirement: Publish confirms without echoing the message

On a successful publish the subcommand SHALL reply to the caller confirming the message was
published to the given topic. The subcommand SHALL NOT send the message body back to the
caller directly; delivery to recipients SHALL occur only through normal topic subscription
routing.

#### Scenario: Successful publish confirms to the caller

- **WHEN** an admin successfully publishes a message to a topic
- **THEN** the caller receives a confirmation that the message was published to that topic

#### Scenario: Caller not subscribed receives no delivered message

- **WHEN** an admin publishes to a topic they are not subscribed to
- **THEN** the message is not delivered to the caller's chat through the subcommand
- **AND** it is delivered only to chats subscribed to a matching topic
