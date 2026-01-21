// Package store provides persistence for FIT file tracking.
package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// Store provides persistence for FIT file tracking.
type Store struct {
	db *sql.DB
}

// New opens or creates a SQLite database at the given path.
func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Enable WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable WAL: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return s, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// migrate creates tables if they don't exist.
func (s *Store) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS fit_files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		path TEXT UNIQUE NOT NULL,
		hash TEXT,
		size INTEGER,
		discovered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		source TEXT DEFAULT 'watch',
		
		-- Parsed FIT metadata
		activity_type TEXT,
		activity_name TEXT,
		started_at TIMESTAMP,
		duration_secs INTEGER,
		distance_m REAL,
		calories INTEGER,
		avg_power_w INTEGER,
		max_power_w INTEGER,
		norm_power_w INTEGER,
		avg_hr INTEGER,
		max_hr INTEGER,
		avg_cadence INTEGER,
		avg_speed_mps REAL,
		total_ascent_m REAL,
		device_name TEXT,
		software_version TEXT
	);

	CREATE TABLE IF NOT EXISTS sync_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		file_id INTEGER NOT NULL REFERENCES fit_files(id),
		consumer TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending',
		attempted_at TIMESTAMP,
		completed_at TIMESTAMP,
		remote_id TEXT,
		remote_url TEXT,
		error TEXT,
		retries INTEGER DEFAULT 0,
		UNIQUE(file_id, consumer)
	);

	CREATE TABLE IF NOT EXISTS consumers (
		name TEXT PRIMARY KEY,
		enabled BOOLEAN DEFAULT 0,
		config_json TEXT,
		last_sync TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_sync_pending ON sync_records(consumer, status) WHERE status = 'pending';
	CREATE INDEX IF NOT EXISTS idx_sync_failed ON sync_records(status) WHERE status = 'failed';
	CREATE INDEX IF NOT EXISTS idx_files_hash ON fit_files(hash);
	CREATE INDEX IF NOT EXISTS idx_files_started ON fit_files(started_at);
	CREATE INDEX IF NOT EXISTS idx_files_type ON fit_files(activity_type);
	`
	_, err := s.db.Exec(schema)
	return err
}

// InsertFile adds a new FIT file to the database.
// Returns the file ID.
func (s *Store) InsertFile(ctx context.Context, f *FitFile) (int64, error) {
	result, err := s.db.ExecContext(ctx, `
		INSERT INTO fit_files (
			path, hash, size, discovered_at, source,
			activity_type, activity_name, started_at, duration_secs,
			distance_m, calories, avg_power_w, max_power_w, norm_power_w,
			avg_hr, max_hr, avg_cadence, avg_speed_mps, total_ascent_m,
			device_name, software_version
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		f.Path, f.Hash, f.Size, f.DiscoveredAt, f.Source,
		nullString(f.ActivityType), nullString(f.ActivityName), f.StartedAt, nullInt(f.DurationSecs),
		nullFloat(f.DistanceM), nullInt(f.Calories), nullInt(f.AvgPowerW), nullInt(f.MaxPowerW), nullInt(f.NormPowerW),
		nullInt(f.AvgHR), nullInt(f.MaxHR), nullInt(f.AvgCadence), nullFloat(f.AvgSpeedMPS), nullFloat(f.TotalAscentM),
		nullString(f.DeviceName), nullString(f.SoftwareVersion),
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetFileByPath retrieves a file by its path.
func (s *Store) GetFileByPath(ctx context.Context, path string) (*FitFile, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, path, hash, size, discovered_at, source,
			activity_type, activity_name, started_at, duration_secs,
			distance_m, calories, avg_power_w, max_power_w, norm_power_w,
			avg_hr, max_hr, avg_cadence, avg_speed_mps, total_ascent_m,
			device_name, software_version
		FROM fit_files WHERE path = ?
	`, path)
	return s.scanFile(row)
}

// GetFileByHash retrieves a file by its content hash.
func (s *Store) GetFileByHash(ctx context.Context, hash string) (*FitFile, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, path, hash, size, discovered_at, source,
			activity_type, activity_name, started_at, duration_secs,
			distance_m, calories, avg_power_w, max_power_w, norm_power_w,
			avg_hr, max_hr, avg_cadence, avg_speed_mps, total_ascent_m,
			device_name, software_version
		FROM fit_files WHERE hash = ?
	`, hash)
	return s.scanFile(row)
}

// FileExists checks if a file exists by path or hash.
func (s *Store) FileExists(ctx context.Context, path, hash string) (bool, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM fit_files WHERE path = ? OR hash = ?
	`, path, hash).Scan(&count)
	return count > 0, err
}

// ListFiles returns all files, optionally filtered.
func (s *Store) ListFiles(ctx context.Context, limit int) ([]*FitFile, error) {
	query := `
		SELECT id, path, hash, size, discovered_at, source,
			activity_type, activity_name, started_at, duration_secs,
			distance_m, calories, avg_power_w, max_power_w, norm_power_w,
			avg_hr, max_hr, avg_cadence, avg_speed_mps, total_ascent_m,
			device_name, software_version
		FROM fit_files
		ORDER BY started_at DESC, discovered_at DESC
	`
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*FitFile
	for rows.Next() {
		f, err := s.scanFileRows(rows)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, rows.Err()
}

// CreateSyncRecord creates a pending sync record.
func (s *Store) CreateSyncRecord(ctx context.Context, fileID int64, consumer string) (int64, error) {
	result, err := s.db.ExecContext(ctx, `
		INSERT INTO sync_records (file_id, consumer, status)
		VALUES (?, ?, 'pending')
		ON CONFLICT(file_id, consumer) DO NOTHING
	`, fileID, consumer)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// UpdateSyncSuccess marks a sync as successful.
func (s *Store) UpdateSyncSuccess(ctx context.Context, fileID int64, consumer, remoteID, remoteURL string) error {
	now := time.Now()
	_, err := s.db.ExecContext(ctx, `
		UPDATE sync_records
		SET status = 'success', completed_at = ?, remote_id = ?, remote_url = ?, error = NULL
		WHERE file_id = ? AND consumer = ?
	`, now, remoteID, remoteURL, fileID, consumer)
	return err
}

// UpdateSyncFailed marks a sync as failed.
func (s *Store) UpdateSyncFailed(ctx context.Context, fileID int64, consumer, errMsg string) error {
	now := time.Now()
	_, err := s.db.ExecContext(ctx, `
		UPDATE sync_records
		SET status = 'failed', attempted_at = ?, error = ?, retries = retries + 1
		WHERE file_id = ? AND consumer = ?
	`, now, errMsg, fileID, consumer)
	return err
}

// UpdateSyncAttempted marks that a sync was attempted.
func (s *Store) UpdateSyncAttempted(ctx context.Context, fileID int64, consumer string) error {
	now := time.Now()
	_, err := s.db.ExecContext(ctx, `
		UPDATE sync_records
		SET attempted_at = ?
		WHERE file_id = ? AND consumer = ?
	`, now, fileID, consumer)
	return err
}

// GetPendingFiles returns files that haven't been successfully synced to a consumer.
func (s *Store) GetPendingFiles(ctx context.Context, consumer string) ([]*FitFile, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT f.id, f.path, f.hash, f.size, f.discovered_at, f.source,
			f.activity_type, f.activity_name, f.started_at, f.duration_secs,
			f.distance_m, f.calories, f.avg_power_w, f.max_power_w, f.norm_power_w,
			f.avg_hr, f.max_hr, f.avg_cadence, f.avg_speed_mps, f.total_ascent_m,
			f.device_name, f.software_version
		FROM fit_files f
		JOIN sync_records s ON f.id = s.file_id
		WHERE s.consumer = ? AND s.status = 'pending'
		ORDER BY f.started_at ASC
	`, consumer)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*FitFile
	for rows.Next() {
		f, err := s.scanFileRows(rows)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, rows.Err()
}

// GetFailedSyncs returns failed sync records that can be retried.
func (s *Store) GetFailedSyncs(ctx context.Context, consumer string, maxRetries int) ([]*SyncRecord, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, file_id, consumer, status, attempted_at, completed_at,
			remote_id, remote_url, error, retries
		FROM sync_records
		WHERE consumer = ? AND status = 'failed' AND retries < ?
		ORDER BY attempted_at ASC
	`, consumer, maxRetries)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*SyncRecord
	for rows.Next() {
		r := &SyncRecord{}
		err := rows.Scan(
			&r.ID, &r.FileID, &r.Consumer, &r.Status,
			&r.AttemptedAt, &r.CompletedAt,
			&r.RemoteID, &r.RemoteURL, &r.Error, &r.Retries,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

// ResetToRetry changes failed syncs back to pending for retry.
func (s *Store) ResetToRetry(ctx context.Context, consumer string, maxRetries int) (int64, error) {
	result, err := s.db.ExecContext(ctx, `
		UPDATE sync_records
		SET status = 'pending'
		WHERE consumer = ? AND status = 'failed' AND retries < ?
	`, consumer, maxRetries)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// Stats returns aggregate statistics.
func (s *Store) Stats(ctx context.Context) (*StoreStats, error) {
	stats := &StoreStats{
		PendingByConsumer: make(map[string]int),
		FailedByConsumer:  make(map[string]int),
		SuccessByConsumer: make(map[string]int),
	}

	// Total files
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM fit_files").Scan(&stats.TotalFiles)
	if err != nil {
		return nil, err
	}

	// Total syncs
	err = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sync_records").Scan(&stats.TotalSyncs)
	if err != nil {
		return nil, err
	}

	// Syncs by consumer and status
	rows, err := s.db.QueryContext(ctx, `
		SELECT consumer, status, COUNT(*) 
		FROM sync_records 
		GROUP BY consumer, status
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var consumer, status string
		var count int
		if err := rows.Scan(&consumer, &status, &count); err != nil {
			return nil, err
		}
		switch status {
		case "pending":
			stats.PendingByConsumer[consumer] = count
		case "failed":
			stats.FailedByConsumer[consumer] = count
		case "success":
			stats.SuccessByConsumer[consumer] = count
		}
	}

	return stats, rows.Err()
}

// Helper functions

func (s *Store) scanFile(row *sql.Row) (*FitFile, error) {
	f := &FitFile{}
	var activityType, activityName, deviceName, softwareVersion sql.NullString
	var startedAt sql.NullTime
	var durationSecs, calories, avgPowerW, maxPowerW, normPowerW sql.NullInt64
	var avgHR, maxHR, avgCadence sql.NullInt64
	var distanceM, avgSpeedMPS, totalAscentM sql.NullFloat64

	err := row.Scan(
		&f.ID, &f.Path, &f.Hash, &f.Size, &f.DiscoveredAt, &f.Source,
		&activityType, &activityName, &startedAt, &durationSecs,
		&distanceM, &calories, &avgPowerW, &maxPowerW, &normPowerW,
		&avgHR, &maxHR, &avgCadence, &avgSpeedMPS, &totalAscentM,
		&deviceName, &softwareVersion,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	f.ActivityType = activityType.String
	f.ActivityName = activityName.String
	if startedAt.Valid {
		f.StartedAt = &startedAt.Time
	}
	f.DurationSecs = int(durationSecs.Int64)
	f.DistanceM = distanceM.Float64
	f.Calories = int(calories.Int64)
	f.AvgPowerW = int(avgPowerW.Int64)
	f.MaxPowerW = int(maxPowerW.Int64)
	f.NormPowerW = int(normPowerW.Int64)
	f.AvgHR = int(avgHR.Int64)
	f.MaxHR = int(maxHR.Int64)
	f.AvgCadence = int(avgCadence.Int64)
	f.AvgSpeedMPS = avgSpeedMPS.Float64
	f.TotalAscentM = totalAscentM.Float64
	f.DeviceName = deviceName.String
	f.SoftwareVersion = softwareVersion.String

	return f, nil
}

func (s *Store) scanFileRows(rows *sql.Rows) (*FitFile, error) {
	f := &FitFile{}
	var activityType, activityName, deviceName, softwareVersion sql.NullString
	var startedAt sql.NullTime
	var durationSecs, calories, avgPowerW, maxPowerW, normPowerW sql.NullInt64
	var avgHR, maxHR, avgCadence sql.NullInt64
	var distanceM, avgSpeedMPS, totalAscentM sql.NullFloat64

	err := rows.Scan(
		&f.ID, &f.Path, &f.Hash, &f.Size, &f.DiscoveredAt, &f.Source,
		&activityType, &activityName, &startedAt, &durationSecs,
		&distanceM, &calories, &avgPowerW, &maxPowerW, &normPowerW,
		&avgHR, &maxHR, &avgCadence, &avgSpeedMPS, &totalAscentM,
		&deviceName, &softwareVersion,
	)
	if err != nil {
		return nil, err
	}

	f.ActivityType = activityType.String
	f.ActivityName = activityName.String
	if startedAt.Valid {
		f.StartedAt = &startedAt.Time
	}
	f.DurationSecs = int(durationSecs.Int64)
	f.DistanceM = distanceM.Float64
	f.Calories = int(calories.Int64)
	f.AvgPowerW = int(avgPowerW.Int64)
	f.MaxPowerW = int(maxPowerW.Int64)
	f.NormPowerW = int(normPowerW.Int64)
	f.AvgHR = int(avgHR.Int64)
	f.MaxHR = int(maxHR.Int64)
	f.AvgCadence = int(avgCadence.Int64)
	f.AvgSpeedMPS = avgSpeedMPS.Float64
	f.TotalAscentM = totalAscentM.Float64
	f.DeviceName = deviceName.String
	f.SoftwareVersion = softwareVersion.String

	return f, nil
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func nullInt(i int) sql.NullInt64 {
	if i == 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(i), Valid: true}
}

func nullFloat(f float64) sql.NullFloat64 {
	if f == 0 {
		return sql.NullFloat64{}
	}
	return sql.NullFloat64{Float64: f, Valid: true}
}
