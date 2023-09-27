package database

import (
	"context"
	"database/sql"
	"embed"
	"errors"
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
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	res, err := tx.Query(`
		SELECT name
		FROM __migrations
	`)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	for res.Next() {
		var applied string
		res.Scan(&applied)
		found := false
		for _, entry := range entries {
			if entry.Name() == applied {
				found = true
				break
			}
		}
		if !found {
			_ = tx.Rollback()
			return errors.New("Applied migration " + applied + " not found in file system")
		}
	}
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	for _, entry := range entries {
		migration, err := migrations.ReadFile("migrations/" + entry.Name())
		if err != nil {
			_ = tx.Rollback()
			return err
		}

		if err := tx.QueryRow(`
			SELECT FROM __migrations WHERE name = $1
		`, entry.Name()).Scan(); err == nil {
			// No error means we found the migration, meaning it's already applied.
			continue
		} else if err != sql.ErrNoRows {
			_ = tx.Rollback()
			return err
		}

		slog.InfoContext(ctx, "Applying migration", "name", entry.Name())

		if _, err := tx.Exec(string(migration)); err != nil {
			_ = tx.Rollback()
			return err
		}
		if _, err := tx.Exec(`
			INSERT INTO __migrations (name) VALUES ($1)
		`, entry.Name()); err != nil {
			_ = tx.Rollback()
			return err
		}

	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
