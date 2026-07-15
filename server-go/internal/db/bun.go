package db

import (
	"database/sql"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/extra/bundebug"
)

func NewBunDB(sqlDB *sql.DB) *bun.DB {
	bunDB := bun.NewDB(sqlDB, pgdialect.New())
	bunDB.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithVerbose(false),
	))
	return bunDB
}
