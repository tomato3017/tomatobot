package command

import (
	"context"
	"github.com/tomato3017/tomatobot/pkg/command/models"
)

type CommandCallback func(ctx context.Context, params models.CommandParams) error

type SimpleCommand struct {
	BaseCommand
	callback CommandCallback
	desc     string
	help     string
}

var _ TomatobotCommand = &SimpleCommand{}

func NewSimpleCommand(callback CommandCallback, desc, help string) *SimpleCommand {
	return &SimpleCommand{
		BaseCommand: NewBaseCommand(),
		callback:    callback,
		desc:        desc,
		help:        help,
	}
}

func (s *SimpleCommand) Execute(ctx context.Context, params models.CommandParams) error {
	return s.callback(ctx, params)
}

func (s *SimpleCommand) Description() string {
	return s.desc
}

func (s *SimpleCommand) Help() string {
	return s.help
}
