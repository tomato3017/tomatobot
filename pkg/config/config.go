package config

import (
	"fmt"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
)

type Config struct {
	TomatoBot
}

type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
	LogLevelTrace LogLevel = "trace"
)

type TomatoBot struct {
	LogLevel       LogLevel      `yaml:"loglevel" envconfig:"LOGLEVEL"`
	Debug          bool          `yaml:"debug" envconfig:"DEBUG"`
	Token          string        `yaml:"token" envconfig:"TOKEN" validate:"required"`
	CommandTimeout time.Duration `yaml:"command_timeout" envconfig:"COMMAND_TIMEOUT" default:"1m"`
	AllModules     *bool         `yaml:"all_modules"`
	ModulesToLoad  []string      `yaml:"modules"`
	Database       Database      `yaml:"database"`
}

type Database struct {
	ConnectionString string  `yaml:"connection_string" envconfig:"DATABASE_CONNECTION_STRING" validate:"required"`
	LogQueries       bool    `yaml:"log_queries" envconfig:"DATABASE_LOG_QUERIES"`
	DbType           *DBType `yaml:"type" envconfig:"DATABASE_TYPE" validate:"required"` //Intentional as we need to make sure the zero value isn't the first value
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
