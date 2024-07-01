package notifications

import (
	"context"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
	dbmodels "github.com/tomato3017/tomatobot/pkg/bot/models/db"
	"github.com/uptrace/bun"
	"regexp"
	"strings"
	"sync"
)

var (
	ErrSubExists = errors.New("subscription already exists")
)

type Publisher interface {
	Subscribe(sub Subscriber) error
	Publish(msg Message)
	Unsubscribe(sub Subscriber) error
	GetSubscriptions(chatId int64) ([]dbmodels.Subscriptions, error)
}

type Message struct {
	Topic string
	Msg   string
}

type Subscriber struct {
	TopicPattern string
	ChatId       int64
	//TODO priority filter?
}

func (s *Subscriber) DbModel() *dbmodels.Subscriptions {
	return &dbmodels.Subscriptions{
		ChatID:       s.ChatId,
		TopicPattern: s.TopicPattern,
	}
}

type NotificationPublisher struct {
	bus        chan Message
	wg         *sync.WaitGroup
	cancelFunc context.CancelFunc

	subscribers []Subscriber
	dbConn      bun.IDB

	tgbot  *tgbotapi.BotAPI
	logger zerolog.Logger

	sublck sync.RWMutex
}

var _ Publisher = &NotificationPublisher{}

func NewNotificationPublisher(tgbot *tgbotapi.BotAPI, dbConn bun.IDB, options ...PublisherOptions) *NotificationPublisher {
	publisher := NotificationPublisher{
		bus:         make(chan Message),
		subscribers: make([]Subscriber, 0),
		tgbot:       tgbot,
		logger:      zerolog.Logger{},
		dbConn:      dbConn,
	}

	for _, option := range options {
		option(&publisher)
	}

	return &publisher
}

func (n *NotificationPublisher) updateSubsFromDb() error {
	subs := make([]dbmodels.Subscriptions, 0)

	err := n.dbConn.NewSelect().Model(&subs).Scan(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get subscriptions: %w", err)
	}

	for _, sub := range subs {
		n.subscribers = append(n.subscribers, Subscriber{
			ChatId:       sub.ChatID,
			TopicPattern: sub.TopicPattern,
		})
	}

	return nil
}

func (n *NotificationPublisher) GetSubscriptions(chatId int64) ([]dbmodels.Subscriptions, error) {
	subs := make([]dbmodels.Subscriptions, 0)

	err := n.dbConn.NewSelect().Model(&subs).Where("chat_id = ?", chatId).Scan(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions: %w", err)
	}

	return subs, nil
}

func (n *NotificationPublisher) Subscribe(sub Subscriber) error {
	n.sublck.Lock()
	defer n.sublck.Unlock()

	_, err := n.dbConn.NewInsert().Model(sub.DbModel()).Exec(context.TODO())
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrSubExists
		}
		return fmt.Errorf("failed to insert subscription: %w", err)
	}

	n.subscribers = append(n.subscribers, sub)
	return nil
}

func (n *NotificationPublisher) Unsubscribe(sub Subscriber) error {
	n.sublck.Lock()
	defer n.sublck.Unlock()
	for i, currentSub := range n.subscribers {
		if currentSub.ChatId == sub.ChatId && currentSub.TopicPattern == sub.TopicPattern {
			_, err := n.dbConn.NewDelete().Model(sub.DbModel()).
				Where("chat_id = ?", sub.ChatId).
				Where("topic_pattern = ?", sub.TopicPattern).
				Exec(context.TODO())

			if err != nil {
				return fmt.Errorf("failed to delete subscription: %w", err)
			}

			n.subscribers = append(n.subscribers[:i], n.subscribers[i+1:]...)
			return nil
		}
	}
	return nil
}

func (n *NotificationPublisher) Publish(msg Message) {
	n.bus <- msg
}

func (n *NotificationPublisher) Close() error {
	close(n.bus)

	return nil
}

func (n *NotificationPublisher) Stop() error {
	if n.wg == nil {
		return fmt.Errorf("publisher not started")
	}

	n.cancelFunc()
	n.wg.Wait()

	return n.Close()
}

func (n *NotificationPublisher) Start(ctx context.Context) error {
	if n.wg != nil {
		return fmt.Errorf("already started publisher")
	}

	n.wg = &sync.WaitGroup{}
	n.wg.Add(1)

	ctx, cf := context.WithCancel(ctx)
	n.cancelFunc = cf

	go func(ctx context.Context, wg *sync.WaitGroup) {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-n.bus:
				if err := n.handleBusMessage(ctx, msg); err != nil {
					n.logger.Error().Err(err).Msg("failed to handle bus message")
					continue
				}
			}
		}
	}(ctx, n.wg)

	return nil
}

func (n *NotificationPublisher) handleBusMessage(ctx context.Context, msg Message) error {
	logger := n.logger.
		With().Str("func", "handleBusMessage").Logger()
	logger.Trace().Msgf("Handling message for topic: %s", msg.Topic)

	// get the chat ids for the topic
	chatIds, err := n.getChatIdsForTopic(msg.Topic)
	if err != nil {
		return fmt.Errorf("failed to get chat ids for topic: %w", err)
	}

	for _, chatId := range chatIds {
		// send the message to the chat
		_, err := n.tgbot.Send(tgbotapi.NewMessage(chatId, msg.Msg))
		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
	}

	return nil
}

func (n *NotificationPublisher) tokenizeTopicString(topic string) string {
	rtnStr := strings.ReplaceAll(topic, "*", "<star>")
	return rtnStr
}

func (n *NotificationPublisher) detokenizeTopicString(topic string) string {
	rtnStr := strings.ReplaceAll(topic, "<star>", ".*")
	return rtnStr
}

func (n *NotificationPublisher) getChatIdsForTopic(topic string) ([]int64, error) {
	n.sublck.RLock()
	defer n.sublck.RUnlock()

	chatIds := make([]int64, 0)
	for _, subscriber := range n.subscribers {
		pattern := subscriber.TopicPattern

		tokenizedStr := n.tokenizeTopicString(pattern)
		escapedString := regexp.QuoteMeta(tokenizedStr)
		finalPattern := n.detokenizeTopicString(escapedString)

		//compile the regex
		re, err := regexp.Compile(finalPattern)
		if err != nil {
			return nil, fmt.Errorf("failed to compile regex: %w", err)
		}

		//check if the topic matches the regex
		if re.MatchString(topic) {
			chatIds = append(chatIds, subscriber.ChatId)
		}
	}

	//TODO cache the chat ids for the topic
	return chatIds, nil
}
