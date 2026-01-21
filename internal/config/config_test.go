package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_Default(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Write minimal config
	if err := os.WriteFile(configPath, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Defaults should be set for missing WatchDirs
	if len(cfg.WatchDirs) == 0 {
		t.Error("expected default watch dirs")
	}
}

func TestLoad_FromFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Write a test config
	configContent := `
watch_dirs = ["/my/fit/folder", "/another/folder"]
store_path = "/custom/store.json"

[intervals]
enabled = true
athlete_id = "i123456"
api_key = "secret-key"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(cfg.WatchDirs) != 2 {
		t.Fatalf("expected 2 watch dirs, got %d", len(cfg.WatchDirs))
	}
	if cfg.WatchDirs[0] != "/my/fit/folder" {
		t.Errorf("expected /my/fit/folder, got %s", cfg.WatchDirs[0])
	}

	if cfg.StorePath != "/custom/store.json" {
		t.Errorf("expected custom store path, got %s", cfg.StorePath)
	}

	if !cfg.Intervals.Enabled {
		t.Error("expected intervals enabled")
	}
	if cfg.Intervals.AthleteID != "i123456" {
		t.Errorf("expected athlete_id i123456, got %s", cfg.Intervals.AthleteID)
	}
	if cfg.Intervals.APIKey != "secret-key" {
		t.Errorf("expected api_key secret-key, got %s", cfg.Intervals.APIKey)
	}
}

func TestLoad_InvalidTOML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Write invalid TOML
	if err := os.WriteFile(configPath, []byte("this is not valid TOML [[["), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Error("expected error for invalid TOML")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/config.toml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if len(cfg.WatchDirs) == 0 {
		t.Error("expected default watch directories")
	}
	if cfg.Intervals.Enabled {
		t.Error("expected intervals disabled by default")
	}
}

func TestDefaultWatchDirs(t *testing.T) {
	dirs := DefaultWatchDirs()
	if len(dirs) == 0 {
		t.Error("expected default directories")
	}
	for _, dir := range dirs {
		if dir == "" {
			t.Error("expected non-empty directory paths")
		}
	}
}

func TestConfig_Save(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "subdir", "config.toml")

	cfg := DefaultConfig()
	cfg.Intervals.AthleteID = "test-athlete"
	cfg.Intervals.APIKey = "test-key"
	cfg.Intervals.Enabled = true

	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file was not created")
	}

	// Load it back and verify
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load after Save failed: %v", err)
	}

	if loaded.Intervals.AthleteID != "test-athlete" {
		t.Errorf("expected test-athlete, got %s", loaded.Intervals.AthleteID)
	}
	if loaded.Intervals.APIKey != "test-key" {
		t.Errorf("expected test-key, got %s", loaded.Intervals.APIKey)
	}
	if !loaded.Intervals.Enabled {
		t.Error("expected intervals enabled")
	}
}

func TestLoadOrCreate_Creates(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// File doesn't exist yet
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Fatal("config file should not exist")
	}

	cfg, err := LoadOrCreate(configPath)
	if err != nil {
		t.Fatalf("LoadOrCreate failed: %v", err)
	}

	// File should now exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file should have been created")
	}

	// Should have defaults
	if len(cfg.WatchDirs) == 0 {
		t.Error("expected default watch dirs")
	}
}

func TestLoadOrCreate_Loads(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Write existing config
	configContent := `
watch_dirs = ["/custom/path"]

[intervals]
athlete_id = "existing-athlete"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadOrCreate(configPath)
	if err != nil {
		t.Fatalf("LoadOrCreate failed: %v", err)
	}

	if cfg.Intervals.AthleteID != "existing-athlete" {
		t.Errorf("expected existing-athlete, got %s", cfg.Intervals.AthleteID)
	}
}

func TestMultipleWatchDirs(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	configContent := `
watch_dirs = ["/dir/one", "/dir/two", "/dir/three"]
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(cfg.WatchDirs) != 3 {
		t.Errorf("expected 3 watch dirs, got %d", len(cfg.WatchDirs))
	}
}

func TestConfig_EmptyIntervals(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Config with no intervals section
	configContent := `
watch_dirs = ["/my/dir"]
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Should have zero-value intervals
	if cfg.Intervals.Enabled {
		t.Error("intervals should be disabled by default")
	}
	if cfg.Intervals.AthleteID != "" {
		t.Error("athlete_id should be empty")
	}
}
