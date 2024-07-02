package subscribe

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
	"github.com/tomato3017/tomatobot/pkg/command"
	"github.com/tomato3017/tomatobot/pkg/notifications"
	"github.com/tomato3017/tomatobot/pkg/util"
	mfmt "github.com/tomato3017/tomatobot/pkg/util/markdownfmt"
	"regexp"
)

var topicRegex = regexp.MustCompile(`(?m)^\w+[.\w]+[\w*]$`)
var _ command.TomatobotCommand = &SubCreateCmd{}

type SubCreateCmd struct {
	tgbot     *tgbotapi.BotAPI
	publisher notifications.Publisher
	logger    zerolog.Logger
}

func (s *SubCreateCmd) Execute(ctx context.Context, params command.CommandParams) error {
	return s.subscribe(ctx, params.Message, params.Args[0])
}

func (s *SubCreateCmd) subscribe(ctx context.Context, msg *tgbotapi.Message, topic string) error {
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

	subId, err := s.publisher.Subscribe(sub)
	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	_, err = s.tgbot.Send(util.NewMessageReply(msg, tgbotapi.ModeMarkdownV2,
		mfmt.Sprintf("Subscribed to topic %m with id %m!", topic, subId)))
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func (s *SubCreateCmd) Description() string {
	return "Subscribe to a topic"
}

func (s *SubCreateCmd) Help() string {
	return "/subscribe <topic> - Subscribe to a topic"
}

func (s *SubCreateCmd) unsubscribe(ctx context.Context, msg *tgbotapi.Message, topicID string) error {
	return fmt.Errorf("not implemented")
}
