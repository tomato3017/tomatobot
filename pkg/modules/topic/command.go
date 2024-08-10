package topic

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/tomato3017/tomatobot/pkg/bot/proxy"
	"github.com/tomato3017/tomatobot/pkg/command"
	"github.com/tomato3017/tomatobot/pkg/command/middleware"
	"github.com/tomato3017/tomatobot/pkg/notifications"
	"regexp"
)

var topicRegex = regexp.MustCompile(`(?m)^\w+[.\w]+[\w*]$`)
var _ command.TomatobotCommand = &TopicCmd{}

type TopicCmd struct {
	command.BaseCommand
	botProxy  proxy.TGBotImplementation
	publisher notifications.Publisher
	logger    zerolog.Logger
}

func (s *TopicCmd) Description() string {
	return "Subscribe to a topic"
}

func (s *TopicCmd) Help() string {
	return "/topic <topic> - Subscribe to a topic"
}

func newTopicCmd(publisher notifications.Publisher, botProxy proxy.TGBotImplementation, logger zerolog.Logger) (*TopicCmd, error) {
	topicCmd := TopicCmd{
		BaseCommand: command.NewBaseCommand(middleware.WithAdminPermission()),
		botProxy:    botProxy,
		publisher:   publisher,
		logger:      logger,
	}

	err := topicCmd.RegisterSubcommand("sub", newTopicSubCmd(publisher, botProxy, logger))
	if err != nil {
		return nil, fmt.Errorf("unable to register subcommand")
	}

	err = topicCmd.RegisterSubcommand("unsub", newUnSubCmd(publisher, botProxy, logger))
	if err != nil {
		return nil, fmt.Errorf("unable to register subcommand")
	}

	err = topicCmd.RegisterSubcommand("list", newTopicListCmd(publisher, botProxy, logger))
	if err != nil {
		return nil, fmt.Errorf("unable to register subcommand %s. Err: %w", "list", err)
	}

	return &topicCmd, nil
}
