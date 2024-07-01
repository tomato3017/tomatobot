package subscribe

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
	"github.com/tomato3017/tomatobot/pkg/command"
	"github.com/tomato3017/tomatobot/pkg/notifications"
	"github.com/tomato3017/tomatobot/pkg/util"
	"regexp"
	"strings"
)

var topicRegex = regexp.MustCompile(`(?m)^\w+[.\w]+[\w*]$`)
var _ command.TomatobotCommand = &SubscribeCmd{}

type SubscribeCmd struct {
	tgbot     *tgbotapi.BotAPI
	publisher notifications.Publisher
	logger    zerolog.Logger
}

func (s *SubscribeCmd) Execute(ctx context.Context, params command.CommandParams) error {
	if len(params.Args) == 0 {
		return fmt.Errorf("no operation provided")
	}

	operation := params.Args[0]
	switch operation {
	case string(OperationCreate):
		if len(params.Args) < 2 {
			return fmt.Errorf("no topic provided")
		}

		return s.subscribe(ctx, params.Message, params.Args[1])
	case string(OperationList):
		return s.list(ctx, params.Message)
	case string(OperationUnsub): //TODO ensure that this will filter to the chat id
		if len(params.Args) < 2 {
			return fmt.Errorf("no topic id provided(use /subscribe list to get the id)")
		}

		return s.unsubscribe(ctx, params.Message, params.Args[1])
	default:
		return fmt.Errorf("unknown operation. Operations allowed: %v", []operations{OperationCreate, OperationList, OperationUnsub})
	}
}

func (s *SubscribeCmd) subscribe(ctx context.Context, msg *tgbotapi.Message, topic string) error {
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
	if err := s.publisher.Subscribe(sub); err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	_, err := s.tgbot.Send(util.NewMessageReply(msg, fmt.Sprintf("Subscribed to topic %s", topic)))
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

func (s *SubscribeCmd) list(ctx context.Context, message *tgbotapi.Message) error {
	currentSubs, err := s.publisher.GetSubscriptions(message.Chat.ID)
	if err != nil {
		return fmt.Errorf("failed to get subscriptions: %w", err)
	}

	if len(currentSubs) == 0 {
		_, err = s.tgbot.Send(util.NewMessageReply(message, "No subscriptions found"))
		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
		return nil
	}

	outMsg := strings.Builder{}
	outMsg.WriteString("Current subscriptions:\n")
	outMsg.WriteString("ID - Topic\n")
	outMsg.WriteString("```\n")
	for _, sub := range currentSubs {
		outMsg.WriteString(fmt.Sprintf("%d - %s\n", sub.ID, sub.TopicPattern))
	}
	outMsg.WriteString("```\n")

	_, err = s.tgbot.Send(util.NewMessageReply(message, outMsg.String()))
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func (s *SubscribeCmd) unsubscribe(ctx context.Context, msg *tgbotapi.Message, topicID string) error {
	return fmt.Errorf("not implemented")
}
