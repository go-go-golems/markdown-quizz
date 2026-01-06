package db

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"

	"github.com/pkg/errors"
)

//go:embed migrations/0001_init.sql
var migration0001Init string

type sqliteMigration struct {
	version int
	name    string
	upSQL   string
}

var sqliteMigrations = []sqliteMigration{
	{
		version: 1,
		name:    "init schema",
		upSQL:   migration0001Init,
	},
}

func MigrateSQLite(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "begin sqlite migration transaction")
	}
	defer func() {
		_ = tx.Rollback()
	}()

	currentVersion, err := getSQLiteUserVersion(ctx, tx)
	if err != nil {
		return err
	}

	for _, m := range sqliteMigrations {
		if m.version <= currentVersion {
			continue
		}

		if _, err := tx.ExecContext(ctx, m.upSQL); err != nil {
			return errors.Wrap(err, fmt.Sprintf("apply sqlite migration %d (%s)", m.version, m.name))
		}
		if _, err := tx.ExecContext(ctx, fmt.Sprintf("PRAGMA user_version = %d", m.version)); err != nil {
			return errors.Wrap(err, fmt.Sprintf("set sqlite user_version=%d", m.version))
		}
		currentVersion = m.version
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "commit sqlite migrations")
	}

	return nil
}

func getSQLiteUserVersion(ctx context.Context, q sqlQueryer) (int, error) {
	var version int
	row := q.QueryRowContext(ctx, "PRAGMA user_version")
	if err := row.Scan(&version); err != nil {
		return 0, errors.Wrap(err, "read sqlite user_version")
	}
	return version, nil
}

type sqlQueryer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}
