package notifications

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"github.com/jellydator/ttlcache/v3"
	"github.com/rs/zerolog"
	dbmodels "github.com/tomato3017/tomatobot/pkg/bot/models/db"
	"github.com/tomato3017/tomatobot/pkg/util"
	"github.com/uptrace/bun"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	ErrSubExists = errors.New("subscription already exists")
)

type Publisher interface {
	Subscribe(sub Subscriber) (string, error)
	Publish(msg Message)
	Unsubscribe(topicId uuid.UUID, chatId int64) error
	GetSubscriptions(chatId int64) ([]dbmodels.Subscriptions, error)
	UnsubscribeAll(chatId int64) error
}

type Message struct {
	Topic   string
	Msg     string
	DupeKey string
	DupeTTL time.Duration
}

func (m Message) String() string {
	return fmt.Sprintf("Topic: %s, Message: %s", m.Topic, m.Msg)
}

func (m Message) DuplicationKey() string {
	return fmt.Sprintf("%s-%s", m.Topic, m.checksum())
}

func (m Message) checksum() string {
	chkMsg := m.Msg
	if m.DupeKey != "" {
		chkMsg = m.DupeKey
	}

	hasher := sha256.New()
	hasher.Write([]byte(chkMsg))
	return hex.EncodeToString(hasher.Sum(nil))
}

type Subscriber struct {
	ID           uuid.UUID
	TopicPattern string
	ChatId       int64
	//TODO priority filter?
}

func (s *Subscriber) DbModel() *dbmodels.Subscriptions {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}

	return &dbmodels.Subscriptions{
		ID:           s.ID,
		ChatID:       s.ChatId,
		TopicPattern: s.TopicPattern,
	}
}

// telegramSender abstracts the tgbotapi.BotAPI.Send method for testing.
type telegramSender interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
}

type deliveryTarget struct {
	chatId int64
	dupKey string
}

type NotificationPublisher struct {
	bus        chan Message
	wg         *sync.WaitGroup
	cancelFunc context.CancelFunc

	subscribers []Subscriber
	dbConn      bun.IDB

	tgbot    *tgbotapi.BotAPI
	tgSender telegramSender // non-nil overrides tgbot (used in tests)
	logger   zerolog.Logger

	sublck sync.RWMutex

	subCache  *ttlcache.Cache[string, []int64]
	dupeCache *ttlcache.Cache[string, struct{}]

	hooks HookRegistrar
}

var _ Publisher = &NotificationPublisher{}

func NewNotificationPublisher(tgbot *tgbotapi.BotAPI, dbConn bun.IDB, options ...PublisherOptions) *NotificationPublisher {
	publisher := NotificationPublisher{
		bus:         make(chan Message),
		subscribers: make([]Subscriber, 0),
		tgbot:       tgbot,
		logger:      zerolog.Logger{},
		dbConn:      dbConn,
		dupeCache:   ttlcache.New[string, struct{}](ttlcache.WithTTL[string, struct{}](5 * time.Minute)),
		subCache:    ttlcache.New[string, []int64](ttlcache.WithTTL[string, []int64](5 * time.Minute)),
		hooks:       newHookRegistrar(),
	}
	if err := publisher.populateDupeCache(); err != nil {
		publisher.logger.Fatal().Err(err).Msg("failed to populate dupe cache")
	}

	if err := publisher.updateSubsFromDb(); err != nil {
		publisher.logger.Fatal().Err(err).Msg("failed to update subscriptions from db")
	}

	publisher.dupeCache.OnInsertion(publisher.insertDupeCache)

	for _, option := range options {
		option(&publisher)
	}
	if publisher.tgSender == nil && publisher.tgbot != nil {
		publisher.tgSender = publisher.tgbot
	}

	return &publisher
}

func (n *NotificationPublisher) insertDupeCache(ctx context.Context, item *ttlcache.Item[string, struct{}]) {
	n.logger.Trace().Msgf("Inserting dupe cache: %s TTL: %s", item.Key(), item.ExpiresAt())
	dbCacheMdl := &dbmodels.NotificationsDupeCache{
		DupeKey:    item.Key(),
		DupeTTLEnd: item.ExpiresAt(),
	}

	_, err := n.dbConn.NewInsert().
		Model(dbCacheMdl).
		On("CONFLICT(dupe_key) DO UPDATE").
		Set("dupe_ttl_end = EXCLUDED.dupe_ttl_end").
		Where("dupe_ttl_end < EXCLUDED.dupe_ttl_end").
		Exec(ctx)
	if err != nil {
		n.logger.Error().Err(err).Msg("failed to insert dupe cache")
	}
}

func (n *NotificationPublisher) UnsubscribeAll(chatId int64) error {
	n.sublck.Lock()
	defer n.sublck.Unlock()

	//create a list of topics to remove
	topics := make([]uuid.UUID, 0)
	for _, subscriber := range n.subscribers {
		if subscriber.ChatId == chatId {
			topics = append(topics, subscriber.ID)
		}
	}

	//unsubscribe from each topic
	for _, topic := range topics {
		if err := n.unsubUnSafe(topic, chatId); err != nil {
			return fmt.Errorf("failed to unsubscribe from topic: %w", err)
		}
	}

	return nil
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

func (n *NotificationPublisher) Subscribe(sub Subscriber) (string, error) {
	n.sublck.Lock()
	defer n.sublck.Unlock()

	_, err := n.dbConn.NewInsert().Model(sub.DbModel()).Exec(context.TODO())
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return "", ErrSubExists
		}
		return "", fmt.Errorf("failed to insert subscription: %w", err)
	}

	n.subscribers = append(n.subscribers, sub)
	n.invalidateSubCache()

	return sub.ID.String(), nil
}

func (n *NotificationPublisher) invalidateSubCache() {
	n.logger.Trace().Msgf("Invalidating subscription cache")
	n.subCache.DeleteAll()
}

func (n *NotificationPublisher) unsubUnSafe(topicId uuid.UUID, chatId int64) error {
	if topicId == uuid.Nil {
		return fmt.Errorf("invalid subscription id")
	}

	changed := false
	for i, currentSub := range n.subscribers {
		if currentSub.ID == topicId && currentSub.ChatId == chatId {
			_, err := n.dbConn.NewDelete().Model(currentSub.DbModel()).
				WherePK().
				Where("chat_id = ?", chatId).
				Exec(context.TODO())

			if err != nil {
				return fmt.Errorf("failed to delete subscription: %w", err)
			}

			n.subscribers = append(n.subscribers[:i], n.subscribers[i+1:]...)
			changed = true
			break
		}
	}

	if changed {
		n.invalidateSubCache()
	}

	return nil
}

func (n *NotificationPublisher) Unsubscribe(topicId uuid.UUID, chatId int64) error {
	n.sublck.Lock()
	defer n.sublck.Unlock()

	return n.unsubUnSafe(topicId, chatId)
}

func (n *NotificationPublisher) Publish(msg Message) {
	n.bus <- msg
}

func (n *NotificationPublisher) Close() error {
	close(n.bus)
	n.subCache.Stop()
	n.dupeCache.Stop()

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

func (n *NotificationPublisher) startCaches() {
	go func() {
		n.subCache.Start()
	}()
	go func() {
		n.dupeCache.Start()
	}()
}

func (n *NotificationPublisher) Start(ctx context.Context) error {
	if n.wg != nil {
		return fmt.Errorf("already started publisher")
	}
	n.startCaches()

	n.wg = &sync.WaitGroup{}
	n.wg.Add(1)

	ctx, cf := context.WithCancel(ctx)
	n.cancelFunc = cf

	//start the message handler
	go func(ctx context.Context, wg *sync.WaitGroup) {
		n.logger.Trace().Msgf("Starting message handler")
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

	//start db cache cleanup routine
	n.wg.Add(1)
	go func(ctx context.Context, wg *sync.WaitGroup) {
		defer wg.Done()
		n.logger.Trace().Msg("Starting db cache cleanup routine")
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Minute):
				n.cleanupDb(ctx)
			}
		}
	}(ctx, n.wg)

	return nil
}

func (n *NotificationPublisher) send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	if n.tgSender == nil {
		return tgbotapi.Message{}, errors.New("telegram sender not configured")
	}
	return n.tgSender.Send(c)
}

func (n *NotificationPublisher) handleBusMessage(ctx context.Context, msg Message) error {
	logger := n.logger.
		With().Str("func", "handleBusMessage").Logger()
	logger.Trace().Msgf("Handling message for topic: %s", msg.Topic)

	chatIds, err := n.getChatIdsForTopic(msg.Topic)
	if err != nil {
		return fmt.Errorf("failed to get chat ids for topic: %w", err)
	}

	deliveryTargets := n.selectDeliveryTargets(ctx, logger, msg, chatIds)
	msg, shouldSend := n.applyAfterDedupeHooks(ctx, logger, msg, deliveryTargets)
	if !shouldSend {
		return nil
	}

	if err := n.deliverMessage(ctx, logger, msg, deliveryTargets); err != nil {
		return err
	}

	return nil
}

func (n *NotificationPublisher) selectDeliveryTargets(ctx context.Context, logger zerolog.Logger, msg Message, chatIds []int64) []deliveryTarget {
	deliveryTargets := make([]deliveryTarget, 0, len(chatIds))
	trunMsg := util.TruncateString(msg.String(), 1024, "")

	for _, chatId := range chatIds {
		if err := n.hooks.FirePreDupe(ctx, msg, chatId); err != nil {
			if errors.Is(err, ErrCancelNotification) {
				logger.Trace().Msgf("predupe hook cancelled recipient %d", chatId)
				continue
			}
			logger.Error().Err(err).Msgf("predupe hook error; skipping recipient %d", chatId)
			continue
		}
		dupKey := fmt.Sprintf("%d-%s", chatId, msg.DuplicationKey())
		logger.Trace().Msgf("Message dupe key: %s", dupKey)
		if n.dupeCache.Has(dupKey) {
			logger.Trace().Msgf("Duplicate message detected: %s", trunMsg)
			continue
		}
		logger.Trace().Msgf("Message not a duplicate: %s", trunMsg)
		deliveryTargets = append(deliveryTargets, deliveryTarget{chatId: chatId, dupKey: dupKey})
	}

	return deliveryTargets
}

func (n *NotificationPublisher) applyAfterDedupeHooks(ctx context.Context, logger zerolog.Logger, msg Message, deliveryTargets []deliveryTarget) (Message, bool) {
	if len(deliveryTargets) == 0 {
		return msg, true
	}

	enrichedMsg := msg
	if err := n.hooks.FireAfterDedupe(ctx, &enrichedMsg); err != nil {
		if errors.Is(err, ErrCancelNotification) {
			logger.Trace().Msgf("afterdedupe hook cancelled message: %s", msg.Topic)
			return msg, false
		}
		logger.Error().Err(err).Msg("afterdedupe hook error; sending unenriched")
		return msg, true
	}

	return enrichedMsg, true
}

func (n *NotificationPublisher) deliverMessage(ctx context.Context, logger zerolog.Logger, msg Message, deliveryTargets []deliveryTarget) error {
	for _, target := range deliveryTargets {
		if err := n.hooks.FireOnSend(ctx, msg, target.chatId); err != nil {
			if errors.Is(err, ErrCancelNotification) {
				logger.Trace().Msgf("onsend hook cancelled recipient %d", target.chatId)
				continue
			}
			logger.Error().Err(err).Msgf("onsend hook error; skipping recipient %d", target.chatId)
			continue
		}
		logger.Trace().Msgf("Sending message to chat: %d", target.chatId)
		_, err := n.send(tgbotapi.NewMessage(target.chatId, msg.Msg))
		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
		n.dupeCache.Set(target.dupKey, struct{}{}, msg.DupeTTL)
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

	if ok := n.subCache.Has(topic); ok {
		cacheEntry := n.subCache.Get(topic)
		if cacheEntry != nil {
			n.logger.Trace().Msgf("Cache hit for topic: %s", topic)
			return cacheEntry.Value(), nil
		}
	}
	n.logger.Trace().Msgf("Cache miss for topic: %s", topic)

	chatIdSet := make(map[int64]struct{})
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
			chatIdSet[subscriber.ChatId] = struct{}{}
		}
	}

	for id := range chatIdSet {
		chatIds = append(chatIds, id)
	}

	n.logger.Trace().Msgf("Setting cache for topic: %s TO: %+v", topic, chatIds)

	n.subCache.Set(topic, chatIds, ttlcache.DefaultTTL)
	return chatIds, nil
}

func (n *NotificationPublisher) populateDupeCache() error {
	dbCache := make([]dbmodels.NotificationsDupeCache, 0)

	err := n.dbConn.NewSelect().Model(&dbCache).
		Where("dupe_ttl_end > ?", time.Now()).
		Scan(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get dupe cache: %w", err)
	}

	for _, cache := range dbCache {
		n.dupeCache.Set(cache.DupeKey, struct{}{}, time.Until(cache.DupeTTLEnd))
	}

	return nil
}

func (n *NotificationPublisher) cleanupDb(ctx context.Context) {
	n.logger.Trace().Msg("Cleaning up db")
	_, err := n.dbConn.NewDelete().Model((*dbmodels.NotificationsDupeCache)(nil)).
		Where("dupe_ttl_end < ?", time.Now()).
		Exec(ctx)
	if err != nil {
		n.logger.Error().Err(err).Msg("failed to clean up dupe cache")
	}
}
