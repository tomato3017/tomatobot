package topic

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
	"github.com/tomato3017/tomatobot/pkg/bot/proxy"
	"github.com/tomato3017/tomatobot/pkg/command"
	"github.com/tomato3017/tomatobot/pkg/command/middleware"
	"github.com/tomato3017/tomatobot/pkg/command/models"
	"github.com/tomato3017/tomatobot/pkg/notifications"
	"github.com/tomato3017/tomatobot/pkg/util"
	mfmt "github.com/tomato3017/tomatobot/pkg/util/markdownfmt"
)

type TopicPublishCmd struct {
	command.BaseCommand

	botProxy  proxy.TGBotImplementation
	publisher notifications.Publisher
	logger    zerolog.Logger
}

func (t *TopicPublishCmd) Execute(ctx context.Context, params models.CommandParams) error {
	topic := params.Args[0]
	message := params.Args[1]

	var dedupeKey string
	if len(params.Args) >= 3 {
		dedupeKey = params.Args[2]
	}

	if !topicRegex.MatchString(topic) {
		return fmt.Errorf("invalid topic format")
	}

	if message == "" {
		return fmt.Errorf("message body must not be empty")
	}

	t.publisher.Publish(notifications.Message{
		Topic:   topic,
		Msg:     message,
		DupeKey: dedupeKey,
	})

	confirmation := mfmt.Sprintf("Published to topic %m", topic)
	if dedupeKey != "" {
		confirmation += mfmt.Sprintf(" (dedupe key: %m)", dedupeKey)
	}

	_, err := t.botProxy.Send(util.NewMessageReply(params.Message.InnerMsg(), tgbotapi.ModeMarkdownV2, confirmation))
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func (t *TopicPublishCmd) Description() string {
	return "Publish a message to a topic"
}

func (t *TopicPublishCmd) Help() string {
	return `/topic publish <topic> "<message>" [dedupeKey]`
}

func newTopicPublishCmd(publisher notifications.Publisher, botProxy proxy.TGBotImplementation, logger zerolog.Logger) *TopicPublishCmd {
	bCmd := command.NewBaseCommand(middleware.WithMinArgs(2), middleware.WithMaxArgs(3))
	return &TopicPublishCmd{
		BaseCommand: bCmd,
		publisher:   publisher,
		botProxy:    botProxy,
		logger:      logger,
	}
}
