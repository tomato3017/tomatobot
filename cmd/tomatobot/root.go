/*
Copyright © 2024 Anthony Kirksey
*/
package tomatobot

import (
	_ "embed"
	"github.com/rs/zerolog"
	"os"
	"runtime/debug"
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

//go:generate sh -c "printf %s $(git describe --tags) > VERSION.txt"
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

	return nil

}

func getLogger() zerolog.Logger {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}

	log := zerolog.New(output).With().Timestamp().Logger()

	log.Info().Str("foo", "bar").Msg("Hello World")

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
