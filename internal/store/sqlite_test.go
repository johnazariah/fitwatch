package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStore_InsertAndGetFile(t *testing.T) {
	// Create temp database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer func() { _ = store.Close() }()

	ctx := context.Background()

	// Insert a file
	now := time.Now()
	file := &FitFile{
		Path:         "/path/to/test.fit",
		Hash:         "abc123",
		Size:         1024,
		DiscoveredAt: now,
		Source:       "test",
		ActivityType: "cycling",
		ActivityName: "Morning Ride",
		StartedAt:    &now,
		DurationSecs: 3600,
		DistanceM:    50000,
		AvgPowerW:    200,
		MaxPowerW:    400,
		AvgHR:        140,
		MaxHR:        175,
	}

	id, err := store.InsertFile(ctx, file)
	if err != nil {
		t.Fatalf("failed to insert file: %v", err)
	}
	if id == 0 {
		t.Error("expected non-zero ID")
	}

	// Get by path
	got, err := store.GetFileByPath(ctx, file.Path)
	if err != nil {
		t.Fatalf("failed to get file: %v", err)
	}
	if got == nil {
		t.Fatal("expected file, got nil")
	}
	if got.Hash != file.Hash {
		t.Errorf("hash mismatch: got %s, want %s", got.Hash, file.Hash)
	}
	if got.ActivityType != file.ActivityType {
		t.Errorf("activity type mismatch: got %s, want %s", got.ActivityType, file.ActivityType)
	}
	if got.AvgPowerW != file.AvgPowerW {
		t.Errorf("avg power mismatch: got %d, want %d", got.AvgPowerW, file.AvgPowerW)
	}

	// Get by hash
	got2, err := store.GetFileByHash(ctx, file.Hash)
	if err != nil {
		t.Fatalf("failed to get file by hash: %v", err)
	}
	if got2 == nil {
		t.Fatal("expected file, got nil")
	}
	if got2.ID != got.ID {
		t.Errorf("ID mismatch: got %d, want %d", got2.ID, got.ID)
	}
}

func TestStore_FileExists(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer func() { _ = store.Close() }()

	ctx := context.Background()

	// Check non-existent file
	exists, err := store.FileExists(ctx, "/nonexistent", "nohash")
	if err != nil {
		t.Fatalf("failed to check existence: %v", err)
	}
	if exists {
		t.Error("expected file to not exist")
	}

	// Insert a file
	file := &FitFile{
		Path:         "/path/to/test.fit",
		Hash:         "abc123",
		Size:         1024,
		DiscoveredAt: time.Now(),
		Source:       "test",
	}
	_, err = store.InsertFile(ctx, file)
	if err != nil {
		t.Fatalf("failed to insert file: %v", err)
	}

	// Check by path
	exists, err = store.FileExists(ctx, file.Path, "differenthash")
	if err != nil {
		t.Fatalf("failed to check existence: %v", err)
	}
	if !exists {
		t.Error("expected file to exist by path")
	}

	// Check by hash
	exists, err = store.FileExists(ctx, "/different/path", file.Hash)
	if err != nil {
		t.Fatalf("failed to check existence: %v", err)
	}
	if !exists {
		t.Error("expected file to exist by hash")
	}
}

func TestStore_SyncRecords(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer func() { _ = store.Close() }()

	ctx := context.Background()

	// Insert a file
	file := &FitFile{
		Path:         "/path/to/test.fit",
		Hash:         "abc123",
		Size:         1024,
		DiscoveredAt: time.Now(),
		Source:       "test",
	}
	fileID, err := store.InsertFile(ctx, file)
	if err != nil {
		t.Fatalf("failed to insert file: %v", err)
	}

	// Create sync record
	_, err = store.CreateSyncRecord(ctx, fileID, "intervals.icu")
	if err != nil {
		t.Fatalf("failed to create sync record: %v", err)
	}

	// Check pending
	pending, err := store.GetPendingFiles(ctx, "intervals.icu")
	if err != nil {
		t.Fatalf("failed to get pending: %v", err)
	}
	if len(pending) != 1 {
		t.Errorf("expected 1 pending, got %d", len(pending))
	}

	// Mark success
	err = store.UpdateSyncSuccess(ctx, fileID, "intervals.icu", "activity-123", "https://intervals.icu/activity/123")
	if err != nil {
		t.Fatalf("failed to update success: %v", err)
	}

	// Check no longer pending
	pending, err = store.GetPendingFiles(ctx, "intervals.icu")
	if err != nil {
		t.Fatalf("failed to get pending: %v", err)
	}
	if len(pending) != 0 {
		t.Errorf("expected 0 pending, got %d", len(pending))
	}
}

func TestStore_Stats(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer func() { _ = store.Close() }()

	ctx := context.Background()

	// Empty stats
	stats, err := store.Stats(ctx)
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}
	if stats.TotalFiles != 0 {
		t.Errorf("expected 0 files, got %d", stats.TotalFiles)
	}

	// Add some files and syncs
	for i := 0; i < 3; i++ {
		file := &FitFile{
			Path:         filepath.Join("/path", string(rune('a'+i))+".fit"),
			Hash:         string(rune('a' + i)),
			DiscoveredAt: time.Now(),
			Source:       "test",
		}
		id, _ := store.InsertFile(ctx, file)
		_, _ = store.CreateSyncRecord(ctx, id, "intervals.icu")
	}

	stats, err = store.Stats(ctx)
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}
	if stats.TotalFiles != 3 {
		t.Errorf("expected 3 files, got %d", stats.TotalFiles)
	}
	if stats.PendingByConsumer["intervals.icu"] != 3 {
		t.Errorf("expected 3 pending, got %d", stats.PendingByConsumer["intervals.icu"])
	}
}

func TestStore_PersistsAcrossReopen(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// First session
	store1, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	ctx := context.Background()
	file := &FitFile{
		Path:         "/path/to/test.fit",
		Hash:         "abc123",
		DiscoveredAt: time.Now(),
		Source:       "test",
	}
	_, err = store1.InsertFile(ctx, file)
	if err != nil {
		t.Fatalf("failed to insert file: %v", err)
	}
	_ = store1.Close()

	// Verify file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("database file should exist")
	}

	// Second session
	store2, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to reopen store: %v", err)
	}
	defer func() { _ = store2.Close() }()

	got, err := store2.GetFileByPath(ctx, file.Path)
	if err != nil {
		t.Fatalf("failed to get file: %v", err)
	}
	if got == nil {
		t.Fatal("file should persist across sessions")
	}
	if got.Hash != file.Hash {
		t.Errorf("hash mismatch: got %s, want %s", got.Hash, file.Hash)
	}
}
