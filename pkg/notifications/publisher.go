package notifications

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
	"regexp"
	"strings"
	"sync"
)

type Publisher interface {
	Subscribe(sub Subscriber)
	Publish(msg Message)
	Unsubscribe(sub Subscriber)
}

type Message struct {
	Topic string
	Msg   string
}

type Subscriber struct {
	TopicPattern string
	ChatId       int64
}

type NotificationPublisher struct {
	bus        chan Message
	wg         *sync.WaitGroup
	cancelFunc context.CancelFunc

	subscribers []Subscriber

	tgbot  *tgbotapi.BotAPI
	logger zerolog.Logger

	sublck sync.RWMutex
}

func NewNotificationPublisher(tgbot *tgbotapi.BotAPI, options ...PublisherOptions) *NotificationPublisher {
	publisher := NotificationPublisher{
		bus:         make(chan Message),
		subscribers: make([]Subscriber, 0),
		tgbot:       tgbot,
		logger:      zerolog.Logger{},
	}

	for _, option := range options {
		option(&publisher)
	}

	return &publisher
}

func (n *NotificationPublisher) Subscribe(sub Subscriber) {
	n.sublck.Lock()
	defer n.sublck.Unlock()
	n.subscribers = append(n.subscribers, sub)
}

func (n *NotificationPublisher) Unsubscribe(sub Subscriber) {
	n.sublck.Lock()
	defer n.sublck.Unlock()
	for i, s := range n.subscribers {
		if s.ChatId == sub.ChatId && s.TopicPattern == sub.TopicPattern {
			n.subscribers = append(n.subscribers[:i], n.subscribers[i+1:]...)
			return
		}
	}
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
