package notifications

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tomato3017/tomatobot/pkg/sqlmigrate"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/uptrace/bun/extra/bundebug"
)

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

// 4.1: wildcard pattern fires for multiple matching topics
func (s *TestHookSuite) Test_HookRegistrar_WildcardMatching() {
	reg := newHookRegistrar()
	var fired []string
	reg.Register("weather.*", func(_ context.Context, msg Message) error {
		fired = append(fired, msg.Topic)
		return nil
	})

	require.NoError(s.T(), reg.Fire(context.Background(), Message{Topic: "weather.warning"}))
	require.NoError(s.T(), reg.Fire(context.Background(), Message{Topic: "weather.watch"}))
	require.Equal(s.T(), []string{"weather.warning", "weather.watch"}, fired)
}

// 4.2: exact pattern fires only for the identical topic, not a longer one
func (s *TestHookSuite) Test_HookRegistrar_ExactMatching() {
	reg := newHookRegistrar()
	var fired []string
	reg.Register("weather.warning", func(_ context.Context, msg Message) error {
		fired = append(fired, msg.Topic)
		return nil
	})

	require.NoError(s.T(), reg.Fire(context.Background(), Message{Topic: "weather.warning"}))
	require.NoError(s.T(), reg.Fire(context.Background(), Message{Topic: "weather.warning.severe"}))
	require.Equal(s.T(), []string{"weather.warning"}, fired)
}

// 4.3: non-matching topic does not invoke the hook
func (s *TestHookSuite) Test_HookRegistrar_NonMatchingTopic() {
	reg := newHookRegistrar()
	fired := false
	reg.Register("weather.*", func(_ context.Context, _ Message) error {
		fired = true
		return nil
	})

	require.NoError(s.T(), reg.Fire(context.Background(), Message{Topic: "traffic.alert"}))
	require.False(s.T(), fired)
}

// 4.4: ErrCancelNotification from a hook drops the message; handleBusMessage returns nil
func (s *TestHookSuite) Test_Publisher_Hook_CancelsNotification() {
	hookReg := newHookRegistrar()
	hookReg.Register("weather.*", func(_ context.Context, _ Message) error {
		return ErrCancelNotification
	})

	publisher := NewNotificationPublisher(nil, s.dbConn, WithHookRegistrar(hookReg))
	// subscriber matches; tgbot is nil — if Send were called on nil it would panic,
	// proving the cancel path skips delivery entirely
	publisher.subscribers = append(publisher.subscribers, Subscriber{
		TopicPattern: "weather.*",
		ChatId:       12345,
	})

	err := publisher.handleBusMessage(context.Background(), Message{
		Topic: "weather.warning",
		Msg:   "test",
	})
	require.NoError(s.T(), err)
}

// 4.5: hook fires exactly once per message, before fan-out
func (s *TestHookSuite) Test_Publisher_Hook_FiresOnceBeforeFanout() {
	fireCount := 0
	hookReg := newHookRegistrar()
	hookReg.Register("weather.*", func(_ context.Context, _ Message) error {
		fireCount++
		// cancel to prevent Send on nil tgbot; subscriber below would otherwise trigger it
		return ErrCancelNotification
	})

	publisher := NewNotificationPublisher(nil, s.dbConn, WithHookRegistrar(hookReg))
	// subscriber matches so fan-out would be attempted if hook ran after delivery
	publisher.subscribers = append(publisher.subscribers, Subscriber{
		TopicPattern: "weather.*",
		ChatId:       12345,
	})

	require.NoError(s.T(), publisher.handleBusMessage(context.Background(), Message{
		Topic: "weather.warning",
		Msg:   "test",
	}))
	require.Equal(s.T(), 1, fireCount)
}

func Test_RunHookSuite(t *testing.T) {
	suite.Run(t, new(TestHookSuite))
}
