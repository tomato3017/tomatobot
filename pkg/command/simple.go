package command

import (
	"context"
)

type CommandCallback func(ctx context.Context, params CommandParams) error

type SimpleCommand struct {
	callback CommandCallback
	desc     string
	help     string
}

var _ TomatobotCommand = &SimpleCommand{}

func NewSimpleCommand(callback CommandCallback, desc, help string) *SimpleCommand {
	return &SimpleCommand{
		callback: callback,
		desc:     desc,
		help:     help,
	}
}

func (s *SimpleCommand) Execute(ctx context.Context, params CommandParams) error {
	return s.callback(ctx, params)
}

func (s *SimpleCommand) Description() string {
	return s.desc
}

func (s *SimpleCommand) Help() string {
	return s.help
}
