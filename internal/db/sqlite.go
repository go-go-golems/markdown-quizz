package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
	_ "modernc.org/sqlite"
)

type SQLiteOptions struct {
	Path string
}

func OpenSQLite(ctx context.Context, opts SQLiteOptions) (*sql.DB, error) {
	if opts.Path == "" {
		return nil, errors.New("sqlite path is required")
	}

	db, err := sql.Open("sqlite", opts.Path)
	if err != nil {
		return nil, errors.Wrap(err, "open sqlite")
	}

	if err := configureSQLite(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}

	if err := MigrateSQLite(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func configureSQLite(ctx context.Context, db *sql.DB) error {
	pragmas := []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA busy_timeout = 5000",
		"PRAGMA journal_mode = WAL",
		"PRAGMA synchronous = NORMAL",
	}

	for _, p := range pragmas {
		if _, err := db.ExecContext(ctx, p); err != nil {
			return errors.Wrap(err, fmt.Sprintf("sqlite pragma failed: %s", p))
		}
	}

	if err := db.PingContext(ctx); err != nil {
		return errors.Wrap(err, "ping sqlite")
	}

	return nil
}
