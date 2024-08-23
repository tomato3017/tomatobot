package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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
