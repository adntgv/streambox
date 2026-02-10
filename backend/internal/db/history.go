package db

import (
	"fmt"

	"github.com/streambox/backend/internal/models"
)

// GetHistory returns the most recent watch history entries (up to 50).
func (d *DB) GetHistory() ([]models.WatchHistory, error) {
	rows, err := d.db.Query(`
		SELECT id, tmdb_id, title, poster_path, year, duration, progress,
		       completed, quality, magnet_uri, watched_at, updated_at
		FROM watch_history
		ORDER BY updated_at DESC
		LIMIT 50
	`)
	if err != nil {
		return nil, fmt.Errorf("query history: %w", err)
	}
	defer rows.Close()

	return scanHistoryRows(rows)
}

// GetContinueWatching returns movies that are in-progress (not completed, progress > 0).
func (d *DB) GetContinueWatching() ([]models.WatchHistory, error) {
	rows, err := d.db.Query(`
		SELECT id, tmdb_id, title, poster_path, year, duration, progress,
		       completed, quality, magnet_uri, watched_at, updated_at
		FROM watch_history
		WHERE completed = 0 AND progress > 0
		ORDER BY updated_at DESC
		LIMIT 20
	`)
	if err != nil {
		return nil, fmt.Errorf("query continue watching: %w", err)
	}
	defer rows.Close()

	return scanHistoryRows(rows)
}

// UpsertProgress inserts or updates a watch history record for the given movie.
// A movie is marked as completed if progress/duration exceeds 0.9.
func (d *DB) UpsertProgress(tmdbID int, title, posterPath string, year int, duration int, progress float64, quality, magnetURI string) error {
	completed := 0
	if duration > 0 && progress/float64(duration) > 0.9 {
		completed = 1
	}

	_, err := d.db.Exec(`
		INSERT INTO watch_history (tmdb_id, title, poster_path, year, duration, progress, completed, quality, magnet_uri, watched_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(tmdb_id) DO UPDATE SET
			title       = excluded.title,
			poster_path = excluded.poster_path,
			year        = excluded.year,
			duration    = excluded.duration,
			progress    = excluded.progress,
			completed   = excluded.completed,
			quality     = excluded.quality,
			magnet_uri  = excluded.magnet_uri,
			updated_at  = CURRENT_TIMESTAMP
	`, tmdbID, title, posterPath, year, duration, progress, completed, quality, magnetURI)
	if err != nil {
		return fmt.Errorf("upsert progress for tmdb_id %d: %w", tmdbID, err)
	}
	return nil
}

// DeleteHistory removes a watch history entry by TMDB ID.
func (d *DB) DeleteHistory(tmdbID int) error {
	_, err := d.db.Exec("DELETE FROM watch_history WHERE tmdb_id = ?", tmdbID)
	if err != nil {
		return fmt.Errorf("delete history for tmdb_id %d: %w", tmdbID, err)
	}
	return nil
}

// scanHistoryRows is a helper that scans sql.Rows into a slice of WatchHistory.
func scanHistoryRows(rows interface {
	Next() bool
	Scan(dest ...interface{}) error
	Err() error
}) ([]models.WatchHistory, error) {
	var result []models.WatchHistory
	for rows.Next() {
		var h models.WatchHistory
		var completedInt int
		if err := rows.Scan(
			&h.ID, &h.TMDbID, &h.Title, &h.PosterPath, &h.Year,
			&h.Duration, &h.Progress, &completedInt, &h.Quality,
			&h.MagnetURI, &h.WatchedAt, &h.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan history row: %w", err)
		}
		h.Completed = completedInt != 0
		result = append(result, h)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate history rows: %w", err)
	}
	return result, nil
}
