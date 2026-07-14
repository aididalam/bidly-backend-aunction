package migrations

import (
	"context"
	"database/sql"
	_ "embed"
	"strings"
)

//go:embed 001_init.up.sql
var schema string

//go:embed 002_currency.up.sql
var currencySchema string

func Up(ctx context.Context, db *sql.DB) error {
	return run(ctx, db, schema, currencySchema)
}

func run(ctx context.Context, db *sql.DB, scripts ...string) error {
	for _, script := range scripts {
		for _, q := range strings.Split(script, ";") {
			if strings.TrimSpace(q) != "" {
				if _, err := db.ExecContext(ctx, q); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
