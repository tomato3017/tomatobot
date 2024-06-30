package notifications

import (
	"context"
	"database/sql"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tomato3017/tomatobot/pkg/bot/models/db"
	"github.com/tomato3017/tomatobot/pkg/sqlmigrate"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/uptrace/bun/extra/bundebug"
	"testing"
)

type TestNotificationSuite struct {
	suite.Suite

	dbConn *bun.DB
}

func (t *TestNotificationSuite) SetupTest() {
	sqlDb, err := sql.Open(sqliteshim.ShimName, "file::memory:?cache=shared")
	require.NoError(t.T(), err)

	bunDb := bun.NewDB(sqlDb, sqlitedialect.New())
	bunDb.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true), bundebug.WithEnabled(true)))
	t.dbConn = bunDb

	_, err = sqlmigrate.MigrateDbSchema(context.Background(), t.dbConn)
	require.NoError(t.T(), err)
}

func (t *TestNotificationSuite) TearDownTest() {
	require.NoError(t.T(), t.dbConn.Close())
}

func (t *TestNotificationSuite) Test_NotificationPublisher_Subscribe() {
	publisher := NewNotificationPublisher(nil, t.dbConn)
	require.NotNil(t.T(), publisher)

	require.Zero(t.T(), len(publisher.subscribers))
	count, err := t.dbConn.NewSelect().Model(&db.Subscriptions{}).
		Count(context.Background())
	require.NoError(t.T(), err)
	require.Zero(t.T(), count)

	err = publisher.Subscribe(Subscriber{
		TopicPattern: "test.alert",
		ChatId:       12345,
	})
	require.NoError(t.T(), err)

	require.Equal(t.T(), 1, len(publisher.subscribers))

	checkCount, err := t.dbConn.NewSelect().Model(&db.Subscriptions{}).
		Count(context.Background())
	require.NoError(t.T(), err)
	require.Equal(t.T(), 1, checkCount)
}

func (t *TestNotificationSuite) Test_NotificationPublisher_updateSubsFromDb() {
	publisher := NewNotificationPublisher(nil, t.dbConn)
	require.NotNil(t.T(), publisher)

	subscriptions := []db.Subscriptions{
		{
			ChatID:       12345,
			TopicPattern: "test.alert",
		},
		{
			ChatID:       54321,
			TopicPattern: "test.alert.*",
		},
	}

	_, err := t.dbConn.NewInsert().Model(&subscriptions).Exec(context.Background())
	require.NoError(t.T(), err)

	count, err := t.dbConn.NewSelect().Model(&db.Subscriptions{}).Count(context.Background())
	require.NoError(t.T(), err)
	require.Equal(t.T(), 2, count)

	err = publisher.updateSubsFromDb()
	require.NoError(t.T(), err)

	require.Equal(t.T(), 2, len(publisher.subscribers))

}

func (t *TestNotificationSuite) Test_NotificationPublisher_Subscribe_Conflict() {
	publisher := NewNotificationPublisher(nil, t.dbConn)
	require.NotNil(t.T(), publisher)

	subscription := Subscriber{
		TopicPattern: "test.alert",
		ChatId:       12345,
	}

	_, err := t.dbConn.NewInsert().Model(subscription.DbModel()).Exec(context.Background())
	require.NoError(t.T(), err)

	err = publisher.Subscribe(subscription)
	require.ErrorIs(t.T(), err, ErrSubExists)

}

func (t *TestNotificationSuite) Test_NotificationPublisher_Unsubscribe() {
	publisher := NewNotificationPublisher(nil, t.dbConn)
	require.NotNil(t.T(), publisher)

	checkSub := Subscriber{
		TopicPattern: "test.alert",
		ChatId:       12345,
	}

	// Subscribe first
	publisher.subscribers = append(publisher.subscribers, checkSub)

	_, err := t.dbConn.NewInsert().Model(checkSub.DbModel()).Exec(context.Background())
	require.NoError(t.T(), err)

	// Now lets unsubscribe
	err = publisher.Unsubscribe(Subscriber{
		TopicPattern: "test.alert",
		ChatId:       12345,
	})
	require.NoError(t.T(), err)

	require.Zero(t.T(), len(publisher.subscribers))

	checkCount, err := t.dbConn.NewSelect().Model(&db.Subscriptions{}).
		Count(context.Background())
	require.NoError(t.T(), err)
	require.Zero(t.T(), checkCount)
}

func Test_RunNotificationSuite(t *testing.T) {
	suite.Run(t, new(TestNotificationSuite))
}
