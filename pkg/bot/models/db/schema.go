package db

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/tomato3017/tomatobot/pkg/bot/models/tgapi"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
	"time"
)

type Subscriptions struct {
	bun.BaseModel `bun:"subscriptions"`

	ID           uuid.UUID `bun:"id,pk"`
	ChatID       int64     `bun:"chat_id,notnull,unique:subscriptions_chat_id_topic_pattern_key"`
	TopicPattern string    `bun:"topic_pattern,notnull,unique:subscriptions_chat_id_topic_pattern_key"`
}

type WeatherPollingLocations struct {
	bun.BaseModel `bun:"weather_polling_locations"`

	ID      int                   `bun:"id,pk,autoincrement"`
	Name    string                `bun:"name,notnull"`
	Country string                `bun:"country,notnull"`
	ZipCode string                `bun:"zip_code,notnull,unique:weather_polling_locations_zip_code_key"`
	Lon     float64               `bun:"lon,notnull"`
	Lat     float64               `bun:"lat,notnull"`
	Polling bool                  `bun:"polling,notnull,default:true"`
	Chats   []*WeatherPollerChats `bun:"rel:has-many,join:id=poller_location_id"`
}

func (w WeatherPollingLocations) IsEmpty() bool {
	return w.ID == 0 && w.Name == "" && w.Country == "" && w.ZipCode == "" && w.Lon == 0 && w.Lat == 0
}

type WeatherPollerChats struct {
	bun.BaseModel `bun:"weather_poller_chats"`

	ID                     int                      `bun:"id,pk,autoincrement"`
	ChatID                 int64                    `bun:"chat_id,notnull,unique:weather_poller_chats_chat_id_key"`
	PollerLocationID       int                      `bun:"poller_location_id,notnull,unique:weather_poller_chats_chat_id_key"`
	WeatherPollingLocation *WeatherPollingLocations `bun:"rel:belongs-to,join:poller_location_id=id"`
}

type NotificationsDupeCache struct {
	bun.BaseModel `bun:"notifications_dupe_cache"`

	ID         int       `bun:"id,pk,autoincrement"`
	CreatedAt  time.Time `bun:"created_at,notnull,default:current_timestamp"`
	DupeKey    string    `bun:"dupe_key,notnull,unique"`
	DupeTTLEnd time.Time `bun:"dupe_ttl_end,notnull"`
}

type Birthdays struct {
	bun.BaseModel `bun:"birthdays"`

	ID              uuid.UUID `bun:"id,pk"`
	ChatId          int64     `bun:"chat_id,notnull,unique:birthdays_chat_id_name_key"`
	Name            string    `bun:"name,notnull,unique:birthdays_chat_id_name_key"`
	LastAnnouncedAt time.Time `bun:"last_announced_at"`

	//Day month year instead of a timestamp because we don't care about the time
	Day   int `bun:"day,notnull"`
	Month int `bun:"month,notnull"`
	Year  int `bun:"year"`
}

type TelegramUser struct {
	bun.BaseModel `bun:"telegram_users"`

	ID       int64       `bun:"id,pk"`
	UserName string      `bun:"username,notnull"`
	Chats    []*ChatLogs `bun:"rel:has-many,join:id=user_id"`
}

type ChatLogs struct {
	bun.BaseModel `bun:"chat_logs"`

	ID        int            `bun:"id,pk,autoincrement"`
	ChatID    int64          `bun:"chat_id,notnull"`
	UserID    int64          `bun:"user_id,notnull"`
	MessageID int64          `bun:"message_id,notnull"`
	Type      tgapi.TextData `bun:"type,notnull"`
	Message   string         `bun:"message,notnull"`
}

func (c *ChatLogs) BeforeAppendModel(ctx context.Context, query schema.Query) error {
	if !c.Type.IsValid() {
		return fmt.Errorf("invalid type %d", c.Type)
	}

	return nil
}
