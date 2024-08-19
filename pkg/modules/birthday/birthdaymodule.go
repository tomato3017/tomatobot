package birthday

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/tomato3017/tomatobot/pkg/modules"
	"github.com/tomato3017/tomatobot/pkg/notifications"
	"github.com/uptrace/bun"
)

type BirthdayModule struct {
	dbConn bun.IDB

	publisher notifications.Publisher

	logger zerolog.Logger
	poller *poller
}

func (b *BirthdayModule) Initialize(ctx context.Context, params modules.InitializeParameters) error {
	b.dbConn = params.DbConn
	b.logger = params.Logger
	b.publisher = params.Notifications
	poller, err := newPoller(b.publisher, b.dbConn, b.logger)
	if err != nil {
		return fmt.Errorf("failed to create birthday poller: %w", err)
	}
	b.poller = poller

	//register commands
	cmd, err := newBirthdayCmd(b.dbConn, b.logger, b.publisher)
	if err != nil {
		return fmt.Errorf("failed to create birthday command: %w", err)
	}

	err = params.Tomatobot.RegisterCommand("birthday", cmd)
	if err != nil {
		return err
	}

	return nil
}

func (b *BirthdayModule) Start(ctx context.Context) error {
	b.logger.Debug().Msg("Starting birthday module")
	b.poller.Start(ctx)

	return nil
}

func (b *BirthdayModule) Shutdown(ctx context.Context) error {
	b.logger.Debug().Msg("Shutting down birthday module")
	b.poller.Stop()

	return nil
}

func (b *BirthdayModule) birthdayCheck(ctx context.Context) error {
	//TODO
	return nil
}
