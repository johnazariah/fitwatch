# FitWatch

Local FIT file watcher with pluggable consumers. Watches directories for new `.fit` files and pushes them to configured destinations.

## Installation

### Download Binary (Recommended)

Download the latest release for your platform from [GitHub Releases](https://github.com/johnazariah/fitwatch/releases):

| Platform | Download |
|----------|----------|
| Windows | `fitwatch-windows-amd64.exe` |
| macOS (Intel) | `fitwatch-darwin-amd64` |
| macOS (Apple Silicon) | `fitwatch-darwin-arm64` |
| Linux | `fitwatch-linux-amd64` |

```powershell
# Windows - download and run
Invoke-WebRequest -Uri "https://github.com/johnazariah/fitwatch/releases/latest/download/fitwatch-windows-amd64.exe" -OutFile fitwatch.exe
.\fitwatch.exe --init
```

```bash
# macOS/Linux - download, make executable, and run
curl -L -o fitwatch https://github.com/johnazariah/fitwatch/releases/latest/download/fitwatch-darwin-arm64
chmod +x fitwatch
./fitwatch --init
```

### Build from Source

Requires [Go 1.21+](https://go.dev/dl/):

```bash
git clone https://github.com/johnazariah/fitwatch.git
cd fitwatch
go build -o fitwatch ./cmd/fitwatch
./fitwatch --init
```

### Go Install

```bash
go install github.com/johnazariah/fitwatch/cmd/fitwatch@latest
```

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         FitWatch                                │
│                                                                 │
│  ┌─────────────────┐    ┌─────────────┐    ┌─────────────────┐ │
│  │  Folder Watcher │ →  │  Dispatcher │ →  │   Consumers     │ │
│  │                 │    │             │    │                 │ │
│  │ • Zwift         │    │ Routes FIT  │    │ • Intervals.icu │ │
│  │ • TrainerRoad   │    │ to all      │    │ • (add more)    │ │
│  │ • Custom paths  │    │ consumers   │    │                 │ │
│  └─────────────────┘    └─────────────┘    └─────────────────┘ │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                     SQLite Database                         ││
│  │  • FIT file metadata (parsed)                               ││
│  │  • Sync status per consumer                                 ││
│  │  • Retry tracking                                           ││
│  └─────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
```

## Quick Start

```bash
# Initialize config
./fitwatch --init

# Edit config with your Intervals.icu credentials
# ~/.fitwatch/config.toml

# Run interactively (watches for new files)
./fitwatch

# Or sync existing files once and exit
./fitwatch --once
```

## Install as Service

FitWatch can run as a system service that starts automatically on boot.

### Windows

```powershell
# Run as Administrator
.\fitwatch.exe service install
.\fitwatch.exe service start

# Check status
.\fitwatch.exe service status

# View in Services.msc as "FitWatch"
```

### macOS

```bash
./fitwatch service install
./fitwatch service start

# Installed to ~/Library/LaunchAgents/
# Check with: launchctl list | grep fitwatch
```

### Linux (systemd)

```bash
# Run as root
sudo ./fitwatch service install
sudo ./fitwatch service start

# Check with: systemctl status fitwatch
```

### Service Commands

```bash
fitwatch service install    # Install as system service
fitwatch service uninstall  # Remove the service
fitwatch service start      # Start the service
fitwatch service stop       # Stop the service
fitwatch service restart    # Restart the service
fitwatch service status     # Show service status
```

## Configuration

Config file: `~/.fitwatch/config.toml`

```toml
# Directories to watch for FIT files
watch_dirs = [
    "C:\\Users\\You\\Documents\\Zwift\\Activities",
    "C:\\Users\\You\\Documents\\TrainerRoad"
]

[intervals]
enabled = true
athlete_id = "i12345"      # Your Intervals.icu athlete ID
api_key = "your-api-key"   # Settings → Developer Settings → API Key
```

## Finding Your Intervals.icu Credentials

1. Go to [intervals.icu](https://intervals.icu)
2. **Athlete ID**: Look at the URL when logged in: `intervals.icu/athlete/i12345` → ID is `i12345`
3. **API Key**: Settings → Developer Settings → Create API Key

## Adding New Consumers

Implement the `Consumer` interface:

```go
type Consumer interface {
    Name() string
    Push(ctx context.Context, fitPath string) error
    Validate() error
}
```

Example for a new destination:

```go
// internal/consumer/strava/strava.go
package strava

type Consumer struct {
    AccessToken string
}

func (c *Consumer) Name() string { return "Strava" }

func (c *Consumer) Push(ctx context.Context, fitPath string) error {
    // Upload to Strava API
    return nil
}

func (c *Consumer) Validate() error {
    if c.AccessToken == "" {
        return errors.New("access token required")
    }
    return nil
}
```

Then register in `main.go`:

```go
if cfg.Strava.Enabled {
    dispatcher.AddConsumer(strava.New(cfg.Strava.AccessToken))
}
```

## Default Watch Directories

### Windows
- `Documents\Zwift\Activities`
- `Documents\TrainerRoad`

### macOS
- `~/Documents/Zwift/Activities`
- `~/Documents/TrainerRoad`

### Linux
- `~/Documents/Zwift/Activities`
- `~/.local/share/Zwift/Activities`

## Command Line Options

```
Usage: fitwatch [options]

Options:
  -c string       Config file path (default ~/.fitwatch/config.toml)
  --config        Show config path and exit
  --init          Create default config and exit
  --once          Sync existing files and exit (no watch)
  -v              Verbose logging
  --version       Show version and exit
```

## How It Works

1. **Startup**: Loads config, initializes consumers, opens sync store
2. **Scan**: Checks watch directories for existing FIT files not yet synced
3. **Watch**: Monitors directories for new FIT files using OS file notifications
4. **Dispatch**: When a new FIT file appears, sends it to all enabled consumers
5. **Track**: Records successful syncs to avoid duplicates on restart

## Sync Store

The sync store (`~/.fitwatch/fitwatch.db`) is a SQLite database that tracks which files have been synced to which consumers. This ensures:
- Files aren't re-uploaded on restart
- Each consumer tracks its own sync state
- You can add new consumers and they'll sync existing files
- Full activity metadata is parsed and stored for querying

## Future Consumers

Planned:
- [ ] TrainingPeaks
- [ ] Strava
- [ ] Local copy (backup to folder)
- [ ] Webhook (POST to custom URL)

## License

MIT
