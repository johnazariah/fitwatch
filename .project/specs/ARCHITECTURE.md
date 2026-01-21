# FitBridge Architecture Overview

> **FitBridge** - Bridge the gap between your fitness platforms. Own your data.

## System Overview

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                                                                                      │
│   ┌──────────────────────────────────────────────────────────────────────────────┐  │
│   │  USER'S BROWSER                                                               │  │
│   │                                                                               │  │
│   │  ┌─────────────────────────┐     ┌──────────────────────────────────────────┐│  │
│   │  │ FitBridge Extension     │     │ FitBridge Web App (fitbridge.io)         ││  │
│   │  │                         │     │                                          ││  │
│   │  │ • Captures auth tokens  │────→│ • Dashboard                              ││  │
│   │  │ • Monitors TP, MW, Zwift│     │ • Activity feed                          ││  │
│   │  │ • Sends to backend      │     │ • Sync status                            ││  │
│   │  └─────────────────────────┘     │ • Analytics & insights                   ││  │
│   │                                  └──────────────────────────────────────────┘│  │
│   └──────────────────────────────────────────────────────────────────────────────┘  │
│                                                 │                                    │
└─────────────────────────────────────────────────┼────────────────────────────────────┘
                                                  │
                                                  ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│  AZURE                                                                               │
│                                                                                      │
│  ┌───────────────────────────────────────────────────────────────────────────────┐  │
│  │ Orleans Cluster (Container Apps)                                               │  │
│  │                                                                                │  │
│  │   UserGrain (per user)              ProviderGrain (per user+provider)          │  │
│  │   ┌─────────────────────┐           ┌─────────────────────────────┐            │  │
│  │   │ • Provider tokens   │──────────→│ • Sync logic                │            │  │
│  │   │ • Sync schedules    │           │ • Duplicate detection       │            │  │
│  │   │ • Activity index    │           │ • FIT download/upload       │            │  │
│  │   └─────────────────────┘           └─────────────────────────────┘            │  │
│  └───────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                      │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                 │
│  │ Key Vault   │  │ Blob Storage│  │ Azure OpenAI│  │ SignalR     │                 │
│  │             │  │             │  │             │  │             │                 │
│  │ Encrypted   │  │ FIT files   │  │ Activity    │  │ Real-time   │                 │
│  │ tokens      │  │ Metadata    │  │ analysis    │  │ updates     │                 │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘                 │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

## Components

### 1. FitBridge Connector (Browser Extension)

**Purpose:** Capture authentication tokens from fitness platforms without storing credentials.

**Supported Platforms:**
- TrainingPeaks (token from API requests)
- MyWhoosh (token from API requests)
- Zwift (token from API requests)

**Technology:** Chrome/Edge Extension (Manifest V3)

### 2. FitBridge Web App

**Purpose:** User dashboard for managing connections, viewing activities, insights.

**Features:**
- Entra ID authentication
- Provider connection status
- Activity timeline
- Sync history
- LLM-powered insights

**Technology:** Blazor / React (TBD)

### 3. Orleans Backend

**Purpose:** Per-user state management, sync orchestration, scheduling.

**Grains:**
- `UserGrain` - User's connections, settings, activity index
- `ProviderGrain` - Handles sync for one user+provider
- `AnalysisGrain` - LLM analysis for activities

**Technology:** .NET 10, F#, Orleans 8.x

### 4. Data Stores

| Store | Purpose |
|-------|---------|
| Key Vault | Encrypted tokens (per user+provider) |
| Blob Storage | FIT files, grain state, metadata |
| Table Storage | Activity index, sync history |

## Data Flow

### Token Capture Flow
```
1. User installs FitBridge extension
2. User logs into fitbridge.io (Entra ID)
3. Extension links to user's account
4. User clicks "Connect TrainingPeaks"
5. User logs into TrainingPeaks (any tab)
6. Extension intercepts Bearer token from API requests
7. Extension sends token to backend → Key Vault
8. Backend confirms connection, starts sync
```

### Sync Flow
```
1. Scheduler triggers ProviderGrain.SyncNow() (or user clicks sync)
2. Grain retrieves token from Key Vault
3. Grain fetches activities from source (TrainingPeaks, etc.)
4. Domain model: API response → ActivityMetadata
5. Duplicate detection against Intervals.icu
6. Download FIT files for new activities
7. Upload to Intervals.icu
8. Store FIT files in Blob Storage
9. Update sync state, notify web app via SignalR
```

## Key ADRs

| ADR | Decision |
|-----|----------|
| [ADR-012](decisions/012-duplicate-detection-strategy.md) | Multi-factor confidence-based duplicate detection |
| [ADR-013](decisions/013-source-activity-domain-model.md) | Unified domain model for cross-platform sync |
| [ADR-014](decisions/014-platform-api-integration-patterns.md) | Platform API patterns (auth, endpoints) |
| [ADR-015](decisions/015-fsharp-for-cli-and-domain.md) | F# for domain logic |
| [ADR-016](decisions/016-browser-extension-token-capture.md) | Browser extension for token capture |
| [ADR-017](decisions/017-orleans-cloud-architecture.md) | Orleans on Container Apps |

## Project Structure (Planned)

```
fitbridge/
├── src/
│   ├── FitBridge.Domain/           # F# - Core types, duplicate detection
│   ├── FitBridge.Providers/        # F# - TrainingPeaks, MyWhoosh, Intervals adapters
│   ├── FitBridge.Grains/           # F# - Orleans grain implementations
│   ├── FitBridge.Silo/             # Orleans silo host (Container Apps)
│   ├── FitBridge.Web/              # Blazor/React web app
│   └── FitBridge.Extension/        # Browser extension (TypeScript)
├── infra/
│   ├── bicep/                      # Azure infrastructure as code
│   └── container-apps/             # Container Apps config
├── .project/
│   ├── specs/decisions/            # ADRs
│   └── research/                   # Spike findings
└── spike/
    └── FitSync.Cli.FSharp/         # Current working spike
```

## Status

| Component | Status |
|-----------|--------|
| Domain model | ✅ Spike complete |
| TrainingPeaks adapter | ✅ Spike complete |
| MyWhoosh adapter | ✅ Spike complete |
| Intervals.icu adapter | ✅ Spike complete |
| Duplicate detection | ✅ Spike complete |
| Browser extension | 📋 Designed |
| Orleans grains | 📋 Designed |
| Web app | 📋 Not started |
| Azure infra | 📋 Not started |
