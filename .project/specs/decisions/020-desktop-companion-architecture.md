# ADR-020: Desktop Companion Architecture

## Status
Proposed

## Context

ADR-019 defines a cloud-first serverless architecture. However, analysis reveals an alternative **desktop-first** topology that:

1. Eliminates infrastructure costs entirely
2. Solves the token refresh problem naturally
3. Addresses local FIT file sync (USB bike computers)
4. Provides offline-first operation

This ADR defines the desktop companion as an **alternate deployment topology**, not a replacement for cloud-first. Users may choose based on their needs.

## Decision

### Desktop-First Topology

```
┌─────────────────────────────────────────────────────────┐
│                 FitBridge Desktop                        │
│                                                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐   │
│  │   Browser    │  │   Folder     │  │    Sync      │   │
│  │  Extension   │  │   Watcher    │  │   Engine     │   │
│  │   Bridge     │  │              │  │              │   │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘   │
│         │                 │                 │            │
│         ▼                 ▼                 ▼            │
│  ┌─────────────────────────────────────────────────┐    │
│  │              Local Activity Store               │    │
│  │  ~/.fitbridge/                                  │    │
│  │  ├── activities/     (FIT files)                │    │
│  │  ├── register.json   (provenance, ADR-018)      │    │
│  │  └── config.json     (tokens, settings)         │    │
│  └─────────────────────────────────────────────────┘    │
│                          │                               │
│                          ▼                               │
│  ┌─────────────────────────────────────────────────┐    │
│  │ Sources               │ Sinks                   │    │
│  │ ────────              │ ─────                   │    │
│  │ • Zwift API           │ • Intervals.icu         │    │
│  │ • MyWhoosh API        │ • Strava                │    │
│  │ • iGPSport API        │ • TrainingPeaks         │    │
│  │ • TrainingPeaks API   │ • Local archive         │    │
│  │ • Garmin USB folder   │                         │    │
│  │ • Wahoo USB folder    │                         │    │
│  └─────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────┘
```

### Core Responsibilities

| Component | Responsibility |
|-----------|----------------|
| **Extension Bridge** | Receive tokens from browser extension via localhost HTTP or native messaging |
| **Folder Watcher** | Monitor directories for new FIT files (Garmin, Wahoo USB mounts) |
| **Sync Engine** | Poll sources, download FIT files, push to sinks, deduplicate |
| **Local Store** | Persist FIT files and provenance metadata |
| **System Tray** | Background operation, status, quick actions |

### Tech Stack Evaluation

**Constraint:** Must be fully standalone—no runtime installation (Java, .NET, Python) required.

| Technology | Bundles | Binary Size | Memory | Dev Speed | Platforms |
|------------|---------|-------------|--------|-----------|-----------|
| **Tauri + Rust** | Native webview | 3-10MB | ~20MB | Medium | Win/Mac/Linux |
| **Go + Wails** | Native webview | 10-20MB | ~30MB | Fast | Win/Mac/Linux |
| **Electron** | Chromium + Node | 150-200MB | 100-200MB | Fastest | Win/Mac/Linux |
| **Python + PyInstaller** | Python runtime | 50-100MB | ~80MB | Fast | Win/Mac/Linux |

**Eliminated options:**
- .NET MAUI / F# + Avalonia — Requires .NET runtime or 50MB+ self-contained publish
- Java/Kotlin — Requires JVM

### Recommendation: Go + Wails

**Rationale:**

1. **Standalone binary** - Single .exe, no runtime required
2. **Small footprint** - 10-20MB installer, ~30MB memory
3. **Cross-platform** - Windows, macOS, Linux from single codebase
4. **Debuggable** - Clear error messages, easy to trace
5. **Web UI** - Reuse extension popup styling, HTML/CSS/JS
6. **Fast development** - Go compiles instantly, simple language
7. **Good HTTP** - net/http is battle-tested, easy JSON handling

**Wails v2 features:**
- Native webview (Edge/WebKit), not bundled Chromium
- Go ↔ JavaScript bindings (call Go from frontend)
- System tray support
- Auto-update support
- Single binary output

### Alternative: Tauri + Rust (if size is critical)

If 10MB → 3MB matters:
- Tauri produces smaller binaries
- Rust is more memory-safe
- Steeper learning curve
- Consider for v2 if Go proves limiting

## Implementation Phases

### Phase 1: Core Sync (2 weeks)
- [ ] Tauri app scaffold with system tray
- [ ] Local HTTP server for extension bridge
- [ ] Config management (tokens, settings)
- [ ] Single platform sync (start with Intervals.icu sink)

### Phase 2: Sources (2 weeks)
- [ ] Port platform clients (Zwift, MyWhoosh, iGPSport)
- [ ] Folder watcher for USB devices
- [ ] Local FIT file storage with register

### Phase 3: Polish (1 week)
- [ ] Auto-start on login
- [ ] Notification for sync status
- [ ] Settings UI (sync interval, watched folders)
- [ ] Auto-update mechanism

### Phase 4: Premium Features (Future)
- [ ] AI coaching via cloud API or local Ollama
- [ ] Advanced analytics dashboard
- [ ] Backup/restore

## Extension Integration

The browser extension (ADR: browser-extension.md) communicates with desktop app:

```
Extension captures token
         │
         ▼
POST http://localhost:5847/api/tokens
{
  "platform": "zwift",
  "token": "eyJ...",
  "capturedAt": "2026-01-21T..."
}
         │
         ▼
Desktop app stores token, triggers sync
```

Port 5847 chosen to avoid common conflicts (5000 = .NET, 3000 = React, etc.)

## Comparison: Desktop-First vs Cloud-First

| Aspect | Desktop-First | Cloud-First (ADR-019) |
|--------|---------------|----------------------|
| **Infrastructure cost** | $0 | ~$5-20/month |
| **Sync when PC off** | ❌ Waits for PC | ✅ Always running |
| **Multi-device** | ❌ Per-machine | ✅ Shared state |
| **Privacy** | ✅ All data local | ⚠️ Tokens on server |
| **Offline support** | ✅ Native | ❌ Requires internet |
| **Local FIT files** | ✅ Native | ❌ Must upload |
| **Token refresh** | ✅ Extension bridge | ⚠️ Needs notification |
| **Setup complexity** | ⚠️ Install app | ✅ Just extension |
| **Mobile access** | ❌ None | ✅ PWA possible |

## User Segmentation

| User Type | Recommended Topology |
|-----------|---------------------|
| Power user, privacy-conscious | Desktop-First |
| USB bike computer user | Desktop-First |
| Multi-device / mobile needs | Cloud-First |
| "Just works" preference | Cloud-First |
| Developer / self-hoster | Desktop-First |

## Consequences

### Positive
- Zero ongoing infrastructure cost
- Complete privacy—no tokens leave user's machine
- Local FIT file sync solves real user pain
- Offline-first resilient architecture
- Simpler product (no auth, no accounts)

### Negative
- Must install desktop app (friction)
- Sync stops when computer is off
- No mobile access
- Must maintain cross-platform app
- Rust learning curve if choosing Tauri

### Risks
- Platform API changes require app updates (mitigated by auto-update)
- macOS code signing costs ($99/year Apple Developer)
- User may forget to run app (mitigated by auto-start)

## Related Decisions
- ADR-018: Local Activity Register
- ADR-019: Serverless Multi-Tenant Architecture (alternative topology)
- Browser Extension Spec: features/browser-extension.md
