package db

import (
	"database/sql"
	"fmt"
	"github.com/tomato3017/tomatobot/pkg/config"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/uptrace/bun/extra/bundebug"
)

func GetDbConnection(dbCfg config.Database) (*bun.DB, error) {
	dbConn, err := openConnection(dbCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	return dbConn, nil
}

func openConnection(dbCfg config.Database) (*bun.DB, error) {
	switch *dbCfg.DbType {
	case config.DBTypeSQLite:
		return openSQLLiteConnection(dbCfg)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbCfg.DbType)
	}
}

func openSQLLiteConnection(dbCfg config.Database) (*bun.DB, error) {
	dbConn, err := sql.Open(sqliteshim.ShimName, dbCfg.ConnectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite connection: %w", err)
	}

	bunDb := bun.NewDB(dbConn, sqlitedialect.New())

	bunDb.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true), bundebug.WithEnabled(dbCfg.LogQueries)))

	return bunDb, nil

}
