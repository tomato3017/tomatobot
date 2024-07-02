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
