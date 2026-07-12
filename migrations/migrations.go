package migrations

import (
	"context"
	"database/sql"
	_ "embed"
	"strings"
)

//go:embed 001_init.up.sql
var schema string

func Up(ctx context.Context, db *sql.DB) error {
	for _, q := range strings.Split(schema, ";") {
		if strings.TrimSpace(q) != "" {
			if _, err := db.ExecContext(ctx, q); err != nil {
				return err
			}
		}
	}
	return nil
}
