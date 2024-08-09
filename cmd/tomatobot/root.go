/*
Copyright © 2024 Anthony Kirksey
*/
package tomatobot

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/oklog/run"
	"github.com/rs/zerolog"
	"github.com/tomato3017/tomatobot/pkg/bot"
	"github.com/tomato3017/tomatobot/pkg/config"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var COMMIT = func() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value
			}
		}
	}

	return ""
}()

//go:embed VERSION.txt
var embedVERSION string

var version = func() string {
	if embedVERSION != "" {
		return embedVERSION
	}

	return "0.0.0-DEV"
}()

var (
	cfgFile string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tomatobot",
	Short: "Tomatobot telegram bot",
	RunE: func(cmd *cobra.Command, args []string) error {
		return executeBot(args)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func executeBot(args []string) error {
	logger := getLogger()

	logger.Info().Str("commit", COMMIT).Msgf("Tomatobot %s Starting!", version)

	// Load the configuration file
	logger.Info().Msg("Loading configuration file")
	cfg, err := config.NewConfigFromFile(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration file: %w", err)
	}

	logger = deriveLoggerFromLevel(logger, cfg.TomatoBot.LogLevel)
	logger.Debug().Any("config", cfg).Msg("Loaded configuration")

	if err := createDataDir(cfg); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	tomatoBot := bot.NewTomatobot(cfg, logger)

	runGrp := run.Group{}
	ctx, ctxCf := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer ctxCf()

	runGrp.Add(func() error {
		return tomatoBot.Run(ctx)
	}, func(err error) {
		ctxCf()
	})

	if cfg.Heartbeat.Enabled {
		logger.Debug().Msg("Heartbeat enabled")
		if cfg.Heartbeat.URL == "" {
			logger.Fatal().Msg("Heartbeat URL not set")
		}
		// Run the heartbeat
		runGrp.Add(func() error {
			runHeartbeat(ctx, logger, cfg)
			return nil
		}, func(err error) {
			ctxCf()
		})
	}

	if err := runGrp.Run(); err != nil {
		logger.Error().Err(err).Msg("error running bot")
		os.Exit(1)
	}

	return nil

}

func runHeartbeat(ctx context.Context, logger zerolog.Logger, cfg config.Config) {
	ticker := time.NewTicker(cfg.Heartbeat.Interval)
	defer ticker.Stop()

	for {
		if err := heartbeat(ctx, logger, cfg.Heartbeat.URL); err != nil {
			logger.Error().Err(err).Msg("Failed to send heartbeat")
		}

		select {
		case <-ctx.Done():
			logger.Debug().Msg("Heartbeat stopped")
			return
		case <-ticker.C:
		}
	}
}

func heartbeat(ctx context.Context, logger zerolog.Logger, url string) error {
	logger.Trace().Msg("Sending heartbeat")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func createDataDir(cfg config.Config) error {
	// Check if the directory exists
	if _, err := os.Stat(cfg.DataDir); os.IsNotExist(err) {
		// Directory does not exist, create it
		err := os.MkdirAll(cfg.DataDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create data directory: %w", err)
		}
	}

	// Set the current working directory to the specified directory
	err := os.Chdir(cfg.DataDir)
	if err != nil {
		log.Fatalf("Failed to set current working directory: %s", err)
	}

	return nil
}

func deriveLoggerFromLevel(logger zerolog.Logger, level config.LogLevel) zerolog.Logger {
	switch level {
	case config.LogLevelDebug:
		return logger.Level(zerolog.DebugLevel)
	case config.LogLevelInfo:
		return logger.Level(zerolog.InfoLevel)
	case config.LogLevelWarn:
		return logger.Level(zerolog.WarnLevel)
	case config.LogLevelError:
		return logger.Level(zerolog.ErrorLevel)
	case config.LogLevelTrace:
		return logger.Level(zerolog.TraceLevel)
	default:
		return logger.Level(zerolog.InfoLevel)
	}
}

func getLogger() zerolog.Logger {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}

	log := zerolog.New(output).With().Timestamp().Logger()

	return log
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "tomatobot.yml",
		"config file")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
