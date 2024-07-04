package topic

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
	"github.com/tomato3017/tomatobot/pkg/command"
	"github.com/tomato3017/tomatobot/pkg/command/middleware"
	"github.com/tomato3017/tomatobot/pkg/notifications"
	"regexp"
)

var topicRegex = regexp.MustCompile(`(?m)^\w+[.\w]+[\w*]$`)
var _ command.TomatobotCommand = &TopicCmd{}

type TopicCmd struct {
	command.BaseCommand
	tgbot     *tgbotapi.BotAPI
	publisher notifications.Publisher
	logger    zerolog.Logger
}

func (s *TopicCmd) Description() string {
	return "Subscribe to a topic"
}

func (s *TopicCmd) Help() string {
	return "/topic <topic> - Subscribe to a topic"
}

func newTopicCmd(publisher notifications.Publisher, tgbot *tgbotapi.BotAPI, logger zerolog.Logger) (*TopicCmd, error) {
	topicCmd := TopicCmd{
		BaseCommand: command.NewBaseCommand(middleware.WithAdminPermission()),
		tgbot:       tgbot,
		publisher:   publisher,
		logger:      logger,
	}

	err := topicCmd.RegisterSubcommand("sub", newTopicSubCmd(publisher, tgbot, logger))
	if err != nil {
		return nil, fmt.Errorf("unable to register subcommand")
	}

	err = topicCmd.RegisterSubcommand("list", newTopicListCmd(publisher, tgbot, logger))
	if err != nil {
		return nil, fmt.Errorf("unable to register subcommand %s. Err: %w", "list", err)
	}

	return &topicCmd, nil
}
