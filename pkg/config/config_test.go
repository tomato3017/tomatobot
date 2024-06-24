package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	badCfg := Config{
		TomatoBot: TomatoBot{
			Token: "",
		},
	}
	require.Error(t, badCfg.Validate())

	goodCfg := Config{
		TomatoBot: TomatoBot{
			Token: "123345",
		},
	}
	require.NoError(t, goodCfg.Validate())
}

func TestNewConfigFromFile(t *testing.T) {
	path := "tomatobot.sample.yml"
	cfg, err := NewConfigFromFile(path)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Assert the values from the loaded config file
	expectedToken := "testtoken"
	require.Equal(t, expectedToken, cfg.TomatoBot.Token)
}
