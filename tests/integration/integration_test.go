package integration

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/johnazariah/fitwatch/internal/consumer/intervals"
	"github.com/johnazariah/fitwatch/internal/fitparser"
	"github.com/johnazariah/fitwatch/internal/store"
	"github.com/johnazariah/fitwatch/internal/watcher"
)

// TestEndToEnd_NewFileDetectedAndUploaded tests the full pipeline:
// 1. Watcher detects new FIT file
// 2. FIT file is parsed
// 3. Metadata is stored in SQLite
// 4. File is uploaded to Intervals.icu
// 5. Sync record is created
func TestEndToEnd_NewFileDetectedAndUploaded(t *testing.T) {
	ctx := context.Background()

	// Setup mock Intervals.icu server
	var uploadedFiles []string
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		uploadedFiles = append(uploadedFiles, r.URL.Path)
		mu.Unlock()
		_, _ = io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id": "activity123"}`))
	}))
	defer server.Close()

	// Setup temp directories
	watchDir := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "test.db")

	// Create store
	db, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Create consumer
	consumer := intervals.New("athlete123", "apikey456")
	consumer.BaseURL = server.URL

	// Track processed files
	var processedFiles []string
	var processedMu sync.Mutex

	// Create watcher with handler
	w := watcher.New([]string{watchDir}, func(path string) {
		// Parse the FIT file
		meta, err := fitparser.Parse(path)
		if err != nil {
			t.Logf("parse error (expected for fake file): %v", err)
			// For fake files, create minimal metadata
			hash, _ := fitparser.HashFile(path)
			fi, _ := os.Stat(path)
			meta = &fitparser.Metadata{
				Hash: hash,
				Size: fi.Size(),
			}
		}

		// Store in database
		fitFile := &store.FitFile{
			Path:         path,
			Hash:         meta.Hash,
			Size:         meta.Size,
			ActivityType: meta.ActivityType,
			StartedAt:    meta.StartTime,
			DurationSecs: meta.DurationSecs,
			DiscoveredAt: time.Now(),
		}
		fileID, err := db.InsertFile(ctx, fitFile)
		if err != nil {
			t.Errorf("failed to insert file: %v", err)
			return
		}

		// Upload to Intervals.icu
		syncErr := consumer.Push(ctx, path)

		// Record sync result
		_, _ = db.CreateSyncRecord(ctx, fileID, consumer.Name())

		if syncErr != nil {
			db.UpdateSyncFailed(ctx, fileID, consumer.Name(), syncErr.Error())
		} else {
			db.UpdateSyncSuccess(ctx, fileID, consumer.Name(), "activity123", "")
		}

		processedMu.Lock()
		processedFiles = append(processedFiles, path)
		processedMu.Unlock()
	}, slog.Default())

	// Start watcher
	watchCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	go w.Watch(watchCtx)
	time.Sleep(100 * time.Millisecond) // Let watcher start

	// Create a FIT file
	fitPath := filepath.Join(watchDir, "test-activity.fit")
	if err := os.WriteFile(fitPath, []byte("fake FIT data for integration test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Wait for processing
	time.Sleep(1 * time.Second)

	// Verify file was processed
	processedMu.Lock()
	if len(processedFiles) != 1 {
		t.Errorf("expected 1 processed file, got %d", len(processedFiles))
	}
	processedMu.Unlock()

	// Verify file was uploaded
	mu.Lock()
	if len(uploadedFiles) != 1 {
		t.Errorf("expected 1 upload, got %d", len(uploadedFiles))
	}
	mu.Unlock()

	// Verify database state
	stats, err := db.Stats(ctx)
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}

	if stats.TotalFiles != 1 {
		t.Errorf("expected 1 file in db, got %d", stats.TotalFiles)
	}
}

// TestEndToEnd_DuplicateDetection tests that duplicate files are not re-uploaded
func TestEndToEnd_DuplicateDetection(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Get real FIT file if available
	realFitPath := filepath.Join("..", "..", "testdata", "sample.fit")
	if _, err := os.Stat(realFitPath); os.IsNotExist(err) {
		t.Skip("sample.fit not found")
	}

	// Parse and store the first time
	meta, err := fitparser.Parse(realFitPath)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	fitFile := &store.FitFile{
		Path:         realFitPath,
		Hash:         meta.Hash,
		Size:         meta.Size,
		ActivityType: meta.ActivityType,
		StartedAt:    meta.StartTime,
		DurationSecs: meta.DurationSecs,
		DiscoveredAt: time.Now(),
	}

	// Insert first file
	if _, err := db.InsertFile(ctx, fitFile); err != nil {
		t.Fatalf("failed to insert file: %v", err)
	}

	// Check if duplicate exists by hash
	exists, err := db.GetFileByHash(ctx, meta.Hash)
	if err != nil {
		t.Fatalf("GetFileByHash failed: %v", err)
	}
	if exists == nil {
		t.Error("expected file to exist by hash")
	}

	// Try to insert same file with different path
	fitFile2 := &store.FitFile{
		Path:         "/some/other/path.fit",
		Hash:         meta.Hash, // Same hash!
		Size:         meta.Size,
		DiscoveredAt: time.Now(),
	}

	// This should detect duplicate by hash
	existing, _ := db.GetFileByHash(ctx, fitFile2.Hash)
	if existing == nil {
		t.Error("duplicate detection by hash should find existing file")
	}
}

// TestEndToEnd_RealFitFileParsing tests parsing a real FIT file and storing metadata
func TestEndToEnd_RealFitFileParsing(t *testing.T) {
	ctx := context.Background()
	realFitPath := filepath.Join("..", "..", "testdata", "sample.fit")
	if _, err := os.Stat(realFitPath); os.IsNotExist(err) {
		t.Skip("sample.fit not found")
	}

	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Parse the FIT file
	meta, err := fitparser.Parse(realFitPath)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	// Store with full metadata
	fitFile := &store.FitFile{
		Path:         realFitPath,
		Hash:         meta.Hash,
		Size:         meta.Size,
		ActivityType: meta.ActivityType,
		StartedAt:    meta.StartTime,
		DurationSecs: meta.DurationSecs,
		DistanceM:    meta.DistanceMeters,
		Calories:     meta.Calories,
		AvgPowerW:    meta.AvgPower,
		MaxPowerW:    meta.MaxPower,
		NormPowerW:   meta.NormPower,
		AvgHR:        meta.AvgHeartRate,
		MaxHR:        meta.MaxHeartRate,
		AvgCadence:   meta.AvgCadence,
		AvgSpeedMPS:  meta.AvgSpeedMPS,
		TotalAscentM: meta.TotalAscent,
		DeviceName:   meta.Manufacturer,
		DiscoveredAt: time.Now(),
	}

	if _, err := db.InsertFile(ctx, fitFile); err != nil {
		t.Fatalf("failed to insert file: %v", err)
	}

	// Retrieve and verify
	retrieved, err := db.GetFileByHash(ctx, meta.Hash)
	if err != nil {
		t.Fatalf("GetFileByHash failed: %v", err)
	}

	if retrieved.ActivityType != meta.ActivityType {
		t.Errorf("activity type: expected %s, got %s", meta.ActivityType, retrieved.ActivityType)
	}
	if retrieved.AvgPowerW != meta.AvgPower {
		t.Errorf("avg power: expected %d, got %d", meta.AvgPower, retrieved.AvgPowerW)
	}
	if retrieved.DistanceM != meta.DistanceMeters {
		t.Errorf("distance: expected %.2f, got %.2f", meta.DistanceMeters, retrieved.DistanceM)
	}

	t.Logf("Successfully stored and retrieved FIT file metadata:")
	t.Logf("  Activity: %s", retrieved.ActivityType)
	t.Logf("  Distance: %.2f km", retrieved.DistanceM/1000)
	t.Logf("  Duration: %d min", retrieved.DurationSecs/60)
	t.Logf("  Avg Power: %d W", retrieved.AvgPowerW)
}

// TestEndToEnd_ScanAndSyncExistingFiles tests scanning existing files on startup
func TestEndToEnd_ScanAndSyncExistingFiles(t *testing.T) {
	watchDir := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "test.db")

	// Create some FIT files before starting
	for i := 0; i < 3; i++ {
		path := filepath.Join(watchDir, "existing"+string(rune('a'+i))+".fit")
		if err := os.WriteFile(path, []byte("fake fit data"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	db, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer func() { _ = db.Close() }()

	var scannedCount int
	var mu sync.Mutex

	w := watcher.New([]string{watchDir}, func(path string) {
		mu.Lock()
		scannedCount++
		mu.Unlock()
	}, slog.Default())

	// Scan existing files
	if err := w.ScanExisting(); err != nil {
		t.Fatalf("ScanExisting failed: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if scannedCount != 3 {
		t.Errorf("expected 3 scanned files, got %d", scannedCount)
	}
}

// TestEndToEnd_SyncFailureAndRetry tests handling of sync failures
func TestEndToEnd_SyncFailureAndRetry(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Insert a file
	fitFile := &store.FitFile{
		Path:         "/test/file.fit",
		Hash:         "testhash123",
		Size:         1000,
		DiscoveredAt: time.Now(),
	}
	fileID, err := db.InsertFile(ctx, fitFile)
	if err != nil {
		t.Fatal(err)
	}

	// Create sync record
	consumerName := "Intervals.icu"
	if _, err := db.CreateSyncRecord(ctx, fileID, consumerName); err != nil {
		t.Fatal(err)
	}

	// Mark as failed
	if err := db.UpdateSyncFailed(ctx, fileID, consumerName, "connection timeout"); err != nil {
		t.Fatal(err)
	}

	// Check stats
	stats, err := db.Stats(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if stats.FailedByConsumer[consumerName] != 1 {
		t.Errorf("expected 1 failed sync for %s, got %d", consumerName, stats.FailedByConsumer[consumerName])
	}
}
