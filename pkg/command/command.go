package command

import (
	"context"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TODO convert to parsing command text and providing to the command
// TODO command filtering based on permissions
type TomatobotCommand interface {
	BaseICommand
	Execute(ctx context.Context, params CommandParams) error
	Description() string
	Help() string
}

// TODO nested commands with middleware appliciation at each level

type CommandParams struct {
	CommandName string
	Args        []string
	Message     *tgbotapi.Message
}
