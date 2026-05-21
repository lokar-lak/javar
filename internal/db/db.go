package db

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/glebarez/sqlite"
)

// Open opens SQLite DB and runs migration.
func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// SQLite works better with a single write connection
	db.SetMaxOpenConns(1)

	if err := ping(db); err != nil {
		return nil, err
	}

	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return db, nil
}

func ping(db *sql.DB) error {
	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping sqlite: %w", err)
	}

	// Migration: add orthography column if it does not exist yet
	db.Exec(`ALTER TABLE translations ADD COLUMN orthography TEXT NOT NULL DEFAULT 'academic'`)
	return nil
}

func migrate(db *sql.DB) error {
	// Read SQL file and execute it
	data, err := os.ReadFile("migrations/001_init.sql")
	if err != nil {
		return fmt.Errorf("read migration: %w", err)
	}
	if _, err := db.Exec(string(data)); err != nil {
		// Ignore "already exists" errors on repeated runs
		// In real projects, use goose or migrate
		_ = err
	}

	// Backfill schema for existing databases: SteamDB source link.
	if _, err := db.Exec(`ALTER TABLE games ADD COLUMN steamdb_url TEXT NOT NULL DEFAULT ''`); err != nil {
		// SQLite returns duplicate column error if this column already exists.
		if !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
			return fmt.Errorf("add steamdb_url: %w", err)
		}
	}

	// Remove legacy page-specific cover column if present.
	if _, err := db.Exec(`ALTER TABLE games DROP COLUMN page_cover_url`); err != nil {
		errMsg := strings.ToLower(err.Error())
		if !strings.Contains(errMsg, "no such column") && !strings.Contains(errMsg, "syntax error") {
			return fmt.Errorf("drop page_cover_url: %w", err)
		}
	}

	// Remove legacy translations version column if present.
	if _, err := db.Exec(`ALTER TABLE translations DROP COLUMN version`); err != nil {
		errMsg := strings.ToLower(err.Error())
		if !strings.Contains(errMsg, "no such column") && !strings.Contains(errMsg, "syntax error") {
			return fmt.Errorf("drop translations.version: %w", err)
		}
	}

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS translation_submissions (
			id                INTEGER PRIMARY KEY AUTOINCREMENT,
			game_title        TEXT    NOT NULL,
			platforms         TEXT    NOT NULL DEFAULT '[]',
			category          TEXT    NOT NULL CHECK(category IN ('official','unofficial')),
			localization_type TEXT    NOT NULL DEFAULT '[]',
			authors           TEXT    NOT NULL,
			game_url          TEXT    NOT NULL,
			translation_url   TEXT    NOT NULL,
			description       TEXT    NOT NULL DEFAULT '',
			status            TEXT    NOT NULL DEFAULT 'new' CHECK(status IN ('new','accepted','rejected')),
			created_at        DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at        DATETIME DEFAULT CURRENT_TIMESTAMP
		)`); err != nil {
		return fmt.Errorf("create translation_submissions: %w", err)
	}
	return nil
}
