package subscribe

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
	"github.com/tomato3017/tomatobot/pkg/notifications"
	"github.com/tomato3017/tomatobot/pkg/util"
	"regexp"
)

var topicRegex = regexp.MustCompile(`(?m)^\w+[.\w]+[\w*]$`)

type SubscribeCmd struct {
	tgbot     *tgbotapi.BotAPI
	publisher notifications.Publisher
	logger    zerolog.Logger
}

func (s *SubscribeCmd) Execute(ctx context.Context, msg *tgbotapi.Message) error {
	s.logger.Debug().Msg("Subscribe command received")
	args := msg.CommandArguments()

	if args == "" {
		return fmt.Errorf("no topic provided")
	}

	if !topicRegex.MatchString(args) {
		return fmt.Errorf("invalid topic format")
	}

	sub := notifications.Subscriber{
		TopicPattern: args,
		ChatId:       msg.Chat.ID,
	}
	if err := s.publisher.Subscribe(sub); err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	_, err := s.tgbot.Send(util.NewMessageReply(msg, fmt.Sprintf("Subscribed to topic %s", args)))
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func (s *SubscribeCmd) Description() string {
	return "Subscribe to a topic"
}

func (s *SubscribeCmd) Help() string {
	return "/subscribe <topic> - Subscribe to a topic"
}
