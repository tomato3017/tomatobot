package weather

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	dbmodels "github.com/tomato3017/tomatobot/pkg/bot/models/db"
	"github.com/tomato3017/tomatobot/pkg/config"
	"github.com/tomato3017/tomatobot/pkg/modules/weather/owm"
	"github.com/tomato3017/tomatobot/pkg/notifications"
	"strings"
	"testing"
	"time"
)

func TestPoller_publishWeatherForLocation(t *testing.T) {
	mockPublisher := notifications.NewMockPublisher(t)
	mockClient := owm.NewMockOpenWeatherMapIClient(t)

	testPoller := newPoller(pollerNewArgs{
		publisher: mockPublisher,
		locations: make([]dbmodels.WeatherPollingLocations, 0),
		cfg:       config.WeatherConfig{},
		logger:    zerolog.Logger{},
		dbConn:    nil,
	})
	testPoller.client = mockClient

	testLocation := owm.Location{
		Latitude:  55,
		Longitude: -55,
	}
	testLocationDbMdl := dbmodels.WeatherPollingLocations{
		ZipCode: "12345",
		Lat:     testLocation.Latitude,
		Lon:     testLocation.Longitude,
	}

	testTime, err := time.Parse(time.RFC3339, "2021-11-06T00:00:00Z")
	require.NoError(t, err)
	endAlertTime := testTime.Add(time.Hour * 2)

	owmResponse := owm.OneCallCurrentResponse{
		Lat:            testLocation.Latitude,
		Lon:            testLocation.Longitude,
		Timezone:       "America/New_York",
		TimezoneOffset: -17000,
		Alerts: []owm.Alerts{
			{
				SenderName:  "NWS TEST",
				Event:       "Super High heat warning",
				Start:       testTime.Add(time.Hour * -1).Unix(),
				End:         endAlertTime.Unix(),
				Description: "It's hot",
				Tags:        []string{"heat", "warning"},
			},
		},
	}

	//expectedMsg := notifications.Message{
	//	Topic:   fmt.Sprintf("%s.%s.alerts", WeatherPollerTopic, testLocationDbMdl.ZipCode),
	//	Msg:     mock.Anything,
	//	DupeKey: "",
	//	DupeTTL: time.Hour * 6,
	//}
	mockClient.EXPECT().CurrentWeatherByLocation(mock.Anything, testLocation).Return(owmResponse, nil)
	mockPublisher.On("Publish", mock.MatchedBy(func(msg notifications.Message) bool {
		return strings.Contains(msg.Msg, "Super High heat warning")
	})).Return()

	err = testPoller.publishWeatherForLocation(context.Background(), testLocationDbMdl)

	require.NoError(t, err)
}
