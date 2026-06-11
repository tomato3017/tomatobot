## 1. Hook registrar

- [ ] 1.1 Create `pkg/notifications/hooks.go` with `Hook` func type, `HookRegistrar` interface (`Register`, `Fire`), and exported sentinel `ErrCancelNotification`
- [ ] 1.2 Implement a concrete `hookRegistrar` storing pattern/hook entries in registration order, guarded by an `sync.RWMutex`
- [ ] 1.3 Implement anchored topic matching: reuse the tokenize → `regexp.QuoteMeta` → detokenize transform, then wrap as `^...$`; add a comment noting the intentional divergence from `getChatIdsForTopic`
- [ ] 1.4 Implement `Fire` to invoke all matching hooks in order and return the first non-nil error (cancellation or unknown)

## 2. Publisher wiring

- [ ] 2.1 Add a `hooks HookRegistrar` field to `NotificationPublisher`; default-initialize it to an empty registrar in `NewNotificationPublisher`
- [ ] 2.2 Add a `WithHookRegistrar(r HookRegistrar)` option in `pkg/notifications/options.go`
- [ ] 2.3 In `handleBusMessage`, call `n.hooks.Fire(ctx, msg)` once before resolving chat IDs; on `ErrCancelNotification` trace-log and return nil (drop), on any other error `logger.Fatal()`

## 3. Mocks & generation

- [ ] 3.1 Run `make generate` to produce `hookregistrar_mock.go` for the new interface

## 4. Tests

- [ ] 4.1 Test wildcard matching: `weather.*` fires for `weather.warning` and `weather.watch`
- [ ] 4.2 Test exact matching: `weather.warning` fires for `weather.warning` but NOT `weather.warning.severe`
- [ ] 4.3 Test non-matching topic does not invoke the hook
- [ ] 4.4 Test `ErrCancelNotification` from a hook drops the message (no send) and `Publish` returns no error to the caller
- [ ] 4.5 Test that a matching hook fires exactly once per message, before fan-out
- [ ] 4.6 Run `go test -race -v ./pkg/notifications/` and confirm green

## 5. Verification

- [ ] 5.1 `make lint` passes
- [ ] 5.2 `make build` succeeds
