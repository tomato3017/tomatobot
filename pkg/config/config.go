package config

import (
	"bytes"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"os"
	"slices"
	"time"

	"github.com/go-playground/validator/v10"
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
	LogLevel LogLevel `yaml:"loglevel" envconfig:"LOGLEVEL"`

	Debug                         bool          `yaml:"debug" envconfig:"DEBUG"`
	TelegramToken                 string        `yaml:"telegramToken" envconfig:"TELEGRAM_TOKEN" validate:"required"`
	DataDir                       string        `yaml:"data_dir" envconfig:"DATA_DIR" default:"data"`
	CommandTimeout                time.Duration `yaml:"command_timeout" envconfig:"COMMAND_TIMEOUT" default:"1m"`
	SudoTimeout                   time.Duration `yaml:"sudo_timeout" envconfig:"SUDO_TIMEOUT" default:"10m"`
	SendProxiedResponsesToChannel bool          `yaml:"send_proxied_response_to_chat"`
	BotAdminIds                   []int64       `yaml:"bot_admin_ids" envconfig:"BOT_ADMIN_IDS"`
	AllModules                    *bool         `yaml:"all_modules"`
	ModulesToLoad                 []string      `yaml:"load_modules"`
	Database                      Database      `yaml:"database"`
	Modules                       ModuleConfig  `yaml:"modules"`
	Heartbeat                     Heartbeat     `yaml:"heartbeat"`
}

type Heartbeat struct {
	Enabled  bool
	Interval time.Duration `yaml:"interval"`
	URL      string        `yaml:"url"`
}

func (t *TomatoBot) IsBotAdmin(id int64) bool {
	return slices.Contains(t.BotAdminIds, id)
}

type Database struct {
	ConnectionString string  `yaml:"connection_string" envconfig:"DATABASE_CONNECTION_STRING" validate:"required"`
	LogQueries       bool    `yaml:"log_queries" envconfig:"DATABASE_LOG_QUERIES"`
	DbType           *DBType `yaml:"type" envconfig:"DATABASE_TYPE" validate:"required"` //Intentional as we need to make sure the zero value isn't the first value
}

type ModuleConfig struct {
	Weather WeatherConfig `yaml:"weather"`
}

type WeatherConfig struct {
	APIKey          string        `yaml:"api_key" envconfig:"WEATHER_API_KEY" validate:"required"`
	PollingInterval time.Duration `yaml:"polling_interval" envconfig:"WEATHER_POLLING_INTERVAL" default:"5m"`
}

func (c *Config) Validate() error {
	validate := validator.New()

	return validate.Struct(c)
}

// Custom unmarshal function that errors on unknown fields
func unmarshalStrict(data []byte, out interface{}) error {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true) // This makes the decoder error on unknown fields
	if err := dec.Decode(out); err != nil {
		return err
	}
	return nil
}

func NewConfig(data []byte) (Config, error) {
	cfg := Config{}

	err := unmarshalStrict(data, &cfg)

	if err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	//Check the environment variables
	if err := envconfig.Process("", &cfg); err != nil {
		return Config{}, fmt.Errorf("failed to process env variables: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, fmt.Errorf("failed to validate config: %w", err)
	}

	return cfg, nil
}

func NewConfigFromFile(path string) (Config, error) {
	data, err := os.ReadFile(path)
	cwd, _ := os.Getwd()
	var _ = cwd
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file: %w", err)
	}
	cfg, err := NewConfig(data)
	if err != nil {
		return Config{}, fmt.Errorf("failed to create config from file: %w", err)
	}
	return cfg, nil
}
