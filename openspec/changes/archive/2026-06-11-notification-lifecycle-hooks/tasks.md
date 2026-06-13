## 1. Rework hook registrar

- [x] 1.1 In `pkg/notifications/hooks.go` define `Hook = func(ctx context.Context, msg Message, chatId int64) error` (per-recipient) and `EnrichHook = func(ctx context.Context, msg *Message) error` (once-per-message)
- [x] 1.2 Replace the `HookRegistrar` interface with `RegisterPreDupe`, `RegisterAfterDedupe`, `RegisterOnSend`, `FirePreDupe`, `FireAfterDedupe`, `FireOnSend`
- [x] 1.3 Update `hookRegistrar` to hold three entry slices (`preDupe`, `afterDedupe`, `onSend`) guarded by the existing `sync.RWMutex`; keep registration order
- [x] 1.4 Keep `compileHookPattern` (anchored matching); implement each `Fire*` to copy the relevant slice under RLock, run matching hooks in order, and return the first non-nil error
- [x] 1.5 Remove the old `Register`/`Fire` methods and the single-list field

## 2. Publisher wiring (three-phase handleBusMessage)

- [x] 2.1 Remove the pre-fan-out `n.hooks.Fire(ctx, msg)` call at the top of `handleBusMessage` (`publisher.go:341`)
- [x] 2.2 Phase 1 (partition): iterate recipients, call `FirePreDupe(ctx, msg, chatId)` (cancel/error → skip recipient), compute the per-recipient `dupKey` once, run the dupe-cache check, and collect surviving `{chatId, dupKey}` into a fresh set
- [x] 2.3 Phase 2 (enrich): if the fresh set is non-empty, call `FireAfterDedupe(ctx, &msg)` once — on `ErrCancelNotification` return (drop whole message), on any other error log at Error and continue with the original message (fail open)
- [x] 2.4 Phase 3 (send): for each fresh recipient call `FireOnSend(ctx, msg, chatId)` (cancel/error → skip recipient) before `tgbot.Send`, then `dupeCache.Set` using the `dupKey` captured in Phase 1
- [x] 2.5 Confirm `WithHookRegistrar` (nil-guarded) and the default `newHookRegistrar()` initialization still compile against the reworked interface

## 3. Mocks & generation

- [x] 3.1 Run `make generate` to regenerate `hookregistrar_mock.go` for the reworked interface

## 4. Tests

- [x] 4.1 Rewrite `hooks_test.go`: Predupe hook fires for matching topic (`weather.*` → `weather.warning`) and not for non-matching topic
- [x] 4.2 OnSend hook fires for matching topic and not for non-matching topic
- [x] 4.3 AfterDedupe hook fires for matching topic and not for non-matching topic
- [x] 4.4 Per-recipient hook fires once per recipient with the correct `chatId` (message matching two recipients → two invocations)
- [x] 4.5 AfterDedupe fires exactly once for a message with multiple fresh recipients, and its mutation to `msg.Msg` is the body sent to all of them
- [x] 4.6 AfterDedupe is NOT invoked when every recipient is a duplicate (and not invoked when there are no recipients)
- [x] 4.7 The dupe key recorded in Phase 3 equals the key checked in Phase 1 even when AfterDedupe rewrote the body
- [x] 4.8 Predupe `ErrCancelNotification` skips that recipient (no dupe-cache set, no send, excluded from fresh set) and remaining recipients still processed
- [x] 4.9 OnSend `ErrCancelNotification` skips the send for that recipient and remaining recipients still processed
- [x] 4.10 AfterDedupe `ErrCancelNotification` drops the whole message (no sends, no dupe-cache writes)
- [x] 4.11 AfterDedupe non-sentinel error fails open: original body delivered to fresh recipients and they are recorded in the dupe cache
- [x] 4.12 Unknown (non-sentinel) Predupe/OnSend error logs and skips the recipient without crashing; remaining recipients still processed
- [x] 4.13 Run `go test -race -v ./pkg/notifications/` and confirm green

## 5. Verification

- [x] 5.1 `make build` succeeds
- [x] 5.2 `make lint` passes (no new issues from this change)
