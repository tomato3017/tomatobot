package topic

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
	"github.com/tomato3017/tomatobot/pkg/command"
	"github.com/tomato3017/tomatobot/pkg/notifications"
	"github.com/tomato3017/tomatobot/pkg/util"
	mfmt "github.com/tomato3017/tomatobot/pkg/util/markdownfmt"
)

type TopicSubCmd struct {
	command.BaseCommand

	tgbot     *tgbotapi.BotAPI
	publisher notifications.Publisher
	logger    zerolog.Logger
}

func (t *TopicSubCmd) Execute(ctx context.Context, params command.CommandParams) error {
	topic := params.Args[0]
	msg := params.Message

	if topic == "" {
		return fmt.Errorf("no topic provided")
	}

	if !topicRegex.MatchString(topic) {
		return fmt.Errorf("invalid topic format")
	}

	sub := notifications.Subscriber{
		TopicPattern: topic,
		ChatId:       msg.Chat.ID,
	}

	subId, err := t.publisher.Subscribe(sub)
	if err != nil {
		return fmt.Errorf("failed to topic: %w", err)
	}

	_, err = t.tgbot.Send(util.NewMessageReply(msg, tgbotapi.ModeMarkdownV2,
		mfmt.Sprintf("Subscribed to topic %m with id %m!", topic, subId)))
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func (t *TopicSubCmd) Description() string {
	return "Subscribe to a topic"
}

func (t *TopicSubCmd) Help() string {
	return "/topic subscribe <topic> - Subscribe to a topic"
}

func newTopicSubCmd(publisher notifications.Publisher, tgbot *tgbotapi.BotAPI, logger zerolog.Logger) *TopicSubCmd {
	return &TopicSubCmd{
		BaseCommand: command.NewBaseCommand(),
		publisher:   publisher,
		tgbot:       tgbot,
		logger:      logger,
	}
}
