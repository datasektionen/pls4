package database

import (
	"context"
	"database/sql"
	"embed"
	"log/slog"
	"sort"
)

//go:embed migrations/*.sql
var migrations embed.FS

func Migrate(db *sql.DB, ctx context.Context) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS __migrations (
			name  	   TEXT PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL DEFAULT NOW()
		);
	`); err != nil {
		return err
	}
	entries, err := migrations.ReadDir("migrations")
	if err != nil {
		return err
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})
	for _, entry := range entries {
		migration, err := migrations.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return err
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		if err := tx.QueryRowContext(ctx, `
			SELECT FROM __migrations WHERE name = $1
		`, entry.Name()).Scan(); err == nil {
			// No error means we found the migration, meaning it's already applied.
			continue
		} else if err != sql.ErrNoRows {
			return err
		}

		slog.InfoContext(ctx, "Applying migration", "name", entry.Name())

		if _, err := tx.ExecContext(ctx, string(migration)); err != nil {
			tx.Rollback()
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO __migrations (name) VALUES ($1)
		`, entry.Name()); err != nil {
			tx.Rollback()
			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}
