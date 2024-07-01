package command

import (
	"context"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TODO convert to parsing command text and providing to the command
type TomatobotCommand interface {
	Execute(ctx context.Context, msg *tgbotapi.Message) error
	Description() string
	Help() string
}
