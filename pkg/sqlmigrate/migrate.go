package sqlmigrate

import (
	"context"
	"fmt"
	dbmodels "github.com/tomato3017/tomatobot/pkg/bot/models/db"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
	"strings"
	"time"
)

func MigrateDbSchema(ctx context.Context, db *bun.DB) (int, error) {

	migrations := migrate.NewMigrations()

	// Create subscriptions table
	migrations.Add(migrate.Migration{
		Name: "00001_create_subscriptions_table",
		Up: func(ctx context.Context, db *bun.DB) error {
			_, err := db.NewCreateTable().
				Model((*dbmodels.Subscriptions)(nil)).
				IfNotExists().
				Exec(ctx)
			return err
		},
		Down: func(ctx context.Context, db *bun.DB) error {
			_, err := db.NewDropTable().
				Model((*dbmodels.Subscriptions)(nil)).
				IfExists().
				Exec(ctx)
			return err
		},
	})

	// Create Weather Polling table

	migrations.Add(migrate.Migration{
		Name: "00002_create_weather_polling_table",
		Up: func(ctx context.Context, db *bun.DB) error {
			_, err := db.NewCreateTable().
				Model((*dbmodels.WeatherPollingLocations)(nil)).
				IfNotExists().
				Exec(ctx)
			if err != nil {
				return err
			}

			_, err = db.NewCreateTable().
				Model((*dbmodels.WeatherPollerChats)(nil)).
				IfNotExists().
				Exec(ctx)
			return err
		},
		Down: func(ctx context.Context, db *bun.DB) error {
			_, err := db.NewDropTable().
				Model((*dbmodels.WeatherPollerChats)(nil)).
				IfExists().
				Exec(ctx)
			if err != nil {
				return err
			}

			_, err = db.NewDropTable().
				Model((*dbmodels.WeatherPollingLocations)(nil)).
				IfExists().
				Exec(ctx)
			return err
		},
	})

	migrations.Add(migrate.Migration{
		Name: "00003_create_dedupe_table",
		Up: func(ctx context.Context, db *bun.DB) error {
			_, err := db.NewCreateTable().
				Model((*dbmodels.NotificationsDupeCache)(nil)).
				IfNotExists().
				Exec(ctx)
			return err
		},
		Down: func(ctx context.Context, db *bun.DB) error {
			_, err := db.NewDropTable().
				Model((*dbmodels.NotificationsDupeCache)(nil)).
				IfExists().
				Exec(ctx)
			return err
		},
	})

	migrations.Add(migrate.Migration{
		Name: "00004_create_birthdays_table",
		Up: func(ctx context.Context, db *bun.DB) error {
			_, err := db.NewCreateTable().
				Model((*dbmodels.Birthdays)(nil)).
				IfNotExists().
				Exec(ctx)

			return err
		},
		Down: func(ctx context.Context, db *bun.DB) error {
			_, err := db.NewDropTable().
				Model((*dbmodels.Birthdays)(nil)).
				IfExists().
				Exec(ctx)
			return err
		},
	})

	migrations.Add(migrate.Migration{
		Name: "00005_create_chat_log_table",
		Up: func(ctx context.Context, db *bun.DB) error {
			_, err := db.NewCreateTable().
				Model((*dbmodels.TelegramUser)(nil)).
				IfNotExists().
				Exec(ctx)
			if err != nil {
				return err
			}

			_, err = db.NewCreateTable().
				Model((*dbmodels.ChatLogs)(nil)).
				IfNotExists().
				Exec(ctx)
			return err
		},
		Down: func(ctx context.Context, db *bun.DB) error {
			_, err := db.NewDropTable().
				Model((*dbmodels.ChatLogs)(nil)).
				IfExists().
				Exec(ctx)
			return err
		},
	})

	migrations.Add(migrate.Migration{
		Name: "00006_add_tz_to_birthdays",
		Up: func(ctx context.Context, db *bun.DB) error {
			err := db.NewSelect().Model((*dbmodels.Birthdays)(nil)).Column("tz").Limit(1).Scan(ctx)
			if err == nil || strings.Contains(err.Error(), "no rows in result set") {
				return nil
			}

			_, err = db.NewAddColumn().
				Model((*dbmodels.Birthdays)(nil)).
				ColumnExpr("tz VARCHAR Default 'America/New_York'").Exec(ctx)

			return err
		},
		Down: func(ctx context.Context, db *bun.DB) error {
			_, err := db.NewDropColumn().
				Model((*dbmodels.Birthdays)(nil)).
				ColumnExpr("tz").Exec(ctx)

			return err
		},
	})

	ctx, cf := context.WithTimeout(ctx, 30*time.Second)
	defer cf()

	migrator := migrate.NewMigrator(db, migrations)

	if err := migrator.Init(ctx); err != nil {
		return 0, fmt.Errorf("failed to initialize migrator: %w", err)
	}

	mGroup, err := migrator.Migrate(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to migrate database schema: %w", err)
	}

	return len(mGroup.Migrations), nil
}
