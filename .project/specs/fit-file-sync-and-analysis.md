# FIT File Sync & Analysis Platform

## Vision

**Your workouts. Your data. Your choice.**

A platform that puts athletes in control of their fitness data by treating FIT files as the canonical, user-owned source of truth — not platform APIs with restrictive terms.

---

## Overview

An open fitness data interchange that:
- Aggregates FIT files from devices and user-friendly platforms
- Preserves complete data provenance and ownership
- Uploads to user-chosen destinations (prioritizing open platforms)
- Provides LLM-powered workout analysis with actionable improvement suggestions

---

## Problem Statement

Athletes and fitness enthusiasts face a fragmented landscape:

1. **Platform lock-in**: Workout data trapped in walled gardens (especially Strava)
2. **Ownership ambiguity**: Unclear who "owns" data once uploaded to platforms  
3. **Portability friction**: Difficult to move historical data between services
4. **Incomplete analysis**: Platforms optimize for engagement, not athlete improvement

**Our position:** The FIT file from your device is YOUR data. Platforms are just temporary homes for copies of that data. You should be able to:

1. Archive your complete workout history in a format you control
2. Sync to any platform without restriction
3. Leave any platform and take your data with you
4. Get intelligent insights that serve YOU, not advertisers

---

## Core Features

### 1. FIT File Sources (Import)

See [ADR-005: Platform Integration Policies](decisions/005-platform-integration-policies.md) for full rationale.

#### Tier 1: Open Platforms (Full Integration)
| Source | Method | Priority |
|--------|--------|----------|
| Local folder watch | Filesystem | P0 |
| Manual FIT upload | Web UI | P0 |
| Garmin Connect | OAuth API | P1 |
| Wahoo | OAuth API | P1 |
| Intervals.icu | API Key | P1 |

#### Tier 2: User-Initiated Import
| Source | Method | Priority |
|--------|--------|----------|
| MyWhoosh | Export + Import | P0 |
| Zwift | Local files from Companion | P1 |
| TrainerRoad | Export + Import | P2 |
| Platform bulk exports | ZIP import | P1 |

#### Tier 3: Restricted (Manual Export Only)
| Source | Method | Priority |
|--------|--------|----------|
| Strava | GDPR/bulk export only | P3 |

### 2. FIT File Sinks (Export)

Priority given to open platforms that respect data portability:

| Sink | Method | Priority | Notes |
|------|--------|----------|-------|
| Local archive | Filesystem | P0 | User's permanent record |
| Intervals.icu | API Key | P0 | Open, developer-friendly |
| Strava | OAuth | P1 | Upload only (no download) |
| TrainingPeaks | OAuth/API | P1 | Good API support |
| Garmin Connect | OAuth | P2 | FIT upload supported |
| Cloud backup | OAuth | P2 | OneDrive, Dropbox, GDrive |

### 3. LLM-Powered Analysis

#### 3.1 Workout Summary
- Parse FIT file data (power, heart rate, cadence, GPS, etc.)
- Generate natural language summary of the workout
- Identify key metrics and achievements

#### 3.2 Performance Analysis
- Compare against historical data
- Identify trends (improving/declining fitness)
- Detect anomalies (unusual HR, power drops, etc.)

#### 3.3 Training Suggestions
- Recovery recommendations based on training load
- Pacing strategy improvements
- Cadence/power optimization tips
- Heart rate zone adherence feedback
- Comparison to training plan (if available)

#### 3.4 Conversation Interface
- Chat with your workout data
- Ask questions like "How did my intervals compare to last week?"
- Get explanations for technical metrics

---

## Technical Architecture

See [ADR-006: Azure + .NET Aspire Architecture](decisions/006-azure-aspire-architecture.md) and [ADR-011: Cost-Optimized Storage](decisions/011-cost-optimized-storage.md) for full details.

### High-Level Components

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              Azure                                       │
├─────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                    .NET Aspire Application                       │    │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐   │    │
│  │  │   Web UI     │  │   REST API   │  │  Background Worker   │   │    │
│  │  │   (Blazor)   │  │  (Minimal    │  │  (Sync, Analysis)    │   │    │
│  │  │              │  │   APIs)      │  │                      │   │    │
│  │  └──────────────┘  └──────────────┘  └──────────────────────┘   │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │                    Azure Functions                               │    │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐   │    │
│  │  │  OAuth       │  │  Timer       │  │  FIT Parser          │   │    │
│  │  │  Callbacks   │  │  Triggers    │  │  (Python isolated)   │   │    │
│  │  └──────────────┘  └──────────────┘  └──────────────────────┘   │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │              Azure Storage Account (~$1-5/month)                 │    │
│  │  ┌────────────────┐ ┌────────────────┐ ┌──────────────────┐     │    │
│  │  │ Blob Storage   │ │ Table Storage  │ │ Queue Storage    │     │    │
│  │  │ (FIT files)    │ │ (metadata)     │ │ (job queues)     │     │    │
│  │  └────────────────┘ └────────────────┘ └──────────────────┘     │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│  ┌──────────────┐  ┌───────────────────────────────────────────────┐   │
│  │  Key Vault   │  │  Azure OpenAI (pay-per-token)                 │   │
│  │  (secrets)   │  │                                                │   │
│  └──────────────┘  └───────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
```

### Technology Stack

| Component | Technology | Rationale |
|-----------|------------|-----------|
| Orchestration | .NET Aspire 9 | Local dev, service discovery, observability |
| Web UI | Blazor Server | Real-time updates, shared .NET code |
| Backend | ASP.NET Minimal APIs | Fast, lightweight, OpenAPI support |
| Background Jobs | .NET Worker Service | Long-running sync and analysis |
| Event Handlers | Azure Functions (.NET) | OAuth callbacks, scheduled polling |
| FIT Parsing | Python Azure Function | Best library support (fitdecode) |
| File Storage | Azure Blob Storage | ~$0.01/GB, geo-redundant option |
| Database | Azure Table Storage | ~$0.00036/10K transactions |
| Message Queue | Azure Queue Storage | ~$0.00004/10K operations |
| LLM Integration | Azure OpenAI + Semantic Kernel | .NET-native, pay-per-token |
| Secrets | Azure Key Vault | Secure token storage |

**Estimated cost: ~$5-20/month** (vs $50+ with Cosmos/Service Bus)

### Sport Focus

**Cycling only for MVP.** See [ADR-010](decisions/010-cycling-first-architecture.md).

- Power-based metrics (NP, IF, TSS)
- Cycling-specific AI coaching
- Designed for future sport-sharding (running, etc.)

### Data Model

```
User
├── Connections (Table: status; Key Vault: tokens)
├── Activities (Table Storage)
│   ├── Summary (power, HR, cadence, TSS, etc.)
│   ├── Provenance (source, method, ownership)
│   ├── AI Summary (generated text)
│   └── Sync Status (per sink)
├── FIT Files (Blob Storage - originals preserved)
└── Settings (Table Storage)
```

---

## User Stories

### Source Integration
- As a user, I want to connect my Garmin account so my rides automatically sync
- As a user, I want to manually upload a FIT file when automatic sync isn't available
- As a user, I want to watch a local folder for new FIT files from my bike computer

### Sink Integration
- As a user, I want my workouts automatically uploaded to Strava
- As a user, I want to choose which sinks receive which types of activities
- As a user, I want to see the sync status for each activity

### Analysis
- As a user, I want a plain-English summary of my workout
- As a user, I want to know if I'm training too hard or not hard enough
- As a user, I want specific suggestions for my next workout
- As a user, I want to ask questions about my training data

---

## API Design (Draft)

### Sources
```
GET    /api/sources                    # List available source types
POST   /api/sources/connect            # Connect a new source
DELETE /api/sources/{id}               # Disconnect a source
POST   /api/sources/{id}/sync          # Trigger manual sync
```

### Sinks
```
GET    /api/sinks                      # List available sink types
POST   /api/sinks/connect              # Connect a new sink
DELETE /api/sinks/{id}                 # Disconnect a sink
POST   /api/sinks/{id}/upload/{activityId}  # Upload specific activity
```

### Activities
```
GET    /api/activities                 # List activities
GET    /api/activities/{id}            # Get activity details
POST   /api/activities/upload          # Manual FIT file upload
DELETE /api/activities/{id}            # Delete activity
```

### Analysis
```
GET    /api/activities/{id}/summary    # Get AI summary
GET    /api/activities/{id}/analysis   # Get detailed analysis
POST   /api/activities/{id}/chat       # Chat about activity
GET    /api/analysis/trends            # Get training trends
POST   /api/analysis/suggestions       # Get training suggestions
```

---

## Security Considerations

1. **OAuth Token Storage**: Encrypt tokens at rest
2. **FIT File Privacy**: Files contain GPS data - handle with care
3. **API Rate Limiting**: Respect source/sink API limits
4. **User Data Isolation**: Multi-tenant data separation
5. **LLM Data Handling**: Option to use local LLM for privacy

---

## MVP Scope

### Phase 1: Core Sync
- [ ] Local folder source
- [ ] Manual FIT upload
- [ ] MyWhoosh integration
- [ ] Strava sink
- [ ] Local archive sink
- [ ] Basic web UI

### Phase 2: Analysis
- [ ] FIT file parsing and storage
- [ ] Basic LLM summary generation
- [ ] Workout metrics dashboard
- [ ] Simple training suggestions

### Phase 3: Expansion
- [ ] Additional sources (Garmin, Wahoo)
- [ ] Additional sinks (TrainingPeaks, Intervals.icu)
- [ ] Chat interface for workout data
- [ ] Training load tracking
- [ ] Historical trend analysis

---

## Open Questions

See [decisions/](decisions/) folder for resolved questions.

| Question | Status | Decision |
|----------|--------|----------|
| Hosted vs self-hosted? | ✅ Decided | Azure-hosted with Aspire ([ADR-006](decisions/006-azure-aspire-architecture.md)) |
| Which LLM provider(s)? | ✅ Decided | Azure OpenAI + Semantic Kernel ([ADR-009](decisions/009-genai-integration.md)) |
| Duplicate detection? | ✅ Decided | SHA256 hash of FIT file ([ADR-004](decisions/004-data-provenance-tracking.md)) |
| Data format? | ✅ Decided | FIT files as canonical format ([ADR-003](decisions/003-fit-files-as-canonical-format.md)) |
| Platform policies? | ✅ Decided | Tiered by openness, avoid hostile APIs ([ADR-005](decisions/005-platform-integration-policies.md)) |
| Authentication? | ✅ Decided | Azure Entra ID + Functions for OAuth ([ADR-007](decisions/007-authentication-strategy.md)) |
| Interfaces? | ✅ Decided | Web (Blazor) + CLI + API ([ADR-008](decisions/008-user-interfaces.md)) |
| Sport scope? | ✅ Decided | Cycling only, sport-shardable later ([ADR-010](decisions/010-cycling-first-architecture.md)) |
| Storage? | ✅ Decided | Table/Queue/Blob for low cost ([ADR-011](decisions/011-cost-optimized-storage.md)) |
| Training plan integration? | 🔄 Pending | Defer to Phase 3+ |

---

## Success Metrics

- Number of activities synced per day
- Sync success rate (%)
- Time from workout completion to all sinks synced
- User engagement with analysis features
- User-reported training improvements

---

## References

- [FIT SDK](https://developer.garmin.com/fit/overview/)
- [Strava API](https://developers.strava.com/)
- [Garmin Connect API](https://developer.garmin.com/)
- [Intervals.icu API](https://intervals.icu/api)
