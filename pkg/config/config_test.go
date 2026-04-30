package config

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestConfig_Validate(t *testing.T) {
	badCfg := Config{
		TomatoBot: TomatoBot{
			TelegramToken: "",
		},
	}
	require.Error(t, badCfg.Validate())

	dbType := DBTypeSQLite
	goodCfg := Config{
		TomatoBot: TomatoBot{
			LogLevel:      "DEBUG",
			Debug:         true,
			TelegramToken: "123345",
			Database: Database{
				ConnectionString: "sqlite://:memory:",
				DbType:           &dbType,
			},
			Modules: ModuleConfig{Weather: WeatherConfig{
				APIKey:          "12345",
				PollingInterval: time.Second,
			}},
		},
	}
	require.NoError(t, goodCfg.Validate())
}

func TestNewConfig_ExpandsEnvironmentVariables(t *testing.T) {
	t.Setenv("TELEGRAM_TOKEN", "telegram-token")
	t.Setenv("WEATHER_API_KEY", "weather-api-key")
	t.Setenv("POSTGRES_USER", "tomatobot")
	t.Setenv("POSTGRES_PASSWORD", "secret")
	t.Setenv("POSTGRES_DB", "tomatobot")

	cfg, err := NewConfig([]byte(`tomatobot:
  loglevel: "trace"
  debug: false
  database:
    type: "postgres"
    log_queries: false
    connection_string: "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=disable"
  modules:
    weather:
      polling_interval: 60s
`))

	require.NoError(t, err)
	require.Equal(t, "postgres://tomatobot:secret@postgres:5432/tomatobot?sslmode=disable", cfg.Database.ConnectionString)
}

func TestNewConfig_KeepsMissingEnvironmentVariables(t *testing.T) {
	t.Setenv("TELEGRAM_TOKEN", "telegram-token")
	t.Setenv("WEATHER_API_KEY", "weather-api-key")

	cfg, err := NewConfig([]byte(`tomatobot:
  loglevel: "trace"
  debug: false
  database:
    type: "postgres"
    log_queries: false
    connection_string: "postgres://${POSTGRES_USER}@postgres:5432/tomatobot?sslmode=disable"
  modules:
    weather:
      polling_interval: 60s
`))

	require.NoError(t, err)
	require.Equal(t, "postgres://${POSTGRES_USER}@postgres:5432/tomatobot?sslmode=disable", cfg.Database.ConnectionString)
}
