# FitWatch Architecture Design

## Problem Statement

We need a local tool that:
1. Watches folders for new FIT files
2. Pushes them to configured destinations (Intervals.icu, etc.)
3. Tracks what's been synced to avoid duplicates
4. Handles failures gracefully (retry, report)

## Design Principles

1. **Right-sized generalization** - Pipeline abstraction, but not over-engineered
2. **Reliable** - Never lose a file, never duplicate an upload
3. **Observable** - Know what happened and when
4. **Simple ops** - Single binary, SQLite database, no external deps

---

## Core Abstractions

### 1. Source (Future - P2)

Where FIT files come from. For now, just folder watching.

```go
type Source interface {
    // Stream emits FIT file paths as they become available
    Stream(ctx context.Context) <-chan string
}
```

Future sources:
- `FolderWatcher` - watches local directories
- `CloudSource` - polls iGPSport/MyWhoosh APIs (uses tokens)
- `USBDevice` - detects Garmin/Wahoo when plugged in

### 2. Consumer (P0)

Where FIT files go.

```go
type Consumer interface {
    Name() string
    Push(ctx context.Context, fitPath string) (*PushResult, error)
    Validate() error
}

type PushResult struct {
    RemoteID    string    // ID assigned by destination (e.g., activity ID)
    URL         string    // Link to view the activity (optional)
}
```

Consumers:
- `IntervalsConsumer` - uploads to Intervals.icu (P0)
- `TrainingPeaksConsumer` - uploads to TrainingPeaks (P2)
- `StravaConsumer` - uploads to Strava (P3)
- `WebhookConsumer` - POSTs to custom URL (P2)
- `CopyConsumer` - copies to backup folder (P2)

### 3. Store (P0)

Tracks files and sync status. **SQLite** for reliability.

```go
type Store interface {
    // File tracking
    RecordFile(file *FitFile) error
    GetFile(path string) (*FitFile, error)
    ListPending(consumer string) ([]*FitFile, error)
    
    // Sync tracking
    RecordSync(sync *SyncRecord) error
    GetSyncs(filePath string) ([]*SyncRecord, error)
    
    // Stats
    Stats() (*StoreStats, error)
}
```

---

## Data Model

### FitFile

Represents a discovered FIT file with parsed metadata.

| Field | Type | Description |
|-------|------|-------------|
| `id` | INTEGER | Primary key |
| `path` | TEXT | Absolute file path (unique) |
| `hash` | TEXT | SHA256 of file content (for dedup) |
| `size` | INTEGER | File size in bytes |
| `discovered_at` | TIMESTAMP | When we first saw the file |
| `source` | TEXT | How we found it: "watch", "scan", "api" |
| **Parsed from FIT:** | | |
| `activity_type` | TEXT | "cycling", "running", "virtual_cycling", etc. |
| `activity_name` | TEXT | Activity name/title if present |
| `started_at` | TIMESTAMP | Activity start time |
| `duration_secs` | INTEGER | Total elapsed time |
| `distance_m` | REAL | Distance in meters |
| `calories` | INTEGER | Calories burned |
| `avg_power_w` | INTEGER | Average power (watts) |
| `max_power_w` | INTEGER | Max power (watts) |
| `norm_power_w` | INTEGER | Normalized power |
| `avg_hr` | INTEGER | Average heart rate |
| `max_hr` | INTEGER | Max heart rate |
| `avg_cadence` | INTEGER | Average cadence |
| `avg_speed_mps` | REAL | Average speed (m/s) |
| `total_ascent_m` | REAL | Total elevation gain |
| `device_name` | TEXT | Recording device |
| `software_version` | TEXT | Device firmware/software |

### SyncRecord

Represents an attempt to sync a file to a consumer.

| Field | Type | Description |
|-------|------|-------------|
| `id` | INTEGER | Primary key |
| `file_id` | INTEGER | FK to FitFile |
| `consumer` | TEXT | Consumer name (e.g., "intervals.icu") |
| `status` | TEXT | "pending", "success", "failed" |
| `attempted_at` | TIMESTAMP | When sync was attempted |
| `completed_at` | TIMESTAMP | When sync finished (null if pending) |
| `remote_id` | TEXT | ID from destination (on success) |
| `remote_url` | TEXT | URL to view (on success) |
| `error` | TEXT | Error message (on failure) |
| `retries` | INTEGER | Number of retry attempts |

### ConsumerConfig

Stores consumer credentials (encrypted).

| Field | Type | Description |
|-------|------|-------------|
| `name` | TEXT | Consumer name (primary key) |
| `enabled` | BOOLEAN | Is this consumer active? |
| `config_json` | TEXT | Encrypted JSON blob of credentials |
| `last_sync` | TIMESTAMP | Last successful sync time |

---

## SQLite Schema

```sql
CREATE TABLE fit_files (
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

CREATE TABLE sync_records (
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

CREATE TABLE consumers (
    name TEXT PRIMARY KEY,
    enabled BOOLEAN DEFAULT 0,
    config_json TEXT,
    last_sync TIMESTAMP
);

-- Indexes for common queries
CREATE INDEX idx_sync_pending ON sync_records(consumer, status) WHERE status = 'pending';
CREATE INDEX idx_sync_failed ON sync_records(status) WHERE status = 'failed';
CREATE INDEX idx_files_hash ON fit_files(hash);
CREATE INDEX idx_files_started ON fit_files(started_at);
CREATE INDEX idx_files_type ON fit_files(activity_type);
```

---

## Why SQLite Over JSON?

| Aspect | JSON File | SQLite |
|--------|-----------|--------|
| **Queries** | Load all, filter in memory | Indexed queries |
| **Concurrent access** | Risk of corruption | ACID transactions |
| **Partial failure** | May lose whole file | Atomic writes |
| **Size** | Loads entire file | Query only what's needed |
| **Tooling** | Custom scripts | `sqlite3` CLI, many viewers |
| **Backup** | Copy file | Copy file (same) |
| **Complexity** | Simpler initial | Slightly more setup |

**Verdict**: SQLite. We're tracking potentially thousands of files with multiple sync states. The reliability and queryability justify the small complexity increase.

---

## Processing Flow

```
┌─────────────────────────────────────────────────────────────────────┐
│                           FitWatch                                  │
│                                                                     │
│  ┌──────────────┐                                                   │
│  │   Sources    │                                                   │
│  │              │                                                   │
│  │ FolderWatch ─┼──┐                                                │
│  │ (future:    │  │                                                 │
│  │  CloudSync) │  │                                                 │
│  └──────────────┘  │                                                │
│                    ▼                                                │
│              ┌─────────────┐    ┌─────────────────────────────────┐ │
│              │  Ingester   │───▶│           SQLite DB             │ │
│              │             │    │                                 │ │
│              │ • Hash file │    │ fit_files: discovered files     │ │
│              │ • Dedup     │    │ sync_records: upload attempts   │ │
│              │ • Store     │    │ consumers: config & state       │ │
│              └─────────────┘    └─────────────────────────────────┘ │
│                                           │                         │
│                                           ▼                         │
│                                    ┌─────────────┐                  │
│                                    │  Dispatcher │                  │
│                                    │             │                  │
│                                    │ For each    │                  │
│                                    │ enabled     │                  │
│                                    │ consumer:   │                  │
│                                    └──────┬──────┘                  │
│                                           │                         │
│                    ┌──────────────────────┼──────────────────────┐  │
│                    ▼                      ▼                      ▼  │
│              ┌───────────┐          ┌───────────┐          ┌───────┐│
│              │ Intervals │          │ TrainingPk│          │ ...   ││
│              │    .icu   │          │   (P2)    │          │       ││
│              └─────┬─────┘          └───────────┘          └───────┘│
│                    │                                                │
│                    ▼                                                │
│              ┌───────────┐                                          │
│              │  Update   │                                          │
│              │  sync_    │                                          │
│              │  records  │                                          │
│              └───────────┘                                          │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Retry Strategy

For failed uploads:

1. **Immediate retry**: 3 attempts with exponential backoff (1s, 5s, 30s)
2. **Mark failed**: After 3 attempts, status = "failed"
3. **Background retry**: Every hour, retry failed uploads (max 5 total attempts)
4. **Give up**: After 5 attempts, leave as "failed" (manual intervention needed)

```go
type RetryPolicy struct {
    MaxImmediateRetries int           // 3
    ImmediateBackoff    []time.Duration // [1s, 5s, 30s]
    BackgroundInterval  time.Duration // 1 hour
    MaxTotalRetries     int           // 5
}
```

---

## CLI Commands

```bash
# Watch mode (default)
fitwatch

# One-time sync
fitwatch --once

# Status/stats (TODO)
fitwatch status
fitwatch status --pending
fitwatch status --failed

# Retry failed (TODO)
fitwatch retry

# Manual upload (TODO)
fitwatch upload /path/to/file.fit

# Service management (cross-platform)
fitwatch service install     # Install as system service
fitwatch service uninstall   # Remove system service
fitwatch service start       # Start the service
fitwatch service stop        # Stop the service
fitwatch service restart     # Restart the service
fitwatch service status      # Show service status

# Database inspection (TODO)
fitwatch db stats
fitwatch db files --limit 10
fitwatch db syncs --consumer intervals
```

## Service Support

Uses `github.com/kardianos/service` for cross-platform service installation:

| Platform | Service Type | Install Location |
|----------|--------------|------------------|
| Windows | Windows Service | Services.msc |
| macOS | launchd | ~/Library/LaunchAgents/ |
| Linux | systemd | /etc/systemd/system/ |

**Benefits of service mode:**
- Starts automatically on boot
- Runs in background without terminal
- Proper shutdown handling
- System logging integration

---

## Configuration

Still use TOML for config (not in DB - version controllable):

```toml
# ~/.fitwatch/config.toml

# Directories to watch
watch_dirs = [
    "~/Documents/Zwift/Activities",
    "~/Documents/TrainerRoad"
]

# Database location (default: ~/.fitwatch/fitwatch.db)
db_path = "~/.fitwatch/fitwatch.db"

# Retry settings
[retry]
max_immediate = 3
background_interval = "1h"
max_total = 5

# Intervals.icu consumer
[consumers.intervals]
enabled = true
athlete_id = "i12345"
api_key = "secret"  # Consider: env var or keyring instead

# Future: more consumers
[consumers.trainingpeaks]
enabled = false
```

---

## File Deduplication

Two levels:

1. **Path-based**: Same path = same file (within a session)
2. **Hash-based**: Same SHA256 = same content (across paths/renames)

If hash already exists in DB:
- Log "duplicate content, skipping"
- Don't create new `fit_file` record
- Link to existing record for sync

---

## Observability

### Logging

Structured logging (slog) with levels:
- `DEBUG`: File events, hash calculations
- `INFO`: Syncs started/completed, new files discovered
- `WARN`: Retries, config issues
- `ERROR`: Upload failures, DB errors

### Metrics (Future)

For dashboard/monitoring:
- Files discovered (total, today)
- Syncs by consumer (success, failed, pending)
- Last sync time per consumer
- Average upload time

---

## Project Structure

```
fitwatch/
├── cmd/
│   └── fitwatch/
│       └── main.go           # CLI entrypoint
├── internal/
│   ├── config/
│   │   └── config.go         # TOML config loading
│   ├── store/
│   │   ├── store.go          # Store interface
│   │   ├── sqlite.go         # SQLite implementation
│   │   └── models.go         # FitFile, SyncRecord types
│   ├── watcher/
│   │   └── watcher.go        # Folder watching
│   ├── ingester/
│   │   └── ingester.go       # Hash, dedup, store
│   ├── consumer/
│   │   ├── consumer.go       # Consumer interface
│   │   ├── dispatcher.go     # Multi-consumer dispatch
│   │   ├── intervals/
│   │   │   └── intervals.go  # Intervals.icu
│   │   └── (future consumers)
│   └── retry/
│       └── retry.go          # Retry logic
├── go.mod
├── go.sum
└── README.md
```

---

## Dependencies

Minimal, pure Go:
- `github.com/fsnotify/fsnotify` - File watching
- `modernc.org/sqlite` - Pure Go SQLite (no CGO)
- `github.com/pelletier/go-toml/v2` - Config parsing
- `github.com/tormoder/fit` - FIT file parsing (pure Go)

---

## Open Questions

1. **Pure Go vs CGO for SQLite?**
   - ✅ **Decision: Pure Go** (`modernc.org/sqlite`)
   - Easier cross-compile, no C compiler needed
   - Slightly slower but negligible for our use case

2. **Credential storage?**
   - ✅ **Decision: Config file**
   - It's on their own machine, keep it simple
   - Can add keyring later if users request it

3. **Watch subdirectories?**
   - Current: Only direct children
   - Option: Recursive watching
   - Recommendation: Configurable, default to non-recursive

4. **Parse FIT files?**
   - ✅ **Decision: Yes, parse and extract rich metadata**
   - Store activity timestamp, type, duration, distance, power, HR
   - Enables richer UI for file management
   - Use `github.com/tormoder/fit` (pure Go FIT parser)

---

## Implementation Priority

### P0 - MVP
- [ ] SQLite store with fit_files, sync_records tables
- [ ] Folder watcher (existing)
- [ ] Intervals.icu consumer (existing)
- [ ] Hash-based deduplication
- [ ] Basic retry (3 immediate attempts)
- [ ] `fitwatch` (watch mode)
- [ ] `fitwatch --once` (sync and exit)

### P1 - Usability
- [ ] `fitwatch status` command
- [ ] `fitwatch retry` command
- [ ] Background retry loop
- [ ] Structured logging improvements

### P2 - Extensibility
- [ ] Additional consumers (TrainingPeaks, Strava)
- [ ] Cloud sources (iGPSport API)
- [ ] FIT file parsing for timestamps
- [ ] Keyring credential storage

---

## Decision: Proceed?

This design:
- Uses SQLite for reliability and queryability
- Keeps the pipeline abstraction (Source → Ingester → Consumer)
- Handles failures with retry logic
- Avoids over-engineering (no message queues, no microservices)

**Next step**: Implement SQLite store, then wire it into existing code.
