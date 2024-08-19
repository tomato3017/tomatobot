package birthday

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	dbmodels "github.com/tomato3017/tomatobot/pkg/bot/models/db"
	"github.com/tomato3017/tomatobot/pkg/command"
	"github.com/tomato3017/tomatobot/pkg/command/middleware"
	"github.com/tomato3017/tomatobot/pkg/command/models"
	"github.com/tomato3017/tomatobot/pkg/util"
	"github.com/uptrace/bun"
	"strings"
)

var _ command.TomatobotCommand = &birthdayCmdList{}

type birthdayCmdList struct {
	command.BaseCommand
	dbConn bun.IDB
	logger zerolog.Logger
}

func newBirthdayListCmd(dbConn bun.IDB, logger zerolog.Logger) command.TomatobotCommand {
	return &birthdayCmdList{
		BaseCommand: command.NewBaseCommand(middleware.WithMaxArgs(0)),
		dbConn:      dbConn,
		logger:      logger,
	}
}

func (b *birthdayCmdList) Execute(ctx context.Context, params models.CommandParams) error {
	//get all birthdays
	var birthdays []dbmodels.Birthdays
	err := b.dbConn.NewSelect().Model(&birthdays).
		Where("chat_id = ?", params.Message.AssumedChatID()).
		Scan(ctx)
	if err != nil {
		return fmt.Errorf("failed to get birthdays: %w", err)
	}

	if len(birthdays) == 0 {
		_, err = params.BotProxy.Send(util.NewMessageReply(params.Message.InnerMsg(), "", "No birthdays found"))
		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
		return nil
	}

	msg := strings.Builder{}
	msg.WriteString("Birthdays:\n")
	msg.WriteString("ID - Name: YYYY-MM-DD\n")
	for _, birthday := range birthdays {
		msg.WriteString(fmt.Sprintf("%s - %s: %04d-%02d-%02d\n", birthday.ID, birthday.Name, birthday.Year, birthday.Month, birthday.Day))
	}

	_, err = params.BotProxy.Send(util.NewMessageReply(params.Message.InnerMsg(), "", msg.String()))
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func (b *birthdayCmdList) Description() string {
	return "List all birthdays"
}

func (b *birthdayCmdList) Help() string {
	return "/birthday list - List all birthdays"
}
