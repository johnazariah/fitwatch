// Package watcher monitors directories for new FIT files.
package watcher

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// Watcher monitors directories for new FIT files.
type Watcher struct {
	dirs    []string
	onNew   func(path string)
	logger  *slog.Logger
	watcher *fsnotify.Watcher

	mu   sync.Mutex
	seen map[string]bool // Track files we've already processed
}

// New creates a new FIT file watcher.
func New(dirs []string, onNew func(path string), logger *slog.Logger) *Watcher {
	if logger == nil {
		logger = slog.Default()
	}
	return &Watcher{
		dirs:   dirs,
		onNew:  onNew,
		logger: logger,
		seen:   make(map[string]bool),
	}
}

// Watch starts watching for new FIT files.
// This blocks until the context is cancelled.
func (w *Watcher) Watch(ctx context.Context) error {
	var err error
	w.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer w.watcher.Close()

	// Add all directories to watch
	for _, dir := range w.dirs {
		if err := w.addDir(dir); err != nil {
			w.logger.Warn("failed to watch directory", "dir", dir, "error", err)
		}
	}

	// Process events
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case event, ok := <-w.watcher.Events:
			if !ok {
				return nil
			}
			w.handleEvent(event)

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return nil
			}
			w.logger.Error("watcher error", "error", err)
		}
	}
}

// ScanExisting scans directories for existing FIT files.
// Use this to process files that were added while the watcher wasn't running.
func (w *Watcher) ScanExisting() error {
	for _, dir := range w.dirs {
		if err := w.scanDir(dir); err != nil {
			w.logger.Warn("failed to scan directory", "dir", dir, "error", err)
		}
	}
	return nil
}

// MarkSeen marks a file as already processed (won't trigger onNew).
func (w *Watcher) MarkSeen(path string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.seen[path] = true
}

// IsSeen checks if a file has already been processed.
func (w *Watcher) IsSeen(path string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.seen[path]
}

func (w *Watcher) addDir(dir string) error {
	// Expand home directory
	if strings.HasPrefix(dir, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		dir = filepath.Join(home, dir[1:])
	}

	// Check directory exists
	info, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return os.ErrNotExist
	}

	w.logger.Info("watching directory", "dir", dir)
	return w.watcher.Add(dir)
}

func (w *Watcher) scanDir(dir string) error {
	// Expand home directory
	if strings.HasPrefix(dir, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		dir = filepath.Join(home, dir[1:])
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if isFitFile(entry.Name()) {
			path := filepath.Join(dir, entry.Name())
			if !w.IsSeen(path) {
				w.logger.Debug("found existing FIT file", "path", path)
				w.MarkSeen(path)
				w.onNew(path)
			}
		}
	}

	return nil
}

func (w *Watcher) handleEvent(event fsnotify.Event) {
	// Only care about create and write events
	if !event.Has(fsnotify.Create) && !event.Has(fsnotify.Write) {
		return
	}

	// Only care about FIT files
	if !isFitFile(event.Name) {
		return
	}

	// Skip if already seen
	if w.IsSeen(event.Name) {
		return
	}

	w.logger.Info("new FIT file detected", "path", event.Name)
	w.MarkSeen(event.Name)
	w.onNew(event.Name)
}

func isFitFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".fit"
}
