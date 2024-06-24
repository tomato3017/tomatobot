package config

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
)

type Config struct {
	TomatoBot
}

type TomatoBot struct {
	Token string `yaml:"token" envconfig:"TOKEN" validate:"required"`
}

func (c *Config) Validate() error {
	validate := validator.New()

	return validate.Struct(c)
}

func NewConfig(data []byte) (Config, error) {
	cfg := Config{}
	err := yaml.Unmarshal(data, &cfg)

	if err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	//Check the environment variables
	if err := envconfig.Process("TOMATOBOT", &cfg); err != nil {
		return Config{}, fmt.Errorf("failed to process env variables: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, fmt.Errorf("failed to validate config: %w", err)
	}

	return cfg, nil
}

func NewConfigFromFile(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file: %w", err)
	}
	cfg, err := NewConfig(data)
	if err != nil {
		return Config{}, fmt.Errorf("failed to create config from file: %w", err)
	}
	return cfg, nil
}
