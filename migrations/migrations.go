package migrations

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"strings"

	"github.com/go-sql-driver/mysql"
)

//go:embed 001_init.up.sql
var schema string

func Up(ctx context.Context, db *sql.DB) error {
	if err := run(ctx, db, schema); err != nil {
		return err
	}
	return ensureCurrency(ctx, db)
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

func ensureCurrency(ctx context.Context, db *sql.DB) error {
	var exists bool
	if err := db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'products' AND column_name = 'currency')`).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return nil
	}
	_, err := db.ExecContext(ctx, `ALTER TABLE products ADD COLUMN currency CHAR(3) NOT NULL DEFAULT 'USD' AFTER image_key`)
	var mysqlError *mysql.MySQLError
	if errors.As(err, &mysqlError) && mysqlError.Number == 1060 {
		return nil
	}
	return err
}
