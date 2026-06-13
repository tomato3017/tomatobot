## 1. Hook registrar

- [x] 1.1 Create `pkg/notifications/hooks.go` with `Hook` func type, `HookRegistrar` interface (`Register`, `Fire`), and exported sentinel `ErrCancelNotification`
- [x] 1.2 Implement a concrete `hookRegistrar` storing pattern/hook entries in registration order, guarded by an `sync.RWMutex`
- [x] 1.3 Implement anchored topic matching: reuse the tokenize → `regexp.QuoteMeta` → detokenize transform, then wrap as `^...$`; add a comment noting the intentional divergence from `getChatIdsForTopic`
- [x] 1.4 Implement `Fire` to invoke all matching hooks in order and return the first non-nil error (cancellation or unknown)

## 2. Publisher wiring

- [x] 2.1 Add a `hooks HookRegistrar` field to `NotificationPublisher`; default-initialize it to an empty registrar in `NewNotificationPublisher`
- [x] 2.2 Add a `WithHookRegistrar(r HookRegistrar)` option in `pkg/notifications/options.go`
- [x] 2.3 In `handleBusMessage`, call `n.hooks.Fire(ctx, msg)` once before resolving chat IDs; on `ErrCancelNotification` trace-log and return nil (drop), on any other error `logger.Fatal()`

## 3. Mocks & generation

- [x] 3.1 Run `make generate` to produce `hookregistrar_mock.go` for the new interface

## 4. Tests

- [x] 4.1 Test wildcard matching: `weather.*` fires for `weather.warning` and `weather.watch`
- [x] 4.2 Test exact matching: `weather.warning` fires for `weather.warning` but NOT `weather.warning.severe`
- [x] 4.3 Test non-matching topic does not invoke the hook
- [x] 4.4 Test `ErrCancelNotification` from a hook drops the message (no send) and `Publish` returns no error to the caller
- [x] 4.5 Test that a matching hook fires exactly once per message, before fan-out
- [x] 4.6 Run `go test -race -v ./pkg/notifications/` and confirm green

## 5. Verification

- [ ] 5.1 `make lint` passes — pre-existing failures in unrelated files (birthdaymodule.go, tomatobot.go, add.go); no new issues from this change
- [x] 5.2 `make build` succeeds
