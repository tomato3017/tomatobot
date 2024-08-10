package topic

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
	"github.com/tomato3017/tomatobot/pkg/bot/proxy"
	"github.com/tomato3017/tomatobot/pkg/command"
	"github.com/tomato3017/tomatobot/pkg/command/models"
	"github.com/tomato3017/tomatobot/pkg/notifications"
	"github.com/tomato3017/tomatobot/pkg/util"
	"strings"
)

type TopicListCmd struct {
	command.BaseCommand
	publisher notifications.Publisher
	botProxy  proxy.TGBotImplementation
	logger    zerolog.Logger
}

func (s *TopicListCmd) Execute(ctx context.Context, params models.CommandParams) error {
	message := params.Message
	currentSubs, err := s.publisher.GetSubscriptions(message.Chat.ID)
	if err != nil {
		return fmt.Errorf("failed to get subscriptions: %w", err)
	}

	if len(currentSubs) == 0 {
		_, err = s.botProxy.Send(util.NewMessageReply(message, tgbotapi.ModeMarkdownV2, "No subscriptions found"))
		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
		return nil
	}

	outMsg := strings.Builder{}
	outMsg.WriteString("Current subscriptions:\n")
	outMsg.WriteString("ID \\- Topic\n")
	for _, sub := range currentSubs {
		outMsg.WriteString(fmt.Sprintf("\t`%s - %s`\n", sub.ID, sub.TopicPattern))
	}

	_, err = s.botProxy.Send(util.NewMessageReply(message, tgbotapi.ModeMarkdownV2, outMsg.String()))
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func (s *TopicListCmd) Description() string {
	return "List all subscriptions for the chat channel"
}

func (s *TopicListCmd) Help() string {
	return "/topic list - List all subscriptions for the chat channel"
}

func newTopicListCmd(publisher notifications.Publisher, botProxy proxy.TGBotImplementation, logger zerolog.Logger) *TopicListCmd {
	return &TopicListCmd{
		BaseCommand: command.NewBaseCommand(),
		publisher:   publisher,
		botProxy:    botProxy,
		logger:      logger,
	}
}
