package db

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
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
