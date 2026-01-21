# Research: Intervals.icu Integrations & Import Capabilities

## Objective

Comprehensive inventory of Intervals.icu's current integration landscape to identify gaps that FitBridge could fill.

---

## 1. Official Integrations - Direct Connections

Intervals.icu offers several **direct OAuth/API integrations** with fitness platforms:

### ✅ Fully Supported (OAuth/Direct Sync)

| Platform | Sync Type | Direction | Data Quality | Notes |
|----------|-----------|-----------|--------------|-------|
| **Strava** | OAuth | Import | Full FIT via Strava | Primary integration path. Strava pulls from sources, Intervals.icu pulls from Strava |
| **Garmin Connect** | OAuth | Import | Full FIT | Direct sync, highly reliable. Best data quality |
| **Wahoo** | OAuth | Import | Full FIT | Via Wahoo Cloud API |
| **Polar** | OAuth | Import | Full data | Via Polar AccessLink API |
| **Suunto** | OAuth | Import | Full data | Via Suunto API |
| **Coros** | OAuth | Import | Full data | Added relatively recently (2023) |
| **Hammerhead** | OAuth | Import | Full FIT | Karoo devices |
| **Training Peaks** | OAuth | **Bidirectional** | Workouts/Activities | Can push planned workouts TO TrainingPeaks and import completed activities FROM |
| **Dropbox** | OAuth | Import | FIT files | Auto-sync from Dropbox folder |
| **Google Drive** | OAuth | Import | FIT files | Auto-sync from Drive folder |

### ⚠️ Partial/Indirect Support

| Platform | Status | How to Get Data Into Intervals.icu |
|----------|--------|-----------------------------------|
| **Zwift** | Indirect | Zwift → Strava → Intervals.icu (data preserved), OR Zwift → Garmin → Intervals.icu |
| **TrainerRoad** | Indirect | TrainerRoad → Strava → Intervals.icu |
| **Rouvy** | Indirect | Rouvy → Strava → Intervals.icu |
| **MyWhoosh** | ⚠️ **NEW ACTIVITIES ONLY** | Official OAuth integration added late 2024. **Only syncs new activities from connection date forward. NO historical backfill!** Manual FIT upload required for history. |
| **iGPSport** | ❌ **NOT SUPPORTED** | Manual FIT upload only |
| **Bryton** | Indirect | Limited - manual or Strava |
| **Lezyne** | Indirect | Manual or Strava |

---

## 2. Sync Methods Supported

### OAuth Connections
Intervals.icu's primary sync mechanism. User authorizes once, data flows automatically.

**Process:**
1. User goes to Settings → Connections
2. Clicks "Connect" for desired platform
3. OAuth dance completes
4. Activities sync automatically (polling, not webhooks)

**Sync Frequency:** Intervals.icu polls connected services every 15-30 minutes (not real-time).

### Direct FIT Upload (REST API)
Well-documented REST API for pushing activities:

```
POST /api/v1/athlete/{athleteId}/activities
Content-Type: multipart/form-data
Authorization: Basic base64(API_KEY:your-api-key)

file: <FIT file binary>
```

**API Key Access:** Free for personal use. Get from Settings page.

**Cookbook:** https://forum.intervals.icu/t/intervals-icu-api-integration-cookbook/80090

### Manual File Upload
- Upload via web interface
- Supports: FIT, TCX, GPX, PWX
- Drag-and-drop or file picker

### Folder Sync (Cloud Storage)
- Dropbox and Google Drive folders
- Watches for new FIT files
- Good for platforms that export to cloud storage

---

## 3. Platform-Specific Analysis

### Zwift Integration Status: 🟡 INDIRECT

**Current State:**
- No direct Zwift → Intervals.icu OAuth
- Zwift has NO public API (closed ecosystem)
- Users must chain: Zwift → Strava/Garmin → Intervals.icu

**Data Flow:**
```
Zwift Ride → Zwift Companion → [Strava OR Garmin Connect] → Intervals.icu
```

**What Gets Lost:**
- Generally complete via Strava pathway
- FIT files are uploaded intact
- Power, HR, cadence, GPS all preserved

**Pain Points:**
- Extra hop required (Strava account needed)
- 15-30 minute delay for sync
- If Strava is down, sync breaks

**FitBridge Opportunity:** ✅ Can provide direct FIT file access from local Zwift Activities folder, bypassing Strava entirely.

---

### MyWhoosh Integration Status: ⚠️ PARTIAL SUPPORT (NEW ACTIVITIES ONLY)

**Current State (Updated January 2026):**
- **Official OAuth integration added late 2024** (approx. October-November 2024)
- Integration was announced on Intervals.icu forum by David Tinker
- MyWhoosh provided partner API access to enable this

**How it Works:**
1. Go to Intervals.icu → Settings → Connections
2. Click "Connect" for MyWhoosh  
3. OAuth authorization flow with MyWhoosh
4. Activities sync automatically (polled every 15-30 minutes)
5. Full FIT file data preserved (power, HR, cadence, virtual route)

**⚠️ CRITICAL LIMITATION: No Historical Backfill**
- Integration ONLY syncs activities from **connection date forward**
- Does NOT backfill any historical MyWhoosh rides
- If you've been riding MyWhoosh for 2 years before connecting, those rides stay in MyWhoosh only

**Why No Historical Sync?**
1. MyWhoosh partner API designed for forward-sync events, not bulk historical queries
2. Rate limiting concerns for backfilling years of data
3. Common practice in fitness API integrations (Strava, Garmin work the same way)
4. Resource constraints for small Intervals.icu team

**Manual Process for Historical Data:**
1. Log into MyWhoosh web portal
2. Navigate to each activity individually
3. Download FIT file
4. Go to Intervals.icu
5. Upload FIT file
6. Repeat for every historical activity 😓

**FitBridge Opportunity:** ✅ **MAJOR GAP REMAINS** - The integration exists but historical backfill is a major pain point. FitBridge can:
- Bulk download all historical activities via token capture
- Automatically upload to Intervals.icu API
- One-time backfill + ongoing sync for platforms without integration

---

### iGPSport Integration Status: ❌ NOT SUPPORTED

**Current State:**
- No direct integration
- iGPSport Cloud has limited export options
- Users manually export FIT files from iGPSport app
- Manual upload to Intervals.icu required

**FitBridge Opportunity:** ✅ **MAJOR GAP** - Significant user base with no automation.

---

### TrainingPeaks Integration Status: ✅ BIDIRECTIONAL

**Current State:**
- OAuth integration exists
- Can **import activities** FROM TrainingPeaks
- Can **push workouts** TO TrainingPeaks (planned workouts, workout library)

**Unique Feature:** Only sink that Intervals.icu can push TO (not just pull from).

**Limitations:**
- Requires TrainingPeaks premium account
- Not real-time sync

**FitBridge Opportunity:** 🟡 Limited - integration already exists. BUT: Could bridge TrainingPeaks → Intervals.icu for users without premium TP accounts.

---

### Garmin Connect Integration Status: ✅ FULL SUPPORT

**Current State:**
- OAuth direct integration
- Highly reliable
- Full FIT file access
- Most recommended path for device users

**No FitBridge Value Here:** Garmin integration works great.

---

### Strava Integration Status: ✅ FULL SUPPORT

**Current State:**
- OAuth direct integration
- Primary path for many platforms
- Full FIT file access via Strava
- Works well

**Considerations:**
- Strava acts as a hub
- 1000s of apps push to Strava
- Intervals.icu pulls from Strava

**FitBridge Opportunity:** 🟡 Could bypass Strava hop for direct platform connections.

---

### Wahoo Integration Status: ✅ FULL SUPPORT

Direct OAuth via Wahoo Cloud. No gap.

---

### Polar Integration Status: ✅ FULL SUPPORT

Via Polar AccessLink API. No gap.

---

### Suunto Integration Status: ✅ FULL SUPPORT

Via Suunto API. No gap.

---

### Coros Integration Status: ✅ SUPPORTED

Added in 2023. Direct OAuth. No gap.

---

### Apple Health Integration Status: ❌ NOT SUPPORTED

**Current State:**
- No direct Apple Health/HealthKit integration
- Intervals.icu is web-based, can't access HealthKit directly
- Users must use intermediary (Strava app, Health Fit app, etc.)

**Technical Reason:** HealthKit requires native iOS app access.

**FitBridge Opportunity:** 🟡 Challenging - would require iOS app component. Not desktop-friendly.

---

### Samsung Health / Google Fit / Health Connect Status: ❌ NOT SUPPORTED

**Current State:**
- No Android health aggregator integrations
- Same limitation as Apple Health - web app can't access mobile health APIs

**FitBridge Opportunity:** 🟡 Challenging - would require Android app component.

---

## 4. What's Missing: Gap Analysis

### Major Gaps (High FitBridge Value)

| Platform | Gap Type | User Pain Level | FitBridge Solution |
|----------|----------|-----------------|-------------------|
| **MyWhoosh** | Integration exists BUT no historical backfill | 🔴 High for existing users | Bulk historical download + upload |
| **iGPSport** | No integration at all | 🔴 High | Direct API/token sync |
| **Zwift (direct)** | Indirect only via Strava | 🟡 Medium | Local FIT folder watching |

### Minor Gaps (Some FitBridge Value)

| Platform | Gap Type | User Pain Level | FitBridge Solution |
|----------|----------|-----------------|-------------------|
| **TrainingPeaks (free tier)** | API requires paid account | 🟡 Medium | Token-based access for free users |
| **Bryton** | Indirect/manual only | 🟡 Medium | Could add support |
| **Lezyne** | Indirect/manual only | 🟡 Medium | Could add support |

### Non-Gaps (Intervals.icu Already Solves)

- Garmin Connect ✅
- Strava ✅
- Wahoo ✅
- Polar ✅
- Suunto ✅
- Coros ✅
- TrainingPeaks ✅
- Hammerhead ✅
- Dropbox/Google Drive file sync ✅

---

## 5. Intervals.icu API Capabilities

### Documented REST API

**Base URL:** `https://intervals.icu/api/v1`

**Authentication:** HTTP Basic Auth
- Username: `API_KEY`
- Password: Your personal API key from Settings

**Key Endpoints:**
```
GET  /athlete/{id}                    # Get athlete profile
GET  /athlete/{id}/activities         # List activities
POST /athlete/{id}/activities         # Upload FIT file
PUT  /activity/{id}                   # Update activity
GET  /athlete/{id}/events             # Get calendar events
POST /athlete/{id}/events             # Create planned workout
GET  /athlete/{id}/wellness           # Get wellness data
POST /athlete/{id}/wellness           # Update wellness data
```

**Rate Limits:** Not explicitly documented, but reasonable use expected.

### Developer-Friendly

Intervals.icu's creator, David Tinker, is:
- Active on the community forum
- Responsive to integration questions
- Open to third-party tools

**Forum Evidence:**
- API integration cookbook maintained by community
- Multiple third-party tools use the API
- Golden Cheetah integration exists
- WKO5 import scripts exist

---

## 6. Is Intervals.icu Open to Third-Party Integrations?

### ✅ YES - Very Open

**Evidence:**
1. **Public API with free access** - No approval process needed
2. **Active forum with integration discussions** - David answers questions directly
3. **Integration cookbook** - Community-maintained documentation
4. **No rate limiting lockdowns** - Reasonable use policy
5. **Open to feature requests** - Has added integrations on request (Coros, etc.)

### Potential for Deeper Partnership?

Intervals.icu is:
- Small operation (1-2 developers)
- Premium model ($10/month for full features)
- Might be interested in FitBridge as it brings more users to platform

**Could FitBridge:**
- Be listed as a recommended tool?
- Get webhook support for real-time sync?
- Collaborate on gap platforms (MyWhoosh, iGPSport)?

---

## 7. Technical Integration Details

### For FitBridge → Intervals.icu

**Already implemented in your codebase:**
- [Intervals.fs](../../spike/FitSync.Cli.FSharp/Intervals.fs) - F# API client
- [IntervalsIcuClient.cs](../../spike/FitSync.Cli/Services/IntervalsIcuClient.cs) - C# API client

**Required Config:**
- `intervals:apikey` - API key from user's Intervals.icu settings
- `intervals:athleteid` - Athlete ID (e.g., `i12345`)

**Upload Flow:**
```
1. User provides API key + Athlete ID (one-time setup)
2. FitBridge downloads FIT from source (MyWhoosh, iGPSport, Zwift)
3. FitBridge calls POST /api/v1/athlete/{id}/activities
4. Intervals.icu receives FIT, processes, creates activity
```

### Duplicate Handling

Intervals.icu has built-in duplicate detection:
- Based on start time + duration + device serial
- Will reject obvious duplicates
- FitBridge should also pre-check (time-based matching)

---

## 8. Recommendations for FitBridge

### Primary Value Proposition
Focus on the **platforms Intervals.icu CAN'T integrate with**:

1. **MyWhoosh** - No existing integration, users frustrated
2. **iGPSport** - No existing integration, growing user base
3. **Zwift Direct** - Bypass Strava hop, reduce latency

### Secondary Value
- **TrainingPeaks (free users)** - API requires premium
- **Consolidation** - Single tool for multiple gaps

### Integration Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        FitBridge                            │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────────────┐ │
│  │MyWhoosh │  │iGPSport │  │  Zwift  │  │ TrainingPeaks   │ │
│  │ Token   │  │ Token   │  │ Folder  │  │ Token (free)    │ │
│  │ Capture │  │ Capture │  │ Watch   │  │ Capture         │ │
│  └────┬────┘  └────┬────┘  └────┬────┘  └────────┬────────┘ │
│       │            │            │                │          │
│       └────────────┴─────┬──────┴────────────────┘          │
│                          │                                  │
│                    ┌─────▼─────┐                            │
│                    │   Sync    │                            │
│                    │   Engine  │                            │
│                    └─────┬─────┘                            │
│                          │                                  │
└──────────────────────────┼──────────────────────────────────┘
                           │
                    ┌──────▼──────┐
                    │ Intervals   │
                    │   .icu API  │
                    │ (REST/FIT)  │
                    └─────────────┘
```

### Messaging
Position FitBridge as:
> "The missing link for platforms Intervals.icu can't reach"

Not:
> "A replacement for Intervals.icu integrations" (they're good!)

---

## 9. Summary Table

| Platform | Intervals.icu Status | FitBridge Value | Priority |
|----------|---------------------|-----------------|----------|
| MyWhoosh | ⚠️ New activities only (no backfill) | 🔥 **Critical for historical data** | P0 |
| iGPSport | ❌ None | 🔥 **High** | P0 |
| Zwift (direct) | 🟡 Indirect | 🟡 Medium | P1 |
| TrainingPeaks (free) | 🟡 Premium only | 🟡 Medium | P2 |
| Garmin | ✅ Full | ❌ None | Skip |
| Strava | ✅ Full | ❌ None | Skip |
| Wahoo | ✅ Full | ❌ None | Skip |
| Polar | ✅ Full | ❌ None | Skip |
| Suunto | ✅ Full | ❌ None | Skip |
| Coros | ✅ Full | ❌ None | Skip |
| Apple Health | ❌ None | 🟡 Needs iOS app | P3+ |
| Samsung Health | ❌ None | 🟡 Needs Android app | P3+ |

---

## 10. Sources & References

1. Intervals.icu Settings/Connections page
2. Intervals.icu API documentation (forum)
3. https://forum.intervals.icu/t/intervals-icu-api-integration-cookbook/80090
4. Community discussions on missing integrations
5. Your existing codebase research

---

## Next Steps

1. [ ] Finalize MyWhoosh → Intervals.icu as MVP
2. [ ] Add iGPSport as second source
3. [ ] Implement Zwift folder watching for direct sync
4. [ ] Consider reaching out to David Tinker about partnership
5. [ ] Document FitBridge publicly for Intervals.icu community discovery
