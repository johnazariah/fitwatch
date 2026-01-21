# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.1.0] - 2026-01-22

### Added
- File system watcher for automatic FIT file detection
- Intervals.icu integration for automatic activity upload
- Activity name extraction from filename (e.g., `2025-02-23_Hudayriyat_Ascend.fit` → "Hudayriyat Ascend")
- Retry with exponential backoff for failed uploads (3 retries, 1s→2s→4s→8s)
- Wait for file to be fully written before processing (fixes Windows copy issues)
- SQLite-based sync tracking database
- Service daemon support for Windows, Linux, and macOS
- CLI with interactive and one-shot modes
- Configuration via TOML file (`~/.fitwatch/config.toml`)

### Platforms
- Windows (amd64)
- macOS (amd64, arm64)
- Linux (amd64, arm64)

[v0.1.0]: https://github.com/johnazariah/fitwatch/releases/tag/v0.1.0
