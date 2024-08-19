package birthday

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	dbmodels "github.com/tomato3017/tomatobot/pkg/bot/models/db"
	"github.com/tomato3017/tomatobot/pkg/command"
	"github.com/tomato3017/tomatobot/pkg/command/middleware"
	"github.com/tomato3017/tomatobot/pkg/command/models"
	"github.com/tomato3017/tomatobot/pkg/notifications"
	"github.com/tomato3017/tomatobot/pkg/util"
	"github.com/uptrace/bun"
)

type BirthdayCmdRemove struct {
	command.BaseCommand
	dbConn    bun.IDB
	logger    zerolog.Logger
	publisher notifications.Publisher
}

var _ command.TomatobotCommand = &BirthdayCmdRemove{}

func newBirthdayRemoveCmd(dbConn bun.IDB, logger zerolog.Logger, publisher notifications.Publisher) command.TomatobotCommand {
	return &BirthdayCmdRemove{
		BaseCommand: command.NewBaseCommand(middleware.WithNArgs(1)),
		dbConn:      dbConn,
		logger:      logger,
		publisher:   publisher,
	}
}

func (b *BirthdayCmdRemove) Execute(ctx context.Context, params models.CommandParams) error {
	targetUUID, err := uuid.Parse(params.Args[0])
	if err != nil {
		return fmt.Errorf("failed to parse UUID: %w", err)
	}

	count, err := b.dbConn.NewSelect().Model(&dbmodels.Birthdays{}).Where("id = ? AND chat_id = ?", targetUUID, params.Message.AssumedChatID()).Count(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if birthday exists: %w", err)
	}
	if count == 0 {
		return errors.New("birthday not found")
	}

	_, err = b.dbConn.NewDelete().Model(&dbmodels.Birthdays{}).Where("id = ? AND chat_id = ?", targetUUID, params.Message.AssumedChatID()).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to remove birthday: %w", err)
	}

	_, _ = params.BotProxy.Send(util.NewMessagePrivate(params.Message.InnerMsg(), "", "Birthday removed"))

	return nil
}

func (b *BirthdayCmdRemove) Description() string {
	return "Remove a birthday"
}

func (b *BirthdayCmdRemove) Help() string {
	return "/birthday remove <uuid> - Remove a birthday"
}
