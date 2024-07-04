package topic

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"github.com/tomato3017/tomatobot/pkg/command"
	"github.com/tomato3017/tomatobot/pkg/notifications"
	"github.com/tomato3017/tomatobot/pkg/util"
)

type UnSubCmd struct {
	command.BaseCommand
	publisher notifications.Publisher
	tgbot     *tgbotapi.BotAPI
}

func (u *UnSubCmd) Execute(ctx context.Context, params command.CommandParams) error {
	topicId := params.Args[0]
	topicUUID, err := uuid.Parse(topicId)
	if err != nil {
		return fmt.Errorf("failed to parse topic id: %w", err)
	}

	if err := u.publisher.Unsubscribe(topicUUID, params.Message.Chat.ID); err != nil {
		return fmt.Errorf("failed to unsubscribe: %w", err)
	}

	_, err = u.tgbot.Send(util.NewMessageReply(params.Message, tgbotapi.ModeMarkdownV2, "Unsubscribed from topic"))
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func (u *UnSubCmd) Description() string {
	return "Unsubscribe from a topic"
}

func (u *UnSubCmd) Help() string {
	return "/unsubscribe <topic_id> - Unsubscribe from a topic"
}

func newUnSubCmd(publisher notifications.Publisher, tgbot *tgbotapi.BotAPI) *UnSubCmd {
	return &UnSubCmd{
		BaseCommand: command.NewBaseCommand(),
		publisher:   publisher,
		tgbot:       tgbot,
	}
}
