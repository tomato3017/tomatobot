package command

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type SimpleCommand struct {
	callback func(ctx context.Context, msg *tgbotapi.Message) error
	desc     string
	help     string
}

func NewSimpleCommand(callback func(ctx context.Context, msg *tgbotapi.Message) error, desc, help string) *SimpleCommand {
	return &SimpleCommand{
		callback: callback,
		desc:     desc,
		help:     help,
	}
}

func (s *SimpleCommand) Execute(ctx context.Context, msg *tgbotapi.Message) error {
	return s.callback(ctx, msg)
}

func (s *SimpleCommand) Description() string {
	return s.desc
}

func (s *SimpleCommand) Help() string {
	return s.help
}
