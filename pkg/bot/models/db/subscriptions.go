package db

import "github.com/uptrace/bun"

type Subscriptions struct {
	bun.BaseModel `bun:"subscriptions"`

	ID           int64  `bun:"id,pk,autoincrement"`
	ChatID       int64  `bun:"chat_id,notnull,unique:subscriptions_chat_id_topic_pattern_key"`
	TopicPattern string `bun:"topic_pattern,notnull,unique:subscriptions_chat_id_topic_pattern_key"`
}
