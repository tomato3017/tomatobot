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

type WeatherPollingLocation struct {
	bun.BaseModel `bun:"weather_polling_locations"`

	ID      int     `bun:"id,pk"`
	Name    string  `bun:"name,notnull"`
	Country string  `bun:"country,notnull"`
	ZipCode string  `bun:"zip_code,notnull,unique:weather_polling_locations_zip_code_key"`
	Lon     float64 `bun:"lon,notnull"`
	Lat     float64 `bun:"lat,notnull"`
	Polling bool    `bun:"polling,notnull,default:true"`
}
