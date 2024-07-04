package command

import (
	"context"
	"fmt"
)

type BaseICommand interface {
	RegisterSubcommand(cmdname string, cmd TomatobotCommand) error
	Execute(ctx context.Context, params CommandParams) error
	CmdName() string
}

type BaseCommand struct {
	name        string
	middleware  []interface{}               //TODO middleware
	subCommands map[string]TomatobotCommand //TODO nested commands

	parent BaseICommand
}

func (b *BaseCommand) CmdName() string {
	return b.name
}

func (b *BaseCommand) RegisterSubcommand(cmdname string, cmd TomatobotCommand) error {
	if _, ok := b.subCommands[cmdname]; ok {
		return fmt.Errorf("subcommand %s already exists", cmdname)
	}
	b.subCommands[cmdname] = cmd

	return nil
}

func (b *BaseCommand) newBaseSubCommand(cmdName string) BaseCommand {
	bCmd := NewBaseCommand()
	bCmd.parent = b

	return bCmd
}

func (b *BaseCommand) Execute(ctx context.Context, params CommandParams) error {
	if len(params.Args) == 0 {
		return fmt.Errorf("no subcommand provided")
	}

	cmdname := params.Args[0]
	cmd, ok := b.subCommands[cmdname]
	if !ok {
		return fmt.Errorf("subcommand %s not found", cmdname)
	}

	return cmd.Execute(ctx, CommandParams{
		CommandName: cmdname,
		Args:        params.Args[1:],
		Message:     params.Message,
	})

}

func NewBaseCommand() BaseCommand {
	return BaseCommand{
		name:        "",
		middleware:  make([]interface{}, 0),
		subCommands: make(map[string]TomatobotCommand),
	}
}
