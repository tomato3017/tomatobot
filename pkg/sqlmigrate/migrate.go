package sqlmigrate

import (
	"context"
	"fmt"
	dbmodels "github.com/tomato3017/tomatobot/pkg/bot/models/db"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
	"time"
)

func MigrateDbSchema(ctx context.Context, db *bun.DB) (int, error) {

	migrations := migrate.NewMigrations()

	// Create subscriptions table
	migrations.Add(migrate.Migration{
		Name: "create_subscriptions_table",
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
		Name: "create_weather_polling_table",
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
		Name: "create_dedupe_table",
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
		Name: "create_birthdays_table",
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
