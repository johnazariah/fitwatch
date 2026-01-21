package watcher

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestWatcher_Creation(t *testing.T) {
	tmpDir := t.TempDir()

	w := New([]string{tmpDir}, func(path string) {}, slog.Default())
	if w == nil {
		t.Fatal("New returned nil")
	}
}

func TestWatcher_DetectsNewFitFile(t *testing.T) {
	tmpDir := t.TempDir()

	var mu sync.Mutex
	var received []string

	w := New([]string{tmpDir}, func(path string) {
		mu.Lock()
		received = append(received, path)
		mu.Unlock()
	}, slog.Default())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() { _ = w.Watch(ctx) }()

	// Give watcher time to start
	time.Sleep(100 * time.Millisecond)

	// Create a FIT file
	fitPath := filepath.Join(tmpDir, "test.fit")
	if err := os.WriteFile(fitPath, []byte("fake fit data"), 0644); err != nil {
		t.Fatal(err)
	}

	// Wait for event
	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(received) == 0 {
		t.Error("expected to receive file event")
		return
	}
	if received[0] != fitPath {
		t.Errorf("expected %s, got %s", fitPath, received[0])
	}
}

func TestWatcher_IgnoresNonFitFiles(t *testing.T) {
	tmpDir := t.TempDir()

	var mu sync.Mutex
	var received []string

	w := New([]string{tmpDir}, func(path string) {
		mu.Lock()
		received = append(received, path)
		mu.Unlock()
	}, slog.Default())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go func() { _ = w.Watch(ctx) }()

	// Give watcher time to start
	time.Sleep(100 * time.Millisecond)

	// Create non-FIT files
	for _, ext := range []string{".txt", ".csv", ".gpx", ".tcx", ".json"} {
		path := filepath.Join(tmpDir, "test"+ext)
		if err := os.WriteFile(path, []byte("test data"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Give time for events
	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(received) > 0 {
		t.Errorf("should not have received events for non-FIT files: %v", received)
	}
}

func TestWatcher_MultipleFitFiles(t *testing.T) {
	tmpDir := t.TempDir()

	var mu sync.Mutex
	var received []string

	w := New([]string{tmpDir}, func(path string) {
		mu.Lock()
		received = append(received, path)
		mu.Unlock()
	}, slog.Default())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() { _ = w.Watch(ctx) }()

	// Give watcher time to start
	time.Sleep(100 * time.Millisecond)

	// Create multiple FIT files
	expected := make(map[string]bool)
	for i := 0; i < 3; i++ {
		fitPath := filepath.Join(tmpDir, "test"+string(rune('a'+i))+".fit")
		expected[fitPath] = true
		if err := os.WriteFile(fitPath, []byte("fake fit data"), 0644); err != nil {
			t.Fatal(err)
		}
		time.Sleep(50 * time.Millisecond) // Small delay between files
	}

	// Wait for events
	time.Sleep(1 * time.Second)

	mu.Lock()
	defer mu.Unlock()

	receivedMap := make(map[string]bool)
	for _, p := range received {
		receivedMap[p] = true
	}

	for path := range expected {
		if !receivedMap[path] {
			t.Errorf("missing event for %s", path)
		}
	}
}

func TestWatcher_MultipleDirectories(t *testing.T) {
	tmpDir1 := t.TempDir()
	tmpDir2 := t.TempDir()

	var mu sync.Mutex
	var received []string

	w := New([]string{tmpDir1, tmpDir2}, func(path string) {
		mu.Lock()
		received = append(received, path)
		mu.Unlock()
	}, slog.Default())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() { _ = w.Watch(ctx) }()

	// Give watcher time to start
	time.Sleep(100 * time.Millisecond)

	// Create FIT files in both directories
	fitPath1 := filepath.Join(tmpDir1, "test1.fit")
	fitPath2 := filepath.Join(tmpDir2, "test2.fit")

	if err := os.WriteFile(fitPath1, []byte("fake fit data"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fitPath2, []byte("fake fit data"), 0644); err != nil {
		t.Fatal(err)
	}

	// Wait for events
	time.Sleep(1 * time.Second)

	mu.Lock()
	defer mu.Unlock()

	receivedMap := make(map[string]bool)
	for _, p := range received {
		receivedMap[p] = true
	}

	if !receivedMap[fitPath1] {
		t.Errorf("missing event for %s", fitPath1)
	}
	if !receivedMap[fitPath2] {
		t.Errorf("missing event for %s", fitPath2)
	}
}

func TestWatcher_CaseInsensitiveFitExtension(t *testing.T) {
	tmpDir := t.TempDir()

	var mu sync.Mutex
	var received []string

	w := New([]string{tmpDir}, func(path string) {
		mu.Lock()
		received = append(received, path)
		mu.Unlock()
	}, slog.Default())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() { _ = w.Watch(ctx) }()

	// Give watcher time to start
	time.Sleep(100 * time.Millisecond)

	// Create FIT files with different case extensions
	extensions := []string{".fit", ".FIT", ".Fit", ".FiT"}
	expected := make(map[string]bool)
	for i, ext := range extensions {
		path := filepath.Join(tmpDir, "test"+string(rune('a'+i))+ext)
		expected[path] = true
		if err := os.WriteFile(path, []byte("fake fit data"), 0644); err != nil {
			t.Fatal(err)
		}
		time.Sleep(50 * time.Millisecond)
	}

	// Wait for events
	time.Sleep(1 * time.Second)

	mu.Lock()
	defer mu.Unlock()

	receivedMap := make(map[string]bool)
	for _, p := range received {
		receivedMap[p] = true
	}

	for path := range expected {
		if !receivedMap[path] {
			t.Errorf("missing event for %s", path)
		}
	}
}

func TestWatcher_ContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()

	w := New([]string{tmpDir}, func(path string) {}, slog.Default())

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = w.Watch(ctx)
	}()

	// Give watcher time to start
	time.Sleep(100 * time.Millisecond)

	// Cancel context
	cancel()

	// Wait for watcher to exit
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Good - watcher exited
	case <-time.After(2 * time.Second):
		t.Error("watcher did not exit after context cancellation")
	}
}

func TestWatcher_ScanExisting(t *testing.T) {
	tmpDir := t.TempDir()

	// Create FIT files BEFORE starting watcher
	fitPath1 := filepath.Join(tmpDir, "existing1.fit")
	fitPath2 := filepath.Join(tmpDir, "existing2.fit")
	nonFitPath := filepath.Join(tmpDir, "existing.txt")

	if err := os.WriteFile(fitPath1, []byte("fake fit data"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fitPath2, []byte("fake fit data"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(nonFitPath, []byte("text data"), 0644); err != nil {
		t.Fatal(err)
	}

	var mu sync.Mutex
	var received []string

	w := New([]string{tmpDir}, func(path string) {
		mu.Lock()
		received = append(received, path)
		mu.Unlock()
	}, slog.Default())

	// Scan existing files
	err := w.ScanExisting()
	if err != nil {
		t.Fatalf("ScanExisting failed: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if len(received) != 2 {
		t.Errorf("expected 2 files scanned, got %d", len(received))
	}

	receivedMap := make(map[string]bool)
	for _, p := range received {
		receivedMap[p] = true
	}

	if !receivedMap[fitPath1] {
		t.Errorf("missing callback for %s", fitPath1)
	}
	if !receivedMap[fitPath2] {
		t.Errorf("missing callback for %s", fitPath2)
	}
	if receivedMap[nonFitPath] {
		t.Error("should not have called back for non-FIT file")
	}
}
