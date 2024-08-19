package birthday

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/tomato3017/tomatobot/pkg/command"
	"github.com/tomato3017/tomatobot/pkg/command/middleware"
	"github.com/tomato3017/tomatobot/pkg/notifications"
	"github.com/uptrace/bun"
)

type BirthdayCmd struct {
	command.BaseCommand
	dbConn bun.IDB
	logger zerolog.Logger
}

var _ command.TomatobotCommand = &BirthdayCmd{}

func (b *BirthdayCmd) Description() string {
	return "Subscribe to birthday notifications"
}

func (b *BirthdayCmd) Help() string {
	return "/birthday - Subscribe to birthday notifications"
}

func newBirthdayCmd(dbConn bun.IDB, logger zerolog.Logger, publisher notifications.Publisher) (*BirthdayCmd, error) {
	birthdayCmd := BirthdayCmd{
		BaseCommand: command.NewBaseCommand(middleware.WithAdminPermission()),
		dbConn:      dbConn,
		logger:      logger,
	}

	err := birthdayCmd.RegisterSubcommand("list", newBirthdayListCmd(dbConn, logger))
	if err != nil {
		return nil, fmt.Errorf("unable to register subcommand %s. Err: %w", "list", err)
	}

	//birthday add <name> <year> <month> <day>
	err = birthdayCmd.RegisterSubcommand("add", newBirthdayAddCmd(dbConn, logger, publisher), middleware.WithNArgs(2))
	if err != nil {
		return nil, fmt.Errorf("unable to register subcommand %s. Err: %w", "add", err)
	}

	err = birthdayCmd.RegisterSubcommand("remove", newBirthdayRemoveCmd(dbConn, logger, publisher))
	if err != nil {
		return nil, fmt.Errorf("unable to register subcommand %s. Err: %w", "remove", err)
	}

	return &birthdayCmd, nil
}
