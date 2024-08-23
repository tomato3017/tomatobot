package bot

import (
	"context"
	"fmt"
	"github.com/jellydator/ttlcache/v3"
	"github.com/rs/zerolog"
	dbmodels "github.com/tomato3017/tomatobot/pkg/bot/models/db"
	"github.com/tomato3017/tomatobot/pkg/bot/models/tgapi"
	"github.com/uptrace/bun"
	"time"
)

type ChatLogger interface {
	LogChats(ctx context.Context, msg tgapi.TGBotMsg) error
	Close() error
}

type DBChatLogger struct {
	dbConn bun.IDB
	logger zerolog.Logger

	tickerCf context.CancelFunc

	userCache *ttlcache.Cache[int64, struct{}]
}

func NewDBChatLogger(dbConn bun.IDB, logger zerolog.Logger) *DBChatLogger {
	cache := ttlcache.New[int64, struct{}](ttlcache.WithTTL[int64, struct{}](24 * time.Hour))

	return &DBChatLogger{
		dbConn:    dbConn,
		logger:    logger,
		userCache: cache,
	}
}

func (c *DBChatLogger) Start(ctx context.Context) {
	ctx, cf := context.WithCancel(ctx)
	c.tickerCf = cf
	go func() {
		c.userCache.Start()
	}()

	go func() {
		c.purgeChatsTicker(ctx)
	}()
}

func (c *DBChatLogger) purgeChatsTicker(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	c.logger.Trace().Msgf("Starting chat purger")

	for {
		select {
		case <-ctx.Done():
			c.logger.Trace().Msg("Stopping chat purger")
			return
		case <-ticker.C:
			if err := c.purgeOldChats(ctx); err != nil {
				c.logger.Error().Err(err).Msg("failed to purge old chats")
			}
		}
	}
}

func (c *DBChatLogger) purgeOldChats(ctx context.Context) error {
	c.logger.Trace().Msg("Purging old chats")
	_, err := c.dbConn.NewDelete().Model(&dbmodels.ChatLogs{}).Where("created_at < ?", time.Now().Add(-24*14*time.Hour)).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to purge old chats: %w", err)
	}

	return nil
}

func (c *DBChatLogger) LogChats(ctx context.Context, msg tgapi.TGBotMsg) error {
	if err := c.logUser(ctx, msg); err != nil {
		return fmt.Errorf("failed to log user: %w", err)
	}

	for _, entry := range msg.NormalizedTextData() {
		err := c.logChat(ctx, msg, entry)
		if err != nil {
			return fmt.Errorf("failed to log chat: %w", err)
		}
	}

	return nil
}

func (c *DBChatLogger) logChat(ctx context.Context, msg tgapi.TGBotMsg, entry tgapi.SerializableTextData) error {
	var messageData string

	switch entry.Type {
	case tgapi.TextDataText:
		messageData = entry.String()
	default:
		messageData = entry.ToBase64()
	}

	chatMdl := dbmodels.ChatLogs{
		ChatID:    msg.InnerMsg().Chat.ID,
		UserID:    msg.InnerMsg().From.ID,
		MessageID: int64(msg.InnerMsg().MessageID),
		Type:      entry.Type,
		Message:   messageData,
	}

	_, err := c.dbConn.NewInsert().Model(&chatMdl).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to insert chat log: %w", err)
	}

	return nil
}

func (c *DBChatLogger) Close() error {
	c.userCache.Stop()
	c.tickerCf()
	return nil
}

func (c *DBChatLogger) logUser(ctx context.Context, msg tgapi.TGBotMsg) error {
	if ok := c.userCache.Has(msg.InnerMsg().From.ID); ok {
		return nil
	}

	userMdl := dbmodels.TelegramUser{
		ID:       msg.InnerMsg().From.ID,
		UserName: msg.InnerMsg().From.UserName,
	}

	_, err := c.dbConn.NewInsert().Model(&userMdl).On("CONFLICT (id) DO NOTHING").Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	c.userCache.Set(msg.InnerMsg().From.ID, struct{}{}, ttlcache.DefaultTTL)

	return nil
}
