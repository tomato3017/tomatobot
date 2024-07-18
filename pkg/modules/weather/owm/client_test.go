package owm

import (
	"context"
	"github.com/h2non/gock"
	"github.com/stretchr/testify/require"
	"testing"
)

const jsonWeather = `{
  "lat": 29.8131,
  "lon": -95.3098,
  "timezone": "America/Chicago",
  "timezone_offset": -18000,
  "alerts": [
    {
      "sender_name": "NWS Houston/Galveston TX",
      "event": "Heat Advisory",
      "start": 1720212720,
      "end": 1720224000,
      "description": "* WHAT...Heat index values up to 111 degrees.\n\n* WHERE...Portions of south central and southeast Texas.\n\n* WHEN...Until 7 PM CDT this evening.\n\n* IMPACTS...Hot temperatures and high humidity may cause heat\nillnesses.",
      "tags": [
        "Extreme high temperature"
      ]
    }
  ]
}`

func TestOpenWeatherMapClient_CurrentWeatherByLocation(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.openweathermap.org/data/3.0/onecall?appid=test&exclude=minutely%2Chourly%2Cdaily%2Ccurrent&lat=29.8131&lon=-95.3098").
		Get(OWMAPIONECALLPATH).Reply(200).
		JSON(jsonWeather)

	client, err := NewOpenWeatherMapClient("test")
	require.NoError(t, err)

	res, err := client.CurrentWeatherByLocation(context.TODO(), Location{
		Latitude:  29.8131,
		Longitude: -95.3098,
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(res.Alerts))
}
