package topic

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/tomato3017/tomatobot/pkg/bot/proxy"
	"github.com/tomato3017/tomatobot/pkg/command"
	"github.com/tomato3017/tomatobot/pkg/command/middleware"
	"github.com/tomato3017/tomatobot/pkg/command/models"
	"github.com/tomato3017/tomatobot/pkg/notifications"
	"github.com/tomato3017/tomatobot/pkg/util"
)

type UnSubCmd struct {
	command.BaseCommand
	publisher notifications.Publisher
	botProxy  proxy.TGBotImplementation
	logger    zerolog.Logger
}

var _ command.TomatobotCommand = &UnSubCmd{}

func (u *UnSubCmd) Execute(ctx context.Context, params models.CommandParams) error {
	topicId := params.Args[0]
	if topicId == "*" { // Unsubscribe from all topics
		return u.unsubscribeAllTopics(ctx, params)
	} else {
		return u.unsubscribeTopic(ctx, params)
	}
}

func (u *UnSubCmd) unsubscribeTopic(ctx context.Context, params models.CommandParams) error {
	topicId := params.Args[0]

	topicUUID, err := uuid.Parse(topicId)
	if err != nil {
		return fmt.Errorf("failed to parse topic id: %w", err)
	}

	u.logger.Debug().Msgf("Calling unsubscribe on topic %s", topicUUID)
	if err := u.publisher.Unsubscribe(topicUUID, params.Message.Chat.ID); err != nil {
		return fmt.Errorf("failed to unsubscribe: %w", err)
	}

	_, err = u.botProxy.Send(util.NewMessageReply(params.Message, tgbotapi.ModeMarkdownV2, "Unsubscribed from topic"))
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func (u *UnSubCmd) Description() string {
	return "Unsubscribe from a topic"
}

func (u *UnSubCmd) Help() string {
	return "unsubscribe <topic_id> - Unsubscribe from a topic"
}

func (u *UnSubCmd) unsubscribeAllTopics(ctx context.Context, params models.CommandParams) error {
	if err := u.publisher.UnsubscribeAll(params.Message.Chat.ID); err != nil {
		return fmt.Errorf("failed to unsubscribe from all topics: %w", err)
	}

	_, err := u.botProxy.Send(util.NewMessageReply(params.Message, tgbotapi.ModeMarkdownV2, "Unsubscribed from all topics"))
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func newUnSubCmd(publisher notifications.Publisher, botProxy proxy.TGBotImplementation, logger zerolog.Logger) *UnSubCmd {
	return &UnSubCmd{
		BaseCommand: command.NewBaseCommand(middleware.WithNArgs(1)),
		publisher:   publisher,
		botProxy:    botProxy,
	}
}
