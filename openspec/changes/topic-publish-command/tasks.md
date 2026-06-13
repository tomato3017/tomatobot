## 1. Implement the publish subcommand

- [ ] 1.1 Add `pkg/modules/topic/publish.go` with a `TopicPublishCmd` mirroring `sub.go` (embeds `command.BaseCommand`, holds `publisher`, `botProxy`, `logger`)
- [ ] 1.2 `newTopicPublishCmd(publisher, botProxy, logger)` returns the command with `command.NewBaseCommand(middleware.WithMinArgs(2))`
- [ ] 1.3 In `Execute`, read `topic := params.Args[0]`, `message := params.Args[1]`, and optional `dedupeKey := params.Args[2]` when `len(params.Args) >= 3`
- [ ] 1.4 Validate the topic with the existing `topicRegex` (reject empty/malformed), consistent with `sub`
- [ ] 1.5 Build `notifications.Message{Topic: topic, Msg: message, DupeKey: dedupeKey}` (leave `DupeTTL` zero) and call `publisher.Publish(...)`
- [ ] 1.6 Reply via `botProxy.Send` confirming the message was published to the topic (note whether a dedupe key was used); do not echo the body
- [ ] 1.7 Add `Description()` and `Help()` (e.g. `/topic publish <topic> "<message>" [dedupeKey]`)

## 2. Register the subcommand

- [ ] 2.1 In `pkg/modules/topic/command.go`, register the subcommand via `RegisterSubcommand("publish", newTopicPublishCmd(publisher, botProxy, logger))`, matching the existing registration error handling

## 3. Tests

- [ ] 3.1 Publishing with topic, quoted message, and dedupe key submits a `Message` with the expected `Topic`, `Msg`, and `DupeKey` (assert against a mocked/fake `Publisher`)
- [ ] 3.2 Publishing without a dedupe key submits a `Message` with empty `DupeKey`
- [ ] 3.3 Fewer than two arguments is rejected (via `WithMinArgs(2)`) and nothing is published
- [ ] 3.4 A malformed topic is rejected and nothing is published
- [ ] 3.5 On success a confirmation reply is sent and the message body is not echoed
- [ ] 3.6 Run `go test -race -v ./pkg/modules/topic/` and confirm green

## 4. Verification

- [ ] 4.1 `make build` succeeds
- [ ] 4.2 `make lint` passes (no new issues from this change)
- [ ] 4.3 Manual smoke (optional): `/topic sub <pattern>` then `/topic publish <topic> "<msg>" <key>` delivers to the subscribed chat
