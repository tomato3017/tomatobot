package subscribe

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tomato3017/tomatobot/pkg/command"
	"github.com/tomato3017/tomatobot/pkg/notifications"
	"github.com/tomato3017/tomatobot/pkg/util"
	"strings"
)

type SubListCmd struct {
	publisher notifications.Publisher
	tgbot     *tgbotapi.BotAPI
}

func (s *SubListCmd) Execute(ctx context.Context, params command.CommandParams) error {
	message := params.Message
	currentSubs, err := s.publisher.GetSubscriptions(message.Chat.ID)
	if err != nil {
		return fmt.Errorf("failed to get subscriptions: %w", err)
	}

	if len(currentSubs) == 0 {
		_, err = s.tgbot.Send(util.NewMessageReply(message, tgbotapi.ModeMarkdownV2, "No subscriptions found"))
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

	_, err = s.tgbot.Send(util.NewMessageReply(message, tgbotapi.ModeMarkdownV2, outMsg.String()))
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func (s *SubListCmd) Description() string {
	return "List all subscriptions for the chat channel"
}

func (s *SubListCmd) Help() string {
	return "/sublist - List all subscriptions for the chat channel"
}
