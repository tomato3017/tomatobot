package command

import (
	"context"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TODO convert to parsing command text and providing to the command
// TODO command filtering based on permissions
type TomatobotCommand interface {
	Execute(ctx context.Context, params CommandParams) error
	Description() string
	Help() string
}

type CommandParams struct {
	CommandName string
	Args        []string
	Message     *tgbotapi.Message
}
