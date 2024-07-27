package command

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tomato3017/tomatobot/pkg/command/middleware"
	"github.com/tomato3017/tomatobot/pkg/command/models"
	"github.com/tomato3017/tomatobot/pkg/util"
)

// TODO redo this entire pattern, I don't like it but it's what I have for now
type BaseICommand interface {
	RegisterSubcommand(cmdname string, cmd TomatobotCommand, middlewareFuncs ...middleware.MiddlewareFunc) error
	RunMiddleware(ctx context.Context, params models.CommandParams) error
	Execute(ctx context.Context, params models.CommandParams) error
	CmdName() string
}

type BaseCommand struct {
	name        string
	middleware  []middleware.MiddlewareFunc
	subCommands map[string]TomatobotCommand
}

var _ BaseICommand = &BaseCommand{}

func (b *BaseCommand) CmdName() string {
	return b.name
}

func (b *BaseCommand) RegisterSubcommand(cmdname string, cmd TomatobotCommand, middlewareFuncs ...middleware.MiddlewareFunc) error {
	if _, ok := b.subCommands[cmdname]; ok {
		return fmt.Errorf("subcommand %s already exists", cmdname)
	}
	b.subCommands[cmdname] = cmd

	return nil
}

func (b *BaseCommand) RunMiddleware(ctx context.Context, params models.CommandParams) error {
	for i := 0; i < len(b.middleware); i++ {
		if err := b.middleware[i](ctx, params); err != nil {
			return err
		}
	}

	return nil
}

func (b *BaseCommand) Execute(ctx context.Context, params models.CommandParams) error {
	//Run any middleware I have on this command
	if err := b.RunMiddleware(ctx, params); err != nil {
		return err
	}

	if len(params.Args) == 0 {
		return b.printHelp(ctx, params)
	}

	cmdname := params.Args[0]
	cmd, ok := b.subCommands[cmdname]
	if !ok {
		return b.printHelp(ctx, params)
	}

	newParams := models.CommandParams{
		CommandName: params.CommandName,
		Args:        params.Args[1:],
		Message:     params.Message,
		BotProxy:    params.BotProxy,
	}

	bCmd, ok := cmd.(BaseICommand)
	if ok {
		if err := bCmd.RunMiddleware(ctx, newParams); err != nil {
			return err
		}
	}

	return cmd.Execute(ctx, newParams)

}

func (b *BaseCommand) printHelp(ctx context.Context, params models.CommandParams) error {
	helpMsg := "Available subcommands:\n"
	for name, cmd := range b.subCommands {
		helpMsg += fmt.Sprintf("\\- * %s * \\- %s\n", name, cmd.Description())
	}

	_, err := params.BotProxy.Send(util.NewMessageReply(params.Message, tgbotapi.ModeMarkdownV2, helpMsg))
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func NewBaseCommand(middlewareFuncs ...middleware.MiddlewareFunc) BaseCommand {
	return BaseCommand{
		name:        "",
		middleware:  middlewareFuncs,
		subCommands: make(map[string]TomatobotCommand),
	}
}
