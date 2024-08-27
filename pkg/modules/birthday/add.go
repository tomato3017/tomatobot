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
	"strings"
	"time"
)

type birthdayCmdAdd struct {
	command.BaseCommand
	dbConn    bun.IDB
	logger    zerolog.Logger
	publisher notifications.Publisher
}

var _ command.TomatobotCommand = &birthdayCmdAdd{}

func newBirthdayAddCmd(dbConn bun.IDB, logger zerolog.Logger, publisher notifications.Publisher) command.TomatobotCommand {
	return &birthdayCmdAdd{
		BaseCommand: command.NewBaseCommand(middleware.WithNArgs(2)),
		dbConn:      dbConn,
		logger:      logger,
		publisher:   publisher,
	}
}

// Execute adds a birthday to the database
// /birthday add <name> <YYYY-MM-DD>
func (b *birthdayCmdAdd) Execute(ctx context.Context, params models.CommandParams) error {
	name := strings.Join(params.Args[:len(params.Args)-1], " ")
	date := params.Args[len(params.Args)-1]

	birthDate, err := time.Parse(time.DateOnly, date)
	if err != nil {
		b.logger.Error().Err(err).Msg("Failed to parse date")
		return fmt.Errorf("failed to parse date, ensure the date is in the format YYYY-MM-DD")
	}

	if birthDate.After(time.Now()) {
		return fmt.Errorf("birth date is in the future")
	}

	b.logger.Debug().Msgf("Adding birthday for %s on %s", name, birthDate)

	dbBday := dbmodels.Birthdays{
		ID:     uuid.New(),
		ChatId: params.Message.AssumedChatID(),
		Name:   strings.ToLower(name),
		Day:    birthDate.Day(),
		Month:  int(birthDate.Month()),
		Year:   birthDate.Year(),
	}

	_, err = b.dbConn.NewInsert().Model(&dbBday).Exec(ctx)
	if err != nil && strings.Contains(err.Error(), "constraint") {
		return fmt.Errorf("birthday already exists")
	} else if err != nil {
		b.logger.Error().Err(err).Msg("Failed to insert birthday")
		return fmt.Errorf("failed to add birthday")
	}

	_, err = b.publisher.Subscribe(notifications.Subscriber{
		TopicPattern: fmt.Sprintf("%s.%d", BirthdayPollerTopic, params.Message.AssumedChatID()),
		ChatId:       params.Message.AssumedChatID(),
	})

	if err != nil && errors.Is(err, notifications.ErrSubExists) {
		b.logger.Trace().Msg("Already subscribed to birthday topic")
	} else if err != nil {
		return fmt.Errorf("failed to subscribe to birthday topic")
	}

	_, err = params.BotProxy.Send(util.NewMessageReply(params.Message.InnerMsg(), "", "Birthday added"))

	return nil
}

func (b *birthdayCmdAdd) Description() string {
	return "Add a birthday"
}

func (b *birthdayCmdAdd) Help() string {
	return "/birthday add <name> <YYYY-MM-DD> - Add a birthday"
}
