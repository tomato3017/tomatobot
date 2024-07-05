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
				Model((*dbmodels.WeatherPollingLocation)(nil)).
				IfNotExists().
				Exec(ctx)
			return err
		},
		Down: func(ctx context.Context, db *bun.DB) error {
			_, err := db.NewDropTable().
				Model((*dbmodels.WeatherPollingLocation)(nil)).
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
