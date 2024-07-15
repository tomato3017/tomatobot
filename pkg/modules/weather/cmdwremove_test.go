package weather

import (
	"context"
	"database/sql"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	dbmodels "github.com/tomato3017/tomatobot/pkg/bot/models/db"
	"github.com/tomato3017/tomatobot/pkg/sqlmigrate"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/uptrace/bun/extra/bundebug"
	"testing"
)

type TestWCmdRemoveSuite struct {
	suite.Suite
	dbConn *bun.DB
}

func (t *TestWCmdRemoveSuite) TearDownTest() {
	require.NoError(t.T(), t.dbConn.Close())
}

func (t *TestWCmdRemoveSuite) SetupTest() {
	sqlDb, err := sql.Open(sqliteshim.ShimName, "file::memory:?cache=shared")
	require.NoError(t.T(), err)

	bunDb := bun.NewDB(sqlDb, sqlitedialect.New())
	bunDb.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true), bundebug.WithEnabled(true)))
	_, err = bunDb.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t.T(), err)

	t.dbConn = bunDb

	_, err = sqlmigrate.MigrateDbSchema(context.Background(), t.dbConn)
	require.NoError(t.T(), err)
}

func (t *TestWCmdRemoveSuite) Test_WCmdRemove_removeLocationInDb_NoMoreChats() {
	//First lets run it with two entries in the poller db
	checkLocation := dbmodels.WeatherPollingLocations{
		Name:    "SomethingVille",
		Country: "US",
		ZipCode: "90210",
		Lon:     -50.55,
		Lat:     49.87,
		Polling: true,
	}
	_, err := t.dbConn.NewInsert().Model(&checkLocation).Exec(context.Background())
	require.NoError(t.T(), err)

	weatherPollerChats := []dbmodels.WeatherPollerChats{
		{
			ChatID:           12345,
			PollerLocationID: checkLocation.ID,
		},
	}

	_, err = t.dbConn.NewInsert().Model(&weatherPollerChats).Exec(context.Background())
	require.NoError(t.T(), err)

	w := &weatherCmdRemove{
		dbConn: t.dbConn,
	}

	err = w.removeLocationInDb(context.Background(), checkLocation.ZipCode, 12345)
	require.NoError(t.T(), err)

	//Check if the chat was removed
	count, err := t.dbConn.NewSelect().Model(&dbmodels.WeatherPollerChats{}).Where("chat_id = ?", 12345).Count(context.Background())
	require.NoError(t.T(), err)
	require.Zero(t.T(), count)

	//check if the other chat was not removed
	count, err = t.dbConn.NewSelect().Model(&dbmodels.WeatherPollerChats{}).Where("chat_id != ?", 12345).Count(context.Background())
	require.NoError(t.T(), err)
	require.Zero(t.T(), count)

	//check if the location was set to not poll
	err = t.dbConn.NewSelect().Model(&checkLocation).WherePK().Scan(context.Background())
	require.NoError(t.T(), err)
	require.False(t.T(), checkLocation.Polling)

}

func (t *TestWCmdRemoveSuite) Test_WCmdRemove_removeLocationInPublisher() {
	//const zipCode = "90210"
	//subUUID
	//
	//mockPublisher := notifications.NewMockPublisher(t.T())
	//for _, eventType := range weatherPublisherEventTypes {
	//	mockPublisher.EXPECT().Unsubscribe(notifications.Subscriber{
	//		TopicPattern: eventType.fullTopicPath(zipCode),
	//		ChatId:       12345,
	//	}).Return(nil)
	//}

}

func (t *TestWCmdRemoveSuite) Test_WCmdRemove_removeLocationInDb() {
	//First lets run it with two entries in the poller db
	checkLocation := dbmodels.WeatherPollingLocations{
		Name:    "SomethingVille",
		Country: "US",
		ZipCode: "90210",
		Lon:     -50.55,
		Lat:     49.87,
		Polling: true,
	}
	_, err := t.dbConn.NewInsert().Model(&checkLocation).Exec(context.Background())
	require.NoError(t.T(), err)

	weatherPollerChats := []dbmodels.WeatherPollerChats{
		{
			ChatID:           12345,
			PollerLocationID: checkLocation.ID,
		},
		{
			ChatID:           54321,
			PollerLocationID: checkLocation.ID,
		},
	}

	_, err = t.dbConn.NewInsert().Model(&weatherPollerChats).Exec(context.Background())
	require.NoError(t.T(), err)

	w := &weatherCmdRemove{
		dbConn: t.dbConn,
	}

	err = w.removeLocationInDb(context.Background(), checkLocation.ZipCode, 12345)
	require.NoError(t.T(), err)

	//Check if the chat was removed
	count, err := t.dbConn.NewSelect().Model(&dbmodels.WeatherPollerChats{}).Where("chat_id = ?", 12345).Count(context.Background())
	require.NoError(t.T(), err)
	require.Zero(t.T(), count)

	//check if the other chat was not removed
	count, err = t.dbConn.NewSelect().Model(&dbmodels.WeatherPollerChats{}).Where("chat_id != ?", 12345).Count(context.Background())
	require.NoError(t.T(), err)
	require.Equal(t.T(), 1, count)
}

func Test_RunTestwCmdRemoveSuite(t *testing.T) {
	suite.Run(t, new(TestWCmdRemoveSuite))
}
