# Research: Authentication Alternatives to Token Lifting

## Objective

1. Evaluate all possible authentication approaches for syncing user data from fitness platforms
2. **Comprehensively inventory ALL sources and sinks of fitness data** in the ecosystem

## Current Approach: Token Lifting

**How it works:**
- Browser extension intercepts API requests
- Extracts Bearer tokens from Authorization headers
- Sends to desktop app for use in API calls

**Pros:**
- Works for any platform with a web interface
- No partnership or approval needed
- User controls their own data

**Cons:**
- Tokens expire (Zwift: 6hrs, others: 7-15 days)
- Requires user to have extension installed
- Feels "hacky" to technical users
- Could break if platforms detect/block

---

## Part 1: Comprehensive Source/Sink Inventory

### Indoor Training Platforms (Sources)

| Platform | Type | Data Format | Sync Options | API Status |
|----------|------|-------------|--------------|------------|
| **Zwift** | Virtual cycling/running | FIT | Strava, TP, Garmin | Unknown |
| **MyWhoosh** | Virtual cycling | FIT | Strava, TP | Unknown |
| **TrainerRoad** | Structured training | FIT | Strava, TP, Garmin | Has API? |
| **Rouvy** | Virtual cycling | FIT | Strava | Unknown |
| **SYSTM (Wahoo)** | Structured training | FIT | Wahoo Cloud | Unknown |
| **Fulgaz** | Virtual cycling | FIT | Strava | Unknown |
| **RGT (Wahoo)** | Virtual cycling | FIT | Wahoo Cloud | Deprecated? |
| **Kinomap** | Virtual cycling | ? | Strava | Unknown |
| **Bkool** | Virtual cycling | FIT | Strava | Unknown |

### Bike Computers / Head Units (Sources)

| Device | Manufacturer | Data Format | Sync Method | API Status |
|--------|--------------|-------------|-------------|------------|
| **Garmin Edge series** | Garmin | FIT | Garmin Connect | OAuth available |
| **Wahoo ELEMNT/BOLT/ROAM** | Wahoo | FIT | Wahoo Cloud | Unknown |
| **Hammerhead Karoo** | Hammerhead | FIT | Cloud sync | Unknown |
| **Bryton Rider** | Bryton | FIT | Bryton Active | Unknown |
| **Lezyne GPS** | Lezyne | FIT | Lezyne GPS Root | Unknown |
| **Sigma ROX** | Sigma | FIT | Sigma Cloud | Unknown |
| **iGPSport** | iGPSport | FIT | iGPSport Cloud | Token-based |
| **Coros DURA** | Coros | FIT | Coros app | Unknown |

### Watches / Wearables (Sources)

| Device | Manufacturer | Data Format | Sync Method | API Status |
|--------|--------------|-------------|-------------|------------|
| **Apple Watch** | Apple | HealthKit | Apple Health | HealthKit SDK |
| **Garmin watches** | Garmin | FIT | Garmin Connect | OAuth available |
| **Wahoo RIVAL** | Wahoo | FIT | Wahoo Cloud | Unknown |
| **Coros watches** | Coros | FIT | Coros app | Unknown |
| **Polar watches** | Polar | FIT | Polar Flow | Has API |
| **Suunto watches** | Suunto | FIT | Suunto app | Has API |
| **Samsung Galaxy Watch** | Samsung | Samsung Health | Samsung Health | SDK available |
| **Fitbit** | Google | Proprietary | Fitbit Cloud | OAuth API |
| **Whoop** | Whoop | Proprietary | Whoop Cloud | Has API |
| **Oura Ring** | Oura | Proprietary | Oura Cloud | Has API |

### Power Meters / Sensors (Sources - via head unit)

| Device | Data | Notes |
|--------|------|-------|
| **Stages** | Power | Syncs via head unit |
| **Quarq** | Power | Syncs via head unit |
| **Favero Assioma** | Power | Syncs via head unit |
| **Garmin Rally/Vector** | Power | Syncs via head unit |
| **4iiii** | Power | Syncs via head unit |
| **SRM** | Power | Has own software |
| **Power2Max** | Power | Syncs via head unit |

### Smart Trainers (Sources - can record independently)

| Device | Manufacturer | Data | API |
|--------|--------------|------|-----|
| **Wahoo KICKR** | Wahoo | FIT | Via Wahoo app |
| **Tacx NEO** | Garmin | FIT | Via Garmin Connect |
| **Elite Direto** | Elite | FIT | Via my E-Training |
| **Saris H3** | Saris | FIT | Via Rouvy? |
| **JetBlack VOLT** | JetBlack | FIT | ? |

### Mobile Apps (Sources)

| App | Platform | Data Format | Export | API |
|-----|----------|-------------|--------|-----|
| **Strava (record)** | iOS/Android | FIT | Yes | OAuth |
| **Wahoo Fitness** | iOS/Android | FIT | Yes | ? |
| **Garmin Connect** | iOS/Android | FIT | Via web | OAuth |
| **Zwift Companion** | iOS/Android | N/A | N/A | Same as Zwift |
| **MapMyRide** | iOS/Android | ? | ? | ? |
| **Ride with GPS** | iOS/Android | GPX/TCX | Yes | Has API |
| **Komoot** | iOS/Android | GPX | Yes | Has API |

### Health Aggregators (Sources AND Sinks)

| Platform | Type | Data Format | Permissions | API |
|----------|------|-------------|-------------|-----|
| **Apple Health** | iOS aggregator | HealthKit | Per-app | HealthKit SDK |
| **Samsung Health** | Android aggregator | Samsung format | Per-app | SDK |
| **Google Fit** | Android aggregator | Google format | OAuth | REST API |
| **Health Connect** | Android (new) | Standard | Per-app | Android API |

### Analysis Platforms (Sinks - Primary)

| Platform | Focus | Data Format | Import Methods | API |
|----------|-------|-------------|----------------|-----|
| **Intervals.icu** | Cycling analytics | FIT | Upload, Strava, Garmin | REST API ✅ |
| **TrainingPeaks** | Coaching/planning | FIT | Upload, many integrations | Has API |
| **Strava** | Social/segments | FIT/GPX | Upload, many integrations | OAuth |
| **Golden Cheetah** | Desktop analytics | FIT | Manual import | Local |
| **WKO5** | Power analysis | FIT | Manual import | Local |
| **Xert** | AI training | FIT | Strava, Garmin | Unknown |
| **Today's Plan** | Coaching | FIT | Upload | Unknown |
| **Final Surge** | Coaching | FIT | Upload | Unknown |
| **2Peak** | Training plans | FIT | Upload | Unknown |
| **TrainAsONE** | AI running | ? | Strava, Garmin | Unknown |

### Social / Sharing Platforms (Sinks)

| Platform | Focus | Data | Integration |
|----------|-------|------|-------------|
| **Strava** | Social | FIT | Direct upload |
| **Relive** | Video creation | GPX | Strava import |
| **Veloviewer** | Strava analytics | Via Strava | Strava OAuth |
| **StatsHunters** | Strava heatmap | Via Strava | Strava OAuth |
| **Wandrer** | Exploration | Via Strava | Strava OAuth |

### Data Standards

| Format | Description | Contains |
|--------|-------------|----------|
| **FIT** | Garmin binary | Full data: GPS, power, HR, cadence, temp |
| **TCX** | Garmin XML | GPS, HR, cadence, power |
| **GPX** | GPS exchange | GPS only, optional extensions |
| **PWX** | TrainingPeaks | Similar to TCX |
| **JSON** | Various | Platform-specific |

---

## Part 2: Intervals.icu Integration Landscape

**Key Finding:** Intervals.icu is VERY developer-friendly and already has extensive integrations. We should focus on gaps, not reinvent what's solved.

### ✅ Already Integrated (OAuth/Direct Sync)

| Platform | Method | Notes |
|----------|--------|-------|
| **Strava** | OAuth | Primary hub for many users |
| **Garmin Connect** | OAuth | Best data quality (native FIT) |
| **Wahoo** | OAuth | ELEMNT/KICKR devices |
| **Polar** | OAuth | AccessLink API |
| **Suunto** | OAuth | Direct integration |
| **Coros** | OAuth | Added 2023 |
| **TrainingPeaks** | OAuth | **Bidirectional** sync |
| **Hammerhead** | OAuth | Karoo devices |
| **Zwift** | OAuth | Added Dec 2024 - **includes historical data!** |
| **MyWhoosh** | OAuth | Added late 2024 - **NEW activities only** |
| **Dropbox** | Folder sync | Upload FIT files to folder |
| **Google Drive** | Folder sync | Upload FIT files to folder |

### ⚠️ Partial Integration (FitBridge Opportunities)

| Platform | Status | Gap | FitBridge Value |
|----------|--------|-----|-----------------|
| **MyWhoosh** | ✅ OAuth exists | ❌ No historical backfill | 🔥 **CRITICAL** - Import past rides |
| **iGPSport** | ❌ No integration | Full gap | 🔥 **HIGH** - No OAuth at all |

### Zwift Integration Details (Dec 2024)

**What changed:**
- Direct OAuth integration announced December 2024
- **Includes historical data backfill** - users get their full Zwift history
- No longer need to route through Strava
- Source: https://zwiftinsider.com/intervals-integration/

**Why this matters:**
- Zwift (larger company) built API with bulk historical export support
- This sets precedent - historical sync IS technically possible
- Makes MyWhoosh's limitation more glaring by comparison

**FitBridge implication:** Zwift is no longer a gap we need to fill.

### MyWhoosh Integration Details

**Current State (as of late 2024):**
- Intervals.icu added official MyWhoosh OAuth integration
- Connect via Settings → Connections → MyWhoosh
- Polls for new activities every 15-30 minutes
- Full data quality (FIT files with power, HR, cadence, virtual route)

**The Limitation:**
- ⚠️ **Only syncs activities FROM the connection date forward**
- ❌ No historical backfill - can't import past rides via OAuth
- Manual workaround: Download FIT files one-by-one from MyWhoosh, upload to Intervals.icu

**Contrast with Zwift:**
> Zwift's integration includes full historical sync. MyWhoosh's doesn't.
> Users will ask: "Why can Zwift do this but MyWhoosh can't?"

**Why This Matters for FitBridge:**
> "Intervals.icu now syncs MyWhoosh going forward, but can't import your historical rides. FitBridge gives you that one-time backfill to bring your full training history together."

This is a **cleaner value prop** - we complement the official integration for a specific use case they don't handle.

### Intervals.icu Developer Resources

- **Free public API** - No approval process, just get API key
- **REST endpoint for FIT upload** - `POST /api/v1/athlete/{id}/activities`
- **Active community forum** - Creator (David Tinker) responds directly
- **API Cookbook** - https://forum.intervals.icu/t/intervals-icu-api-integration-cookbook/80090
- **Third-party integrations** - Many already exist

### FitBridge Strategic Position

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         INTERVALS.ICU                                   │
│                                                                         │
│   ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐          │
│   │ Strava  │ │ Garmin  │ │ Wahoo   │ │ Polar   │ │  Coros  │   ...    │
│   │   ✅    │ │   ✅    │ │   ✅    │ │   ✅    │ │   ✅    │          │
│   └─────────┘ └─────────┘ └─────────┘ └─────────┘ └─────────┘          │
│                                                                         │
│   ┌───────────────────────────────────────────────────────────┐        │
│   │                    CAN'T REACH                            │        │
│   │   ┌──────────┐  ┌──────────┐  ┌──────────────────────┐   │        │
│   │   │ MyWhoosh │  │ iGPSport │  │ Zwift (direct, not   │   │        │
│   │   │    ❌    │  │    ❌    │  │ via Strava delay)    │   │        │
│   │   └────┬─────┘  └────┬─────┘  └──────────┬───────────┘   │        │
│   │        │             │                   │               │        │
│   └────────┼─────────────┼───────────────────┼───────────────┘        │
│            │             │                   │                         │
│            └─────────────┴─────────┬─────────┘                         │
│                                    │                                   │
│                          ┌─────────▼─────────┐                         │
│                          │    FITBRIDGE      │                         │
│                          │  Desktop + Ext    │                         │
│                          │                   │                         │
│                          │ Token capture OR  │                         │
│                          │ Local FIT watch   │                         │
│                          └─────────┬─────────┘                         │
│                                    │                                   │
│                                    │ Push via API                      │
│                                    ▼                                   │
└─────────────────────────────────────────────────────────────────────────┘
```

### Revised Value Proposition

**We're not building another sync tool for platforms that are already connected.**

We're building a **bridge for the gaps**:

| Gap | User Pain | FitBridge Solution |
|-----|-----------|-------------------|
| MyWhoosh → Intervals.icu | Manual FIT download/upload | Auto-sync |
| iGPSport → Intervals.icu | Export, convert, upload | Auto-sync |
| Zwift → Intervals.icu (direct) | Wait for Strava relay (15-30 min) | Instant local FIT |

### Platforms to SKIP (Already Solved)

Don't waste effort on:
- ❌ Garmin → Intervals.icu (OAuth works great)
- ❌ Strava → Intervals.icu (OAuth works great)  
- ❌ Wahoo → Intervals.icu (OAuth works great)
- ❌ Polar → Intervals.icu (OAuth works great)
- ❌ Suunto → Intervals.icu (OAuth works great)
- ❌ Coros → Intervals.icu (OAuth works great)
- ❌ TrainingPeaks ↔ Intervals.icu (bidirectional OAuth)

### Updated Priority Matrix

| Source | Sink | Priority | Reason |
|--------|------|----------|--------|
| **MyWhoosh (historical)** | Intervals.icu | **P0** | OAuth exists but no backfill - our MVP |
| **iGPSport** | Intervals.icu | **P0** | No integration exists at all |
| MyWhoosh (historical) | TrainingPeaks | P2 | Already have TP client |
| iGPSport | TrainingPeaks | P2 | Already have TP client |

### Platforms NO LONGER Gaps (Solved by Official Integrations)

| Platform | Why We Don't Need To Handle |
|----------|----------------------------|
| **Zwift** | ✅ Full OAuth + historical backfill (Dec 2024) |
| Garmin | ✅ Full OAuth |
| Strava | ✅ Full OAuth |
| Wahoo | ✅ Full OAuth |
| Polar | ✅ Full OAuth |
| Suunto | ✅ Full OAuth |
| Coros | ✅ Full OAuth |
| TrainingPeaks | ✅ Bidirectional OAuth |
| Hammerhead | ✅ Full OAuth |

### Refined Value Proposition

**FitBridge is NOT a replacement for existing integrations.**

We solve **specific gaps** that official OAuth can't:

| Use Case | Official Solution | FitBridge Solution |
|----------|-------------------|-------------------|
| MyWhoosh → Intervals.icu (new) | ✅ OAuth works | Not needed |
| MyWhoosh → Intervals.icu (historical) | ❌ Manual only | ✅ Bulk backfill |
| iGPSport → Intervals.icu | ❌ None | ✅ Full sync |
| Zwift → Intervals.icu | ✅ OAuth + historical | ~~Not needed~~ |

### Narrowed Focus

FitBridge's remaining value is very specific:

1. **MyWhoosh Historical Backfill** - One-time import of past rides
2. **iGPSport Full Sync** - No official integration exists

This is a **smaller but clearer market**. Users who:
- Just switched from Zwift to MyWhoosh and want their history unified
- Have years of iGPSport data with no way to get it into Intervals.icu
- Need to migrate between platforms

---

## Part 3: Revised Strategy - Local-First Approach

### The Insight

Even platforms with OAuth integrations have limitations:
- **Polling delay**: 15-30 minutes before activities appear
- **API dependency**: If OAuth breaks, sync breaks
- **Privacy**: Your data routes through their servers

**Local FIT folder watching** solves all of these:
- ⚡ **Instant**: Activity appears in Intervals.icu within seconds of saving
- 🔒 **Private**: Data goes directly from your machine to Intervals.icu
- 🛡️ **Reliable**: No OAuth tokens to expire, no API changes to break
- 📴 **Offline-capable**: Queue uploads when connection returns

### Where FIT Files Live Locally

| Platform | Local FIT Path | Notes |
|----------|----------------|-------|
| **Zwift** | `Documents/Zwift/Activities/` | Saves after every ride |
| **TrainerRoad** | `Documents/TrainerRoad/` | Check exact path |
| **Rouvy** | TBD | Research needed |
| **MyWhoosh** | TBD | Research needed |
| **Wahoo SYSTM** | TBD | Research needed |
| **Garmin (USB)** | `GARMIN/Activity/` on device | When connected |
| **Wahoo (USB)** | Device storage | When connected |

### Revised Priority Matrix

| Feature | Priority | Value | Effort |
|---------|----------|-------|--------|
| **Local FIT folder watching** | **P0** | Universal, instant sync | Medium |
| **iGPSport cloud sync** | **P0** | No integration exists anywhere | Already built |
| **MyWhoosh historical** | **P1** | One-time backfill bonus | Already built |
| **USB device detection** | **P2** | Garmin/Wahoo when plugged in | Medium |

### Revised Value Proposition

**Before (API-focused):**
> "FitBridge syncs platforms that don't have OAuth"

**After (Local-first):**
> "FitBridge watches your local FIT files and instantly pushes them to Intervals.icu - no waiting for OAuth polling, no cloud dependencies"

### Why This Is Better

| Aspect | OAuth Integration | FitBridge Local Watch |
|--------|-------------------|----------------------|
| **Speed** | 15-30 min delay | ⚡ Instant |
| **Privacy** | Data routes through platform servers | 🔒 Direct to Intervals.icu |
| **Reliability** | OAuth can break | 🛡️ Just file watching |
| **Offline** | Requires connection | 📴 Queues for later |
| **Coverage** | Per-platform integration | 🌐 Any app that saves FIT |

### Target Users

1. **Performance-focused athletes** who want instant feedback
   - Finish ride → check Intervals.icu immediately
   - No waiting for Strava → Intervals.icu relay

2. **Privacy-conscious users** who don't want data routing everywhere
   - Direct path: Local FIT → Intervals.icu
   - No Zwift API, no Strava middleman

3. **iGPSport users** with no other option
   - FitBridge is the only bridge to Intervals.icu

4. **MyWhoosh switchers** (bonus)
   - One-time historical import when migrating

### Architecture Implications

```
┌─────────────────────────────────────────────────────────────────┐
│                    FitBridge Desktop App                        │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │              LOCAL FOLDER WATCHER (P0)                  │   │
│  │                                                         │   │
│  │   Documents/Zwift/Activities/     ──┐                   │   │
│  │   Documents/TrainerRoad/          ──┼──► FIT Parser     │   │
│  │   [User-configured paths]         ──┘        │          │   │
│  │                                              ▼          │   │
│  │                                    Intervals.icu API    │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │              CLOUD SYNC (P0 for gaps)                   │   │
│  │                                                         │   │
│  │   iGPSport API (token) ──────────► FIT Download         │   │
│  │                                          │              │   │
│  │                                          ▼              │   │
│  │                                    Intervals.icu API    │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │              HISTORICAL IMPORT (P1 bonus)               │   │
│  │                                                         │   │
│  │   MyWhoosh API (token) ──────────► Bulk FIT Download    │   │
│  │                                          │              │   │
│  │                                          ▼              │   │
│  │                                    Intervals.icu API    │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Key Differentiator

**FitBridge is not about filling OAuth gaps anymore.**

**FitBridge is about LOCAL-FIRST fitness data sync:**
- Watch folders, not APIs
- Instant, not polling
- Private, not routed through clouds
- Universal, not per-platform

iGPSport cloud sync is the exception (no local files) and MyWhoosh historical is a migration tool.

---

## Part 4: Should We Contribute to Intervals.icu Instead?

### The Realization

Our scope has narrowed to essentially:
1. Watch local FIT folders → push to Intervals.icu
2. iGPSport cloud → push to Intervals.icu  
3. MyWhoosh historical → push to Intervals.icu

**We're building an Intervals.icu companion app.** Maybe we should just... contribute to Intervals.icu?

### What We Know About Intervals.icu

| Aspect | Details |
|--------|---------|
| **Creator** | David Tinker (solo/small team) |
| **Business** | Freemium - $10/year supporter tier |
| **API** | Free, public, well-documented |
| **Community** | Active forum, David responds directly |
| **Source** | Not open source (web app) |
| **Desktop** | No desktop app exists |

### Options to Explore

#### Option A: Propose Integration Directly

Contact David Tinker via the Intervals.icu forum:
- "Would you be interested in adding iGPSport integration?"
- "Have you considered a desktop companion for local FIT watching?"
- Share our research on gaps

**Pros:**
- Gets features to ALL Intervals.icu users
- David knows the codebase
- No separate app to maintain

**Cons:**
- Dependent on his priorities/time
- He may not be interested
- Slower than building ourselves

#### Option B: Build & Offer as Official Companion

Build FitBridge focused purely on Intervals.icu, then:
- Offer it to David as an official companion app
- "Intervals.icu Desktop Sync" branding
- He distributes, we maintain (or hand off)

**Pros:**
- We ship something fast
- Proven concept before proposing
- Could become official tool

**Cons:**
- Duplicated effort if he's already working on it
- May not want third-party code

#### Option C: Build as Independent Tool, Intervals.icu First

Keep FitBridge independent but:
- Make Intervals.icu the primary (only?) sink
- Maybe add TrainingPeaks/Strava later
- Stay "Intervals.icu ecosystem adjacent"

**Pros:**
- Full control
- Can expand to other sinks later
- No dependency on David's approval

**Cons:**
- Small market (Intervals.icu users only)
- Competing with potential official tool

#### Option D: Open Source Contribution

If Intervals.icu ever open-sources or accepts plugins:
- Contribute iGPSport integration as a PR
- Contribute local FIT watcher as a PR

**Pros:**
- Direct integration
- Community benefit

**Cons:**
- Intervals.icu isn't open source
- No plugin system currently

### Research Tasks

- [ ] **Check Intervals.icu forum** for existing discussions about:
  - iGPSport integration requests
  - Local FIT folder watching requests
  - Desktop app requests
- [ ] **Contact David Tinker** - Gauge interest in:
  - iGPSport as a new integration
  - Desktop companion app concept
  - Accepting contributions
- [ ] **Check if others have built this** - Search for:
  - Intervals.icu desktop tools
  - FIT folder watchers for Intervals.icu

### Questions to Answer

1. **Has anyone already built this?**
   - A FIT folder watcher for Intervals.icu seems obvious
   - Someone must have done it?

2. **Why hasn't David added iGPSport?**
   - Not enough demand?
   - iGPSport won't provide API access?
   - Just hasn't gotten to it?

3. **Would David accept a companion app?**
   - He's responsive on forums
   - Worth asking directly

4. **What's the Intervals.icu roadmap?**
   - Is a desktop app planned?
   - Are there feature requests for this?

### Recommended Next Step

Before building more:

1. **Post on Intervals.icu forum** asking:
   > "Is there interest in a desktop companion app that watches local FIT folders (Zwift, etc.) for instant sync? Also interested in iGPSport integration. Happy to contribute if useful."

2. **Wait for response** from David/community

3. **Decide based on feedback:**
   - "Yes please build it" → Build FitBridge
   - "I'm working on it" → Don't duplicate
   - "Here's how to contribute" → Contribute
   - No response → Build independently

### If We Proceed Independently

Rename/rebrand to make the focus clear:

| Current | Proposed |
|---------|----------|
| FitBridge | **Intervals Sync** or **FIT Watch** |
| "Cross-platform fitness sync" | "Instant local sync for Intervals.icu" |

And scope down:
- ❌ TrainingPeaks sync
- ❌ Strava sync  
- ❌ Multi-sink architecture
- ✅ Local FIT folder watching → Intervals.icu
- ✅ iGPSport → Intervals.icu
- ✅ MyWhoosh historical → Intervals.icu (bonus)

---

### 1. Official OAuth / API Programs

For each platform, research:
- Does an official developer API exist?
- Is OAuth available for third-party apps?
- What's the approval process?
- Are there rate limits or costs?
- What data is accessible?

| Platform | Research Questions |
|----------|-------------------|
| **Zwift** | Does Zwift have a public API? Partner program? Has anyone gotten official access? |
| **MyWhoosh** | Any developer documentation? Contact for API access? |
| **iGPSport** | Official API? Export mechanisms? |
| **TrainingPeaks** | Known to have API - what are requirements? |
| **Intervals.icu** | Already using their API - document OAuth flow if available |

### 2. Platform Webhooks / Push

- Do any platforms offer webhooks for new activities?
- Can we register a callback URL?
- This would eliminate polling entirely

### 3. Native App Integration

- Can we integrate with the desktop apps (Zwift launcher, etc.)?
- Do they store tokens we could access locally?
- Are there local databases with activity data?

### 4. FIT File Direct Access

- Where do platforms store FIT files locally?
- Zwift: `Documents/Zwift/Activities/`?
- Can we just watch these folders instead of using APIs?

### 5. Strava as a Hub

- Most platforms sync to Strava
- Strava has OAuth API
- Could we pull from Strava instead of each source?
- What data is lost in the Strava intermediary?

### 6. Garmin Connect IQ / Wahoo SDK

- Can we build a "sync to FitBridge" data field?
- Would capture data at the source device

### 7. Open Source / Community Approaches

Research existing projects:
- How does GoldenCheetah do it?
- How does intervals.icu get data?
- Sauce for Strava extension approach?
- zwift-offline project?

---

## Evaluation Criteria

For each alternative, score on:

| Criterion | Weight | Description |
|-----------|--------|-------------|
| **Reliability** | High | Will it keep working? Official > Scraping |
| **User Experience** | High | How much user effort? OAuth > Token paste |
| **Data Completeness** | Medium | Full FIT file or summary only? |
| **Legal/ToS Risk** | Medium | Could we get blocked or sued? |
| **Implementation Effort** | Low | How hard to build? |
| **Platform Coverage** | High | Works for all 4+ platforms? |

---

## Specific Research Tasks

### Task 1: Zwift API Investigation
- [ ] Search for "Zwift API documentation"
- [ ] Check if Zwift has a developer portal
- [ ] Research Zwift partner integrations (how did TrainerRoad, Wahoo connect?)
- [ ] Check Zwift community forums for API discussions
- [ ] Examine zwift-offline, zwift-packet-monitor projects
- [ ] Find where Zwift stores local FIT files

### Task 2: MyWhoosh API Status
- [ ] Contact MyWhoosh about API access
- [ ] Check if they have Strava integration (implies OAuth capability)
- [ ] Document their current export options

### Task 3: iGPSport Investigation  
- [ ] Check iGPSport developer resources
- [ ] How do they sync to Strava/TrainingPeaks?
- [ ] Can we tap into that existing flow?

### Task 4: TrainingPeaks API
- [ ] Review TrainingPeaks API documentation
- [ ] What's required for API access approval?
- [ ] OAuth flow details
- [ ] Rate limits and restrictions

### Task 5: Strava Hub Approach
- [ ] Can we get FIT files from Strava API, or just summaries?
- [ ] What's lost when activities route through Strava?
- [ ] OAuth implementation complexity
- [ ] Rate limits for free apps

### Task 6: Local FIT File Paths
For each platform, document:
- [ ] Where FIT files are stored locally
- [ ] File naming convention
- [ ] When files appear (immediately after ride? After sync?)
- [ ] Can we avoid API entirely with folder watching?

### Task 7: Legal/ToS Review
- [ ] Review each platform's Terms of Service
- [ ] Check for API abuse clauses
- [ ] Research precedents (has anyone been banned for scraping?)

### Task 8: Apple Health / HealthKit
- [ ] What data is accessible via HealthKit?
- [ ] Can we read workout data from other apps?
- [ ] Export mechanisms (Health app export to folder?)
- [ ] Can a macOS app read HealthKit via iCloud sync?
- [ ] What permissions are required?
- [ ] Can we write back to HealthKit (sink)?

### Task 9: Samsung Health
- [ ] Samsung Health SDK capabilities
- [ ] Export to FIT/TCX options
- [ ] Can we read workouts from other apps via Samsung Health?
- [ ] Health Connect (new Android API) - does it replace Samsung Health SDK?
- [ ] Privacy permissions model

### Task 10: Google Fit / Health Connect
- [ ] Google Fit REST API status (deprecated?)
- [ ] Health Connect as the new standard
- [ ] What fitness apps write to Health Connect?
- [ ] Can we read full workout data or just summaries?
- [ ] OAuth flow for Google Fit

### Task 11: Garmin Connect API
- [ ] Garmin Connect IQ SDK capabilities
- [ ] Garmin Health API (for partners)
- [ ] OAuth flow documentation
- [ ] What data is accessible?
- [ ] Has anyone built unofficial access?

### Task 12: Wahoo Cloud
- [ ] Wahoo Cloud API status
- [ ] How do KICKR/ELEMNT devices sync?
- [ ] Can we intercept the sync flow?
- [ ] SYSTM integration

### Task 13: Polar Flow
- [ ] Polar AccessLink API documentation
- [ ] OAuth requirements
- [ ] Data export capabilities
- [ ] Rate limits

### Task 14: Suunto App
- [ ] Suunto API availability
- [ ] Movescount (deprecated) vs new Suunto app
- [ ] Export mechanisms

### Task 15: Coros
- [ ] Coros API status
- [ ] How does TrainingPeaks/Strava sync work?
- [ ] Can we tap into that?

### Task 16: Fitbit (Google)
- [ ] Fitbit Web API documentation
- [ ] OAuth flow
- [ ] What data is accessible?
- [ ] Is it being merged into Google Fit?

### Task 17: Whoop
- [ ] Whoop API availability
- [ ] Third-party integration options
- [ ] Data export

### Task 18: Intervals.icu Import Methods
- [ ] How does Intervals.icu pull from Garmin/Strava?
- [ ] Can we replicate their approach?
- [ ] Is there a way to push to Intervals.icu via API (not just OAuth)?

---

## Platform Connectivity Map

Research how data flows between platforms:

```
                    ┌─────────────────────────────────────────────┐
                    │           Health Aggregators                │
                    │  ┌──────────┐ ┌──────────┐ ┌─────────────┐ │
                    │  │  Apple   │ │ Samsung  │ │   Google    │ │
                    │  │  Health  │ │  Health  │ │ Fit/Health  │ │
                    │  │          │ │          │ │   Connect   │ │
                    │  └────▲─────┘ └────▲─────┘ └──────▲──────┘ │
                    └───────┼────────────┼─────────────┼─────────┘
                            │            │             │
     ┌──────────────────────┼────────────┼─────────────┼──────────────────────┐
     │                      │            │             │                      │
     │    ┌─────────┐   ┌───┴────┐   ┌───┴────┐   ┌────┴───┐   ┌─────────┐   │
     │    │  Zwift  │   │ Garmin │   │ Wahoo  │   │ Polar  │   │ MyWhoosh│   │
     │    │         │   │Connect │   │ Cloud  │   │  Flow  │   │         │   │
     │    └────┬────┘   └───┬────┘   └───┬────┘   └────┬───┘   └────┬────┘   │
     │         │            │            │             │            │        │
     │         └────────────┴─────┬──────┴─────────────┴────────────┘        │
     │                            │                                          │
     │                       ┌────▼────┐                                     │
     │                       │ STRAVA  │  ◄── Central Hub?                   │
     │                       └────┬────┘                                     │
     │                            │                                          │
     │              ┌─────────────┼─────────────┐                            │
     │              │             │             │                            │
     │         ┌────▼────┐  ┌─────▼─────┐ ┌─────▼─────┐                      │
     │         │Intervals│  │TrainingPks│ │ Veloviewer│                      │
     │         │   .icu  │  │           │ │ etc.      │                      │
     │         └─────────┘  └───────────┘ └───────────┘                      │
     │                                                                       │
     │                         Analysis Sinks                                │
     └───────────────────────────────────────────────────────────────────────┘
```

**Key Questions:**
1. Is Strava truly central, or do many platforms have direct integrations?
2. Which platforms write to health aggregators vs read from them?
3. Where is the best interception point for maximum coverage?

---

## Expected Output

After research, produce:

1. **Comparison Matrix** - All approaches scored on criteria
2. **Recommendation** - Best approach per platform
3. **Hybrid Strategy** - Different methods for different platforms?
4. **Risk Assessment** - What could break and how to mitigate

---

## Initial Hypotheses to Validate

### Authentication Approaches
1. **Zwift likely has no public API** - Large platforms often don't, to control ecosystem
2. **Strava is a viable hub** - But may lose power data fidelity
3. **Local FIT files exist for most platforms** - Folder watching may be simpler
4. **TrainingPeaks has real API** - They're developer-friendly historically
5. **Token lifting works but isn't future-proof** - One platform blocking kills the feature

### Health Aggregators
6. **Apple Health is iOS-only** - No macOS access without iCloud workarounds
7. **Health Connect is Android's future** - Samsung Health SDK may be deprecated
8. **Health aggregators have summary data only** - Full FIT files may not be accessible
9. **Strava syncs to health aggregators** - May be the bridge between fitness and health

### Data Flow
10. **Strava is the de facto hub** - Most platforms push to Strava first
11. **Garmin Connect is a viable alternative hub** - Wide device support
12. **Local FIT files are the most reliable source** - No API dependencies
13. **Some platforms lose data in sync** - Power metrics, especially, get lost

---

## Strategic Considerations

### Desktop-First Implications
Our Go + Wails desktop app can:
- Watch local folders (FIT files)
- Make API calls (official or token-based)
- Integrate with OS-level health APIs (HealthKit on macOS?)
- Run as a background service

### Mobile Consideration
For Apple/Samsung Health, we may need:
- A companion iOS/Android app
- Or accept that we only sync via Strava

### Recommended Architecture Research

```
                     ┌─────────────────────────────────┐
                     │      FitBridge Desktop App      │
                     │  ┌───────────┬───────────────┐  │
                     │  │  Folder   │    API        │  │
                     │  │  Watcher  │    Clients    │  │
                     │  └─────┬─────┴───────┬───────┘  │
                     │        │             │          │
                     └────────┼─────────────┼──────────┘
                              │             │
           ┌──────────────────┼─────────────┼──────────────────┐
           │                  │             │                  │
    ┌──────▼──────┐    ┌──────▼──────┐    ┌▼────────────┐    ┌▼────────────┐
    │ Local FIT   │    │   Strava    │    │   Garmin    │    │   Token     │
    │   Files     │    │    OAuth    │    │   OAuth     │    │   Capture   │
    │             │    │             │    │             │    │  (fallback) │
    └─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
         ▲                   ▲                  ▲                   ▲
         │                   │                  │                   │
    Zwift, etc.         Many apps          Garmin devices      Zwift, MyWhoosh
    save locally        sync here          + software           iGPSport
```

### Priority Matrix

| Approach | Coverage | Reliability | UX | Priority |
|----------|----------|-------------|-----|----------|
| Local FIT folders | Medium | High | Great | P0 |
| Strava OAuth | High | High | Good | P0 |
| Garmin OAuth | Medium | High | Good | P1 |
| TrainingPeaks API | Low | High | Good | P1 |
| Token capture | High | Low | Poor | P2 (fallback) |
| Health aggregators | Low | Medium | Poor | P3 |

---

## Timeline

- Research phase: 4-6 hours (expanded scope)
- Document findings
- Update ADR-020 with conclusions
- Prioritize implementation based on coverage/effort ratio
