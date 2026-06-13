package notifications

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tomato3017/tomatobot/pkg/sqlmigrate"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/uptrace/bun/extra/bundebug"
)

// mockTgSender records every sent message so tests can assert on it.
type mockTgSender struct {
	mu   sync.Mutex
	sent []tgbotapi.MessageConfig
}

func (m *mockTgSender) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	if mc, ok := c.(tgbotapi.MessageConfig); ok {
		m.mu.Lock()
		m.sent = append(m.sent, mc)
		m.mu.Unlock()
	}
	return tgbotapi.Message{}, nil
}

// withTgSender injects a telegramSender into the publisher (test helper only).
func withTgSender(s telegramSender) PublisherOptions {
	return func(p *NotificationPublisher) {
		p.tgSender = s
	}
}

type TestHookSuite struct {
	suite.Suite
	dbConn *bun.DB
}

func (s *TestHookSuite) SetupTest() {
	sqlDb, err := sql.Open(sqliteshim.ShimName, "file::memory:?cache=shared")
	require.NoError(s.T(), err)

	bunDb := bun.NewDB(sqlDb, sqlitedialect.New())
	bunDb.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true), bundebug.WithEnabled(false)))
	s.dbConn = bunDb

	_, err = sqlmigrate.MigrateDbSchema(context.Background(), s.dbConn)
	require.NoError(s.T(), err)
}

func (s *TestHookSuite) TearDownTest() {
	require.NoError(s.T(), s.dbConn.Close())
}

// newPublisher creates a test publisher with the given sender and hook registrar.
func (s *TestHookSuite) newPublisher(tg telegramSender, hooks HookRegistrar, subs ...Subscriber) *NotificationPublisher {
	opts := []PublisherOptions{WithHookRegistrar(hooks)}
	if tg != nil {
		opts = append(opts, withTgSender(tg))
	}
	p := NewNotificationPublisher(nil, s.dbConn, opts...)
	p.subscribers = append(p.subscribers, subs...)
	return p
}

// 4.1: Predupe hook fires for matching topic, not for non-matching
func (s *TestHookSuite) Test_PreDupe_MatchingTopicFires() {
	reg := newHookRegistrar()
	var firedTopics []string
	reg.RegisterPreDupe("weather.*", func(_ context.Context, msg Message, _ int64) error {
		firedTopics = append(firedTopics, msg.Topic)
		return nil
	})

	require.NoError(s.T(), reg.FirePreDupe(context.Background(), Message{Topic: "weather.warning"}, 1))
	require.NoError(s.T(), reg.FirePreDupe(context.Background(), Message{Topic: "news.daily"}, 1))
	require.Equal(s.T(), []string{"weather.warning"}, firedTopics)
}

// 4.2: OnSend hook fires for matching topic, not for non-matching
func (s *TestHookSuite) Test_OnSend_MatchingTopicFires() {
	reg := newHookRegistrar()
	var firedTopics []string
	reg.RegisterOnSend("weather.*", func(_ context.Context, msg Message, _ int64) error {
		firedTopics = append(firedTopics, msg.Topic)
		return nil
	})

	require.NoError(s.T(), reg.FireOnSend(context.Background(), Message{Topic: "weather.warning"}, 1))
	require.NoError(s.T(), reg.FireOnSend(context.Background(), Message{Topic: "news.daily"}, 1))
	require.Equal(s.T(), []string{"weather.warning"}, firedTopics)
}

// 4.3: AfterDedupe hook fires for matching topic, not for non-matching
func (s *TestHookSuite) Test_AfterDedupe_MatchingTopicFires() {
	reg := newHookRegistrar()
	var firedTopics []string
	reg.RegisterAfterDedupe("weather.*", func(_ context.Context, msg *Message) error {
		firedTopics = append(firedTopics, msg.Topic)
		return nil
	})

	weather := Message{Topic: "weather.warning"}
	news := Message{Topic: "news.daily"}
	require.NoError(s.T(), reg.FireAfterDedupe(context.Background(), &weather))
	require.NoError(s.T(), reg.FireAfterDedupe(context.Background(), &news))
	require.Equal(s.T(), []string{"weather.warning"}, firedTopics)
}

// 4.4: Per-recipient hook fires once per recipient with the correct chatId
func (s *TestHookSuite) Test_PerRecipient_FiresOncePerChatId() {
	tg := &mockTgSender{}
	reg := newHookRegistrar()

	type invocation struct{ chatId int64 }
	var invocations []invocation
	reg.RegisterPreDupe("weather.*", func(_ context.Context, _ Message, chatId int64) error {
		invocations = append(invocations, invocation{chatId})
		return nil
	})
	// Cancel in OnSend to avoid actual sends for this test.
	reg.RegisterOnSend("weather.*", func(_ context.Context, _ Message, _ int64) error {
		return ErrCancelNotification
	})

	p := s.newPublisher(tg, reg,
		Subscriber{TopicPattern: "weather.*", ChatId: 111},
		Subscriber{TopicPattern: "weather.*", ChatId: 222},
	)

	err := p.handleBusMessage(context.Background(), Message{Topic: "weather.warning", Msg: "test"})
	require.NoError(s.T(), err)

	require.Len(s.T(), invocations, 2)
	chatIds := []int64{invocations[0].chatId, invocations[1].chatId}
	require.ElementsMatch(s.T(), []int64{111, 222}, chatIds)
}

// 4.5: AfterDedupe fires exactly once and its mutation is the body sent to all recipients
func (s *TestHookSuite) Test_AfterDedupe_FiresOnce_MutationAppliedToAll() {
	tg := &mockTgSender{}
	reg := newHookRegistrar()

	afterDedupeCount := 0
	reg.RegisterAfterDedupe("weather.*", func(_ context.Context, msg *Message) error {
		afterDedupeCount++
		msg.Msg = "enriched"
		return nil
	})

	p := s.newPublisher(tg, reg,
		Subscriber{TopicPattern: "weather.*", ChatId: 111},
		Subscriber{TopicPattern: "weather.*", ChatId: 222},
		Subscriber{TopicPattern: "weather.*", ChatId: 333},
	)

	err := p.handleBusMessage(context.Background(), Message{
		Topic:   "weather.warning",
		Msg:     "original",
		DupeTTL: time.Minute,
	})
	require.NoError(s.T(), err)

	require.Equal(s.T(), 1, afterDedupeCount)
	require.Len(s.T(), tg.sent, 3)
	for _, m := range tg.sent {
		require.Equal(s.T(), "enriched", m.Text)
	}
}

// 4.6: AfterDedupe is NOT invoked when every recipient is a duplicate or when there are no recipients
func (s *TestHookSuite) Test_AfterDedupe_SkippedForDuplicatesAndNoRecipients() {
	tg := &mockTgSender{}
	reg := newHookRegistrar()

	afterDedupeCount := 0
	reg.RegisterAfterDedupe("weather.*", func(_ context.Context, _ *Message) error {
		afterDedupeCount++
		return nil
	})

	p := s.newPublisher(tg, reg,
		Subscriber{TopicPattern: "weather.*", ChatId: 111},
	)

	msg := Message{Topic: "weather.warning", Msg: "test", DupeTTL: time.Minute}

	// First send — fresh recipient, AfterDedupe should fire.
	require.NoError(s.T(), p.handleBusMessage(context.Background(), msg))
	require.Equal(s.T(), 1, afterDedupeCount)

	// Second send — duplicate, AfterDedupe must NOT fire.
	require.NoError(s.T(), p.handleBusMessage(context.Background(), msg))
	require.Equal(s.T(), 1, afterDedupeCount)

	// No-recipient case — AfterDedupe must NOT fire.
	p2 := s.newPublisher(tg, reg)
	require.NoError(s.T(), p2.handleBusMessage(context.Background(), msg))
	require.Equal(s.T(), 1, afterDedupeCount)
}

// 4.7: The dupe key recorded in Phase 3 equals the key checked in Phase 1 even when AfterDedupe rewrote the body
func (s *TestHookSuite) Test_DupeKey_StableAcrossEnrichment() {
	tg := &mockTgSender{}
	reg := newHookRegistrar()

	const chatId = int64(111)
	reg.RegisterAfterDedupe("weather.*", func(_ context.Context, msg *Message) error {
		msg.Msg = "enriched body"
		return nil
	})

	p := s.newPublisher(tg, reg, Subscriber{TopicPattern: "weather.*", ChatId: chatId})

	original := Message{Topic: "weather.warning", Msg: "original body", DupeTTL: time.Minute}
	originalDupKey := fmt.Sprintf("%d-%s", chatId, original.DuplicationKey())

	enrichedRef := Message{Topic: "weather.warning", Msg: "enriched body"}
	enrichedDupKey := fmt.Sprintf("%d-%s", chatId, enrichedRef.DuplicationKey())
	require.NotEqual(s.T(), originalDupKey, enrichedDupKey, "sanity: keys must differ")

	require.NoError(s.T(), p.handleBusMessage(context.Background(), original))

	// Phase 3 must use the Phase 1 key (based on original body).
	require.True(s.T(), p.dupeCache.Has(originalDupKey))
	require.False(s.T(), p.dupeCache.Has(enrichedDupKey))
}

// 4.8: Predupe ErrCancelNotification skips that recipient; remaining recipients are still processed
func (s *TestHookSuite) Test_PreDupe_Cancel_SkipsRecipient_OthersProcessed() {
	tg := &mockTgSender{}
	reg := newHookRegistrar()

	reg.RegisterPreDupe("weather.*", func(_ context.Context, _ Message, chatId int64) error {
		if chatId == 111 {
			return ErrCancelNotification
		}
		return nil
	})

	p := s.newPublisher(tg, reg,
		Subscriber{TopicPattern: "weather.*", ChatId: 111},
		Subscriber{TopicPattern: "weather.*", ChatId: 222},
	)

	msg := Message{Topic: "weather.warning", Msg: "test", DupeTTL: time.Minute}
	require.NoError(s.T(), p.handleBusMessage(context.Background(), msg))

	// Only chatId 222 receives the message.
	require.Len(s.T(), tg.sent, 1)
	require.Equal(s.T(), int64(222), tg.sent[0].ChatID)

	// chatId 111 must NOT be recorded in the dupe cache.
	key111 := fmt.Sprintf("%d-%s", int64(111), msg.DuplicationKey())
	key222 := fmt.Sprintf("%d-%s", int64(222), msg.DuplicationKey())
	require.False(s.T(), p.dupeCache.Has(key111))
	require.True(s.T(), p.dupeCache.Has(key222))
}

// 4.9: OnSend ErrCancelNotification skips the send; remaining recipients still processed
func (s *TestHookSuite) Test_OnSend_Cancel_SkipsSend_OthersProcessed() {
	tg := &mockTgSender{}
	reg := newHookRegistrar()

	reg.RegisterOnSend("weather.*", func(_ context.Context, _ Message, chatId int64) error {
		if chatId == 111 {
			return ErrCancelNotification
		}
		return nil
	})

	p := s.newPublisher(tg, reg,
		Subscriber{TopicPattern: "weather.*", ChatId: 111},
		Subscriber{TopicPattern: "weather.*", ChatId: 222},
	)

	require.NoError(s.T(), p.handleBusMessage(context.Background(), Message{
		Topic: "weather.warning",
		Msg:   "test",
	}))

	require.Len(s.T(), tg.sent, 1)
	require.Equal(s.T(), int64(222), tg.sent[0].ChatID)
}

// 4.10: AfterDedupe ErrCancelNotification drops the whole message (no sends, no dupe-cache writes)
func (s *TestHookSuite) Test_AfterDedupe_Cancel_DropsMessage() {
	tg := &mockTgSender{}
	reg := newHookRegistrar()

	reg.RegisterAfterDedupe("weather.*", func(_ context.Context, _ *Message) error {
		return ErrCancelNotification
	})

	p := s.newPublisher(tg, reg,
		Subscriber{TopicPattern: "weather.*", ChatId: 111},
		Subscriber{TopicPattern: "weather.*", ChatId: 222},
	)

	msg := Message{Topic: "weather.warning", Msg: "test", DupeTTL: time.Minute}
	require.NoError(s.T(), p.handleBusMessage(context.Background(), msg))

	require.Len(s.T(), tg.sent, 0)
	key111 := fmt.Sprintf("%d-%s", int64(111), msg.DuplicationKey())
	key222 := fmt.Sprintf("%d-%s", int64(222), msg.DuplicationKey())
	require.False(s.T(), p.dupeCache.Has(key111))
	require.False(s.T(), p.dupeCache.Has(key222))
}

// 4.11: AfterDedupe non-sentinel error fails open: original body delivered, recipients recorded in dupe cache
func (s *TestHookSuite) Test_AfterDedupe_Error_FailsOpen() {
	tg := &mockTgSender{}
	reg := newHookRegistrar()

	hookErr := errors.New("enrichment failed")
	reg.RegisterAfterDedupe("weather.*", func(_ context.Context, msg *Message) error {
		msg.Msg = "enriched" // mutation before returning error
		return hookErr
	})

	p := s.newPublisher(tg, reg,
		Subscriber{TopicPattern: "weather.*", ChatId: 111},
	)

	msg := Message{Topic: "weather.warning", Msg: "original", DupeTTL: time.Minute}
	require.NoError(s.T(), p.handleBusMessage(context.Background(), msg))

	// Must still send (fail open) with the ORIGINAL body.
	require.Len(s.T(), tg.sent, 1)
	require.Equal(s.T(), "original", tg.sent[0].Text)

	// And the recipient must be recorded in the dupe cache.
	key := fmt.Sprintf("%d-%s", int64(111), msg.DuplicationKey())
	require.True(s.T(), p.dupeCache.Has(key))
}

// 4.12: Unknown (non-sentinel) Predupe/OnSend error logs and skips the recipient; remaining recipients still processed
func (s *TestHookSuite) Test_PerRecipientHook_UnknownError_SkipsRecipient_OthersProcessed() {
	tg := &mockTgSender{}
	reg := newHookRegistrar()

	hookErr := errors.New("unexpected hook error")
	reg.RegisterPreDupe("weather.*", func(_ context.Context, _ Message, chatId int64) error {
		if chatId == 111 {
			return hookErr
		}
		return nil
	})

	p := s.newPublisher(tg, reg,
		Subscriber{TopicPattern: "weather.*", ChatId: 111},
		Subscriber{TopicPattern: "weather.*", ChatId: 222},
	)

	require.NoError(s.T(), p.handleBusMessage(context.Background(), Message{
		Topic: "weather.warning",
		Msg:   "test",
	}))

	// chatId 111 was skipped; chatId 222 still processed.
	require.Len(s.T(), tg.sent, 1)
	require.Equal(s.T(), int64(222), tg.sent[0].ChatID)
}

func Test_RunHookSuite(t *testing.T) {
	suite.Run(t, new(TestHookSuite))
}
