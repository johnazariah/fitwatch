// Package config handles application configuration.
package config

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"

	"github.com/pelletier/go-toml/v2"
)

// Config represents the application configuration.
type Config struct {
	// WatchDirs are directories to monitor for new FIT files.
	WatchDirs []string `toml:"watch_dirs"`

	// Intervals.icu configuration
	Intervals IntervalsConfig `toml:"intervals"`

	// Store path for sync database (optional, defaults to ~/.fitwatch/fitwatch.db)
	StorePath string `toml:"store_path,omitempty"`
}

// IntervalsConfig holds Intervals.icu API settings.
type IntervalsConfig struct {
	Enabled   bool   `toml:"enabled"`
	AthleteID string `toml:"athlete_id"`
	APIKey    string `toml:"api_key"`
}

// DefaultWatchDirs returns platform-specific default directories.
func DefaultWatchDirs() []string {
	home, _ := os.UserHomeDir()

	switch runtime.GOOS {
	case "windows":
		return []string{
			filepath.Join(home, "Documents", "Zwift", "Activities"),
			filepath.Join(home, "Documents", "TrainerRoad"),
		}
	case "darwin":
		return []string{
			filepath.Join(home, "Documents", "Zwift", "Activities"),
			filepath.Join(home, "Documents", "TrainerRoad"),
		}
	default: // linux
		return []string{
			filepath.Join(home, "Documents", "Zwift", "Activities"),
			filepath.Join(home, ".local", "share", "Zwift", "Activities"),
		}
	}
}

// DefaultConfig returns a config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		WatchDirs: DefaultWatchDirs(),
		Intervals: IntervalsConfig{
			Enabled: false,
		},
	}
}

// Load reads configuration from a file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Apply defaults for missing values
	if len(cfg.WatchDirs) == 0 {
		cfg.WatchDirs = DefaultWatchDirs()
	}

	return &cfg, nil
}

// LoadOrCreate loads config from path, or creates a default one if it doesn't exist.
func LoadOrCreate(path string) (*Config, error) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		cfg := DefaultConfig()
		if err := cfg.Save(path); err != nil {
			return nil, err
		}
		return cfg, nil
	}
	return Load(path)
}

// Save writes configuration to a file.
func (c *Config) Save(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := toml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// Validate checks the configuration is valid.
func (c *Config) Validate() error {
	if c.Intervals.Enabled {
		if c.Intervals.AthleteID == "" {
			return errors.New("intervals.athlete_id is required when intervals is enabled")
		}
		if c.Intervals.APIKey == "" {
			return errors.New("intervals.api_key is required when intervals is enabled")
		}
	}
	return nil
}

// DefaultConfigPath returns the default config file location.
func DefaultConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".fitwatch", "config.toml")
}

// DefaultStorePath returns the default sync database location.
func DefaultStorePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".fitwatch", "fitwatch.db")
}
