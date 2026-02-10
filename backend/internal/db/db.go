package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// DB wraps the SQLite database connection and provides data access methods.
type DB struct {
	db *sql.DB
}

// New opens (or creates) the SQLite database at dbPath, runs migrations,
// and returns a ready-to-use DB handle.
func New(dbPath string) (*DB, error) {
	sqlDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// Enable WAL mode for better concurrent read performance.
	if _, err := sqlDB.Exec("PRAGMA journal_mode=WAL"); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("set wal mode: %w", err)
	}

	// Enable foreign keys.
	if _, err := sqlDB.Exec("PRAGMA foreign_keys=ON"); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	d := &DB{db: sqlDB}
	if err := d.migrate(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return d, nil
}

// Close closes the underlying database connection.
func (d *DB) Close() error {
	return d.db.Close()
}

// migrate creates the required tables if they do not already exist.
func (d *DB) migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS stream_sessions (
			id            TEXT PRIMARY KEY,
			tmdb_id       INTEGER NOT NULL,
			title         TEXT NOT NULL,
			magnet_uri    TEXT NOT NULL,
			info_hash     TEXT NOT NULL,
			file_path     TEXT,
			file_size     INTEGER DEFAULT 0,
			content_type  TEXT DEFAULT '',
			status        TEXT DEFAULT 'starting',
			created_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at    DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS watch_history (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			tmdb_id     INTEGER NOT NULL UNIQUE,
			title       TEXT NOT NULL,
			poster_path TEXT DEFAULT '',
			year        INTEGER DEFAULT 0,
			duration    INTEGER DEFAULT 0,
			progress    REAL DEFAULT 0,
			completed   INTEGER DEFAULT 0,
			quality     TEXT DEFAULT '',
			magnet_uri  TEXT DEFAULT '',
			watched_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS torrent_cache (
			info_hash   TEXT PRIMARY KEY,
			tmdb_id     INTEGER NOT NULL,
			magnet_uri  TEXT NOT NULL,
			title       TEXT NOT NULL,
			file_path   TEXT DEFAULT '',
			file_size   INTEGER DEFAULT 0,
			last_used   DATETIME DEFAULT CURRENT_TIMESTAMP,
			created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, m := range migrations {
		if _, err := d.db.Exec(m); err != nil {
			return fmt.Errorf("exec migration: %w", err)
		}
	}

	return nil
}
