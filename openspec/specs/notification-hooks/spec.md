## Purpose

Defines the notification publisher lifecycle hook contract.

## Requirements

### Requirement: Named lifecycle hook points

The notification publisher SHALL expose exactly three named lifecycle hook points —
**Predupe**, **AfterDedupe**, and **OnSend** — instead of a single generic hook. Each hook
point SHALL be independently registerable by topic pattern. A registered hook SHALL run
only for messages whose topic matches the pattern it was registered with, using the same
`*` wildcard semantics as subscription matching.

#### Scenario: Predupe hook registered for a matching topic runs

- **WHEN** a Predupe hook is registered for pattern `weather.*` and a message for topic `weather.warning` is delivered to a recipient
- **THEN** the Predupe hook is invoked for that recipient

#### Scenario: AfterDedupe hook registered for a matching topic runs

- **WHEN** an AfterDedupe hook is registered for pattern `weather.*` and a message for topic `weather.warning` has at least one fresh recipient
- **THEN** the AfterDedupe hook is invoked for that message

#### Scenario: OnSend hook registered for a matching topic runs

- **WHEN** an OnSend hook is registered for pattern `weather.warning` and a message for topic `weather.warning` is delivered to a recipient
- **THEN** the OnSend hook is invoked for that recipient

#### Scenario: Hook does not run for a non-matching topic

- **WHEN** a hook is registered for pattern `weather.*` and a message for topic `news.daily` is delivered
- **THEN** the hook is not invoked

### Requirement: Per-recipient hooks fire per recipient with chat context

The Predupe and OnSend hook points SHALL fire once per resolved recipient (chat) for a
message, not once per message. Each such hook invocation SHALL receive the message and the
recipient's `chatId`.

#### Scenario: Per-recipient hook fires once per recipient

- **WHEN** a message matches two subscribed recipients and a matching Predupe or OnSend hook is registered
- **THEN** the hook is invoked twice, once per recipient, each time with that recipient's `chatId`

### Requirement: AfterDedupe fires once per message after the dedupe check

The AfterDedupe hook point SHALL fire at most once per message, after the publisher has
determined the set of fresh (non-duplicate) recipients and before any send. It SHALL be
invoked only when at least one recipient is fresh, and SHALL NOT be invoked when every
recipient is a duplicate or when there are no recipients. The AfterDedupe hook SHALL receive
a mutable message and MAY rewrite the message body; the mutation SHALL apply to all fresh
recipients of that message.

#### Scenario: AfterDedupe runs once when recipients are fresh

- **WHEN** a message matches three fresh recipients and a matching AfterDedupe hook is registered
- **THEN** the AfterDedupe hook is invoked exactly once for that message
- **AND** the message body it produces is the body sent to all three recipients

#### Scenario: AfterDedupe is skipped when all recipients are duplicates

- **WHEN** every recipient of a message fails the duplicate check
- **THEN** the AfterDedupe hook is not invoked

#### Scenario: AfterDedupe mutation does not change the dedupe key

- **WHEN** an AfterDedupe hook rewrites the message body
- **THEN** the per-recipient duplicate key recorded for the sent recipients is the same key used for the duplicate check before enrichment

### Requirement: Predupe hook fires before the dedupe check

For each recipient, the Predupe hook SHALL be invoked immediately before the publisher
performs its duplicate-message check for that recipient. The Predupe hook SHALL NOT alter
the duplicate-detection key or the dedupe decision.

#### Scenario: Predupe runs ahead of dedupe lookup

- **WHEN** a recipient is processed for a message with a matching Predupe hook
- **THEN** the Predupe hook is invoked before the duplicate-cache lookup for that recipient
- **AND** the dedupe decision is unchanged by the hook's execution

### Requirement: OnSend hook fires immediately before send

For each fresh recipient, the OnSend hook SHALL be invoked immediately before the message is
sent to that recipient, and SHALL observe the message as produced by AfterDedupe (if any
AfterDedupe hook ran).

#### Scenario: OnSend runs just before delivery

- **WHEN** a recipient passes the duplicate check for a message with a matching OnSend hook
- **THEN** the OnSend hook is invoked before the message is sent to that recipient

#### Scenario: OnSend observes the enriched message

- **WHEN** an AfterDedupe hook rewrote the message body and a matching OnSend hook is registered
- **THEN** the OnSend hook observes the enriched body

### Requirement: Per-recipient hooks can cancel delivery to a recipient

The Predupe and OnSend hooks SHALL be able to cancel delivery to the current recipient by
returning the `ErrCancelNotification` sentinel error. On cancellation the publisher SHALL
skip that recipient and continue processing the remaining recipients. A Predupe cancellation
SHALL prevent the dedupe check, enrichment, and send for that recipient; an OnSend
cancellation SHALL prevent the send for that recipient.

#### Scenario: Predupe cancellation skips the recipient

- **WHEN** a Predupe hook returns `ErrCancelNotification` for a recipient
- **THEN** that recipient is not sent the message and is not recorded in the dupe cache
- **AND** the remaining recipients are still processed

#### Scenario: OnSend cancellation skips the send

- **WHEN** an OnSend hook returns `ErrCancelNotification` for a recipient
- **THEN** that recipient is not sent the message
- **AND** the remaining recipients are still processed

### Requirement: AfterDedupe can cancel the whole message

The AfterDedupe hook SHALL be able to cancel delivery of the message to **all** recipients
by returning the `ErrCancelNotification` sentinel error. On cancellation the publisher SHALL
send the message to no recipient and record none in the dupe cache.

#### Scenario: AfterDedupe cancellation drops the message

- **WHEN** an AfterDedupe hook returns `ErrCancelNotification` for a message with fresh recipients
- **THEN** no recipient is sent the message
- **AND** no recipient is recorded in the dupe cache

### Requirement: Unexpected per-recipient hook errors do not crash the publisher

The publisher SHALL, when a Predupe or OnSend hook returns a non-nil error that is not
`ErrCancelNotification`, log the error and skip delivery to the current recipient, then
continue with the remaining recipients. The error SHALL NOT terminate the process or abort
handling of the message.

#### Scenario: Unknown per-recipient hook error skips the recipient

- **WHEN** a Predupe or OnSend hook returns an error other than `ErrCancelNotification`
- **THEN** the error is logged
- **AND** that recipient is skipped
- **AND** the remaining recipients are still processed

### Requirement: AfterDedupe errors fail open

The publisher SHALL, when an AfterDedupe hook returns a non-nil error that is not
`ErrCancelNotification`, log the error and deliver the **original, unenriched** message to
the fresh recipients. The error SHALL NOT terminate the process, abort the message, or
suppress delivery.

#### Scenario: AfterDedupe error delivers the unenriched message

- **WHEN** an AfterDedupe hook returns an error other than `ErrCancelNotification`
- **THEN** the error is logged
- **AND** the fresh recipients are sent the original message body
- **AND** those recipients are recorded in the dupe cache
