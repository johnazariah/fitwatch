# ADR-008: User Interfaces

## Status
Accepted

## Context
We need to decide which user interfaces to build and in what order. Options include web, mobile, CLI, desktop, and API.

## Decision

**Build Web + CLI + API for MVP. All share the same backend.**

### Interface Priority

| Interface | Priority | Technology | Rationale |
|-----------|----------|------------|-----------|
| **REST API** | P0 | ASP.NET Minimal APIs | Foundation for all UIs |
| **Web Dashboard** | P0 | Blazor Server | Primary user interface |
| **CLI** | P1 | System.CommandLine | Power users, automation |
| **Mobile Web** | P1 | Responsive Blazor | Works on phone immediately |
| **Desktop Tray** | P2 | .NET MAUI/WinUI | Status at a glance |
| **Native Mobile** | P3 | Defer | Too much effort for value |

### REST API

Foundation that all interfaces consume:

```
/api/activities          GET, POST      List/import activities
/api/activities/{id}     GET, DELETE    Activity details
/api/activities/{id}/fit GET            Download original FIT
/api/activities/{id}/summary GET        AI-generated summary

/api/sources             GET            List connected sources
/api/sources/{provider}  POST, DELETE   Connect/disconnect
/api/sources/{provider}/sync POST       Trigger manual sync

/api/sinks               GET            List connected sinks
/api/sinks/{provider}    POST, DELETE   Connect/disconnect
/api/sinks/{provider}/upload/{id} POST  Upload specific activity

/api/settings            GET, PUT       User preferences
/api/analysis/trends     GET            Training trends
/api/chat                POST           Chat with workout data
```

### Web Dashboard (Blazor Server)

Primary interface for most users:

```
┌─────────────────────────────────────────────────────────────────┐
│  FitSync                    [Sync Now]  [Settings]  [Profile]  │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────┐  ┌──────────────────────────────────────────┐  │
│  │  Sources    │  │  Recent Activities                        │  │
│  │  ─────────  │  │  ──────────────────                       │  │
│  │  ✓ Garmin   │  │  📍 Morning Ride         Today   45 min  │  │
│  │  ✓ MyWhoosh │  │     ✓ Strava  ✓ Intervals                │  │
│  │  + Add      │  │     "Solid endurance ride..."             │  │
│  │             │  │                                            │  │
│  │  Sinks      │  │  📍 Zwift Race           Yesterday 30min  │  │
│  │  ─────────  │  │     ✓ Strava  ⏳ Intervals               │  │
│  │  ✓ Strava   │  │     "Threshold intervals, good power"    │  │
│  │  ✓ Intervals│  │                                            │  │
│  │  + Add      │  │  📍 Recovery Spin        2 days ago 1hr   │  │
│  └─────────────┘  │     ✓ Strava  ✓ Intervals                │  │
│                    └──────────────────────────────────────────┘  │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Weekly Summary (AI)                                      │   │
│  │  ──────────────────                                       │   │
│  │  You rode 8 hours this week, up 15% from last week.      │   │
│  │  TSS: 450. Consider a rest day before your next hard...  │   │
│  │                                                           │   │
│  │  [💬 Ask a question about your training]                 │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

#### Why Blazor Server?

| Factor | Blazor Server | Blazor WASM |
|--------|---------------|-------------|
| Initial load | Fast | Slow (download runtime) |
| Works offline | No | Yes |
| SEO | Yes | No |
| Real-time updates | Easy (SignalR built-in) | Need to add |
| Server resources | Higher | Lower |

**Decision:** Blazor Server for MVP. Faster to build, real-time sync status updates are natural. Can migrate to WASM later if offline matters.

### CLI

For power users and automation:

```bash
# Import a FIT file
fitsync import ./morning-ride.fit

# Import all files from a folder
fitsync import ./garmin-exports/ --recursive

# List recent activities  
fitsync list --limit 10

# Show activity details
fitsync show abc123

# Trigger sync from all sources
fitsync sync

# Upload to a specific sink
fitsync upload abc123 --to strava

# Get AI summary
fitsync summarize abc123

# Export to folder
fitsync export abc123 --format fit --output ./backup/
```

Implementation with `System.CommandLine`:

```csharp
var rootCommand = new RootCommand("FitSync - Own your fitness data");

var importCommand = new Command("import", "Import FIT files");
importCommand.AddArgument(new Argument<string>("path"));
importCommand.AddOption(new Option<bool>("--recursive", "-r"));
importCommand.SetHandler(async (path, recursive) =>
{
    await _importService.ImportAsync(path, recursive);
}, pathArg, recursiveOpt);

rootCommand.AddCommand(importCommand);
```

### Mobile Strategy

**Responsive web first. PWA as enhancement.**

```csharp
// In Program.cs - enable PWA
builder.Services.AddProgressiveWebApp();
```

Users can "install" the web app on their phone's home screen. Gets 80% of native app benefits with 0% of the effort.

## Consequences

### Positive
- Single API serves all interfaces
- Blazor shares code with backend
- CLI enables scripting and automation
- PWA gives mobile presence without app store

### Negative
- Blazor Server requires constant connection
- CLI needs separate distribution

### Future Options
- Blazor Hybrid for true desktop app
- .NET MAUI for native mobile (if justified by demand)

## Related Decisions
- ADR-006: Azure + .NET Aspire Architecture
