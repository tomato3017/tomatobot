package weather

import (
	"context"
	"database/sql"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	dbmodels "github.com/tomato3017/tomatobot/pkg/bot/models/db"
	"github.com/tomato3017/tomatobot/pkg/modules/weather/owm"
	"github.com/tomato3017/tomatobot/pkg/notifications"
	"github.com/tomato3017/tomatobot/pkg/sqlmigrate"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/uptrace/bun/extra/bundebug"
	"testing"
)

type TestWCmdAddSuite struct {
	suite.Suite

	dbConn *bun.DB
}

func (t *TestWCmdAddSuite) SetupTest() {
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

func (t *TestWCmdAddSuite) TearDownTest() {
	require.NoError(t.T(), t.dbConn.Close())
}

func (t *TestWCmdAddSuite) Test_WCmdAdd_checkZipInDb() {
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

	t.Run("no zip in db", func() {
		w := &weatherCmdAdd{
			dbConn: t.dbConn,
		}

		err = w.dbConn.RunInTx(context.Background(), nil, func(ctx context.Context, tx bun.Tx) error {
			geoLoc, err := w.checkZipInDb(ctx, tx, "90211")
			require.NoError(t.T(), err)
			require.Equal(t.T(), dbmodels.WeatherPollingLocations{}, geoLoc)
			return nil
		})
		require.NoError(t.T(), err)
	})

	t.Run("zip in db", func() {
		w := &weatherCmdAdd{
			dbConn: t.dbConn,
		}

		err = w.dbConn.RunInTx(context.Background(), nil, func(ctx context.Context, tx bun.Tx) error {
			geoLoc, err := w.checkZipInDb(ctx, tx, "90210")
			require.NoError(t.T(), err)
			require.Equal(t.T(), dbmodels.WeatherPollingLocations{}, geoLoc)
			return nil
		})
	})
}

func (t *TestWCmdAddSuite) Test_WCmdAdd_addLocationToSubscriptions() {
	t.Run("add location to subscriptions", func() {
		mockPub := notifications.NewMockPublisher(t.T())
		mockPub.EXPECT().Subscribe(notifications.Subscriber{
			TopicPattern: "weather.90210.warning",
			ChatId:       12345,
		}).Return("", nil)

		mockPub.EXPECT().Subscribe(notifications.Subscriber{
			TopicPattern: "weather.90210.watch",
			ChatId:       12345,
		}).Return("", nil)

		mockPub.EXPECT().Subscribe(notifications.Subscriber{
			TopicPattern: "weather.90210.advisory",
			ChatId:       12345,
		}).Return("", nil)

		w := &weatherCmdAdd{
			dbConn:    t.dbConn,
			publisher: mockPub,
		}

		_, err := w.addLocationToSubscriptions(12345, "90210")
		require.NoError(t.T(), err)
	})

}

func (t *TestWCmdAddSuite) Test_WCmdAdd_addWeatherLocation() {
	mockOwm := owm.NewMockOpenWeatherMapIClient(t.T())
	mockOwm.EXPECT().GetLocationDataForZipCode("90210").Return(owm.GeolocationResponse{
		Zip:     "90210",
		Name:    "Beverly Hills",
		Lat:     1,
		Lon:     2,
		Country: "US",
	}, nil)

	mockPub := notifications.NewMockPublisher(t.T())
	mockPub.EXPECT().Subscribe(notifications.Subscriber{
		TopicPattern: "weather.90210.warning",
		ChatId:       12345,
	}).Return("", nil)

	mockPub.EXPECT().Subscribe(notifications.Subscriber{
		TopicPattern: "weather.90210.watch",
		ChatId:       12345,
	}).Return("", nil)

	mockPub.EXPECT().Subscribe(notifications.Subscriber{
		TopicPattern: "weather.90210.advisory",
		ChatId:       12345,
	}).Return("", nil)

	weatherAdd := weatherCmdAdd{
		owmClient: mockOwm,
		dbConn:    t.dbConn,
		publisher: mockPub,
	}

	_, err := weatherAdd.addWeatherLocation(context.Background(), "90210", 12345)
	require.NoError(t.T(), err)

	//Ensure that the db has the location data
	checkGeoLoc := dbmodels.WeatherPollingLocations{
		ZipCode: "90210",
	}

	err = t.dbConn.NewSelect().Model(&checkGeoLoc).Where("zip_code = ?", "90210").
		Relation("Chats").
		Scan(context.Background())
	require.NoError(t.T(), err)
	require.Equal(t.T(), "Beverly Hills", checkGeoLoc.Name)
	require.Len(t.T(), checkGeoLoc.Chats, 1)
}

func (t *TestWCmdAddSuite) Test_WCmdAdd_addWeatherLocation_alreadyadded() {
	mockOwm := owm.NewMockOpenWeatherMapIClient(t.T())

	mockPub := notifications.NewMockPublisher(t.T())
	dbLoc := dbmodels.WeatherPollingLocations{
		Name:    "Beverly Hills",
		Country: "US",
		ZipCode: "90210",
		Lon:     1,
		Lat:     2,
		Polling: true,
	}

	_, err := t.dbConn.NewInsert().Model(&dbLoc).Exec(context.Background())
	require.NoError(t.T(), err)

	dbLocChat := dbmodels.WeatherPollerChats{
		ChatID:           12345,
		PollerLocationID: dbLoc.ID,
	}
	_, err = t.dbConn.NewInsert().Model(&dbLocChat).Exec(context.Background())
	require.NoError(t.T(), err)

	weatherAdd := weatherCmdAdd{
		owmClient: mockOwm,
		dbConn:    t.dbConn,
		publisher: mockPub,
	}

	_, err = weatherAdd.addWeatherLocation(context.Background(), "90210", 12345)
	require.Error(t.T(), err)
	require.ErrorContains(t.T(), err, "already exists")

	//Ensure that the db has the location data
	checkGeoLoc := dbmodels.WeatherPollingLocations{
		ZipCode: "90210",
	}

	err = t.dbConn.NewSelect().Model(&checkGeoLoc).Where("zip_code = ?", "90210").
		Relation("Chats").
		Scan(context.Background())
	require.NoError(t.T(), err)
	require.Equal(t.T(), "Beverly Hills", checkGeoLoc.Name)
	require.Len(t.T(), checkGeoLoc.Chats, 1)
}

func (t *TestWCmdAddSuite) Test_WCmdAdd_addWeatherLocation_exists() {
	mockOwm := owm.NewMockOpenWeatherMapIClient(t.T())

	mockPub := notifications.NewMockPublisher(t.T())
	mockPub.EXPECT().Subscribe(notifications.Subscriber{
		TopicPattern: "weather.90210.warning",
		ChatId:       12345,
	}).Return("", nil)

	mockPub.EXPECT().Subscribe(notifications.Subscriber{
		TopicPattern: "weather.90210.watch",
		ChatId:       12345,
	}).Return("", nil)

	mockPub.EXPECT().Subscribe(notifications.Subscriber{
		TopicPattern: "weather.90210.advisory",
		ChatId:       12345,
	}).Return("", nil)

	dbLoc := dbmodels.WeatherPollingLocations{
		Name:    "Beverly Hills",
		Country: "US",
		ZipCode: "90210",
		Lon:     1,
		Lat:     2,
		Polling: true,
	}

	_, err := t.dbConn.NewInsert().Model(&dbLoc).Exec(context.Background())
	require.NoError(t.T(), err)

	weatherAdd := weatherCmdAdd{
		owmClient: mockOwm,
		dbConn:    t.dbConn,
		publisher: mockPub,
	}

	_, err = weatherAdd.addWeatherLocation(context.Background(), "90210", 12345)
	require.NoError(t.T(), err)

	//Ensure that the db has the location data
	checkGeoLoc := dbmodels.WeatherPollingLocations{
		ZipCode: "90210",
	}

	err = t.dbConn.NewSelect().Model(&checkGeoLoc).Where("zip_code = ?", "90210").
		Relation("Chats").
		Scan(context.Background())
	require.NoError(t.T(), err)
	require.Equal(t.T(), "Beverly Hills", checkGeoLoc.Name)
	require.Len(t.T(), checkGeoLoc.Chats, 1)
}

func TestRunWCmdAddSuite(t *testing.T) {
	suite.Run(t, new(TestWCmdAddSuite))
}
