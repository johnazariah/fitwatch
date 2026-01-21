# Research: TrainingPeaks → Intervals.icu Sync

## Time Box
**Allocated:** 4 hours  
**Started:** 2026-01-21  
**Completed:** 2026-01-21

## Objective

### Questions to Answer
1. [x] Can we access TrainingPeaks data without official API partnership?
2. [x] What authentication mechanism does TrainingPeaks use?
3. [x] Can we download FIT files from TrainingPeaks?
4. [x] How do we detect duplicate activities when syncing to Intervals.icu?
5. [x] What domain model do we need for cross-platform activity sync?

### Success Criteria
- [x] Can list workouts from TrainingPeaks
- [x] Can download FIT files from TrainingPeaks
- [x] Can detect duplicates before uploading
- [x] Have reusable domain model for future sources

## Background

TrainingPeaks restricts API access to "Partner API" which requires a business relationship. However, users own their own data and should be able to export it. We investigated using the same web API that the TrainingPeaks web application uses.

## Findings

### Question 1: Can we access TrainingPeaks data without official API partnership?

**Answer:** Yes, using the internal web API with a captured bearer token.

**Evidence:** Successfully listed 40 workouts and downloaded FIT files using browser-captured token.

**API Details:**
```
Base URL: https://tpapi.trainingpeaks.com
Auth: Bearer token from browser Authorization header
```

---

### Question 2: What authentication mechanism does TrainingPeaks use?

**Answer:** Bearer token authentication. Token can be captured from browser dev tools (Network tab → any API request → Authorization header).

**Token Lifetime:** Approximately hours (exact TTL not determined). Requires periodic re-capture.

**Capture Process:**
1. Log into trainingpeaks.com
2. Open Dev Tools → Network tab
3. Navigate to Calendar or Workouts page
4. Find request to `tpapi.trainingpeaks.com`
5. Copy `Authorization: Bearer <token>` value

---

### Question 3: Can we download FIT files from TrainingPeaks?

**Answer:** Yes, but requires TWO API calls per workout.

**Key Discovery:** The list endpoint does NOT include the file ID needed for download. Must fetch workout details first.

**API Flow:**
```
1. List workouts (date range):
   GET /fitness/v6/athletes/{athleteId}/workouts/{startDate}/{endDate}
   Returns: [{ workoutId, title, workoutDay, totalTime, ... }]
   NOTE: No fileId in response!

2. Get workout details:
   GET /fitness/v6/athletes/{athleteId}/workouts/{workoutId}/details
   Returns: { workoutDeviceFileInfos: [{ fileId, fileName, ... }] }

3. Download file:
   GET /fitness/v6/athletes/{athleteId}/workouts/{workoutId}/rawfiledata/{fileId}
   Returns: FIT file bytes (may be gzipped)
```

**Gotcha:** `fileId` can be negative (e.g., `-568394398`). This is valid.

---

### Question 4: How do we detect duplicate activities when syncing to Intervals.icu?

**Answer:** Multi-factor confidence-based matching using time, duration, and distance.

**Initial Approach (failed):** Date-only matching
- Problem: Multiple workouts per day (AM/PM sessions) caused false positives
- Problem: TrainingPeaks date format `2024-11-10T00:00:00` vs Intervals.icu `2024-11-10T10:30:00`

**Final Approach:** Confidence scoring

```fsharp
type DuplicateMatch =
    | NoDuplicate
    | ProbableDuplicate of SinkActivity * confidence: float
    | ExactDuplicate of SinkActivity

// Factors:
// - Time within 5 minutes: 0.5 weight (required)
// - Duration within 10%: 0.3 weight
// - Distance within 10%: 0.2 weight
// 
// Score ≥ 0.8 = ExactDuplicate (skip)
// Score 0.5-0.8 = ProbableDuplicate (skip, flag)
// Score < 0.5 = NoDuplicate (upload)
```

See: [ADR-012: Duplicate Detection Strategy](decisions/012-duplicate-detection-strategy.md)

---

### Question 5: What domain model do we need for cross-platform activity sync?

**Answer:** Unified domain types with source adapters.

**Core Types:**
```fsharp
type Source = MyWhoosh | TrainingPeaks | Wahoo | Garmin | Zwift
type Sink = IntervalsIcu | Strava | TrainingPeaks

type ActivityMetadata = {
    SourceId: string
    Source: Source
    Title: string option
    ActivityType: ActivityType
    StartTime: DateTimeOffset
    Duration: TimeSpan option
    Distance: float option
    NormalizedPower: float option
    TSS: float option
}

type SourceActivity = {
    Metadata: ActivityMetadata
    FitData: byte[]
}

type SinkActivity = {
    SinkId: string
    Sink: Sink
    StartTime: DateTimeOffset
    Duration: TimeSpan option
    Distance: float option
}
```

**Adapter Pattern:**
```fsharp
// Each source provides:
module TrainingPeaks =
    let toMetadata (workout: Workout) : ActivityMetadata = ...
    let fetchActivity (token: string) (athleteId: int) (workout: Workout) : Async<SourceActivity option> = ...

// Each sink provides:
module Intervals =
    let toSinkActivity (activity: Activity) : SinkActivity = ...
    let listSinkActivities (...) : Async<Result<SinkActivity list, string>> = ...
```

See: [ADR-013: Source Activity Domain Model](decisions/013-source-activity-domain-model.md)

---

## Recommendations

### Immediate
1. **Token refresh UX:** Add clear error message when token expires, with instructions to refresh
2. **Dry-run mode:** Add `--dry-run` flag to show what would sync without uploading
3. **Rate limiting:** Add delays between API calls to avoid throttling

### Future
1. **OAuth implementation:** Register as TrainingPeaks partner for proper token refresh
2. **Local cache:** Store downloaded FIT files locally before uploading
3. **Incremental sync:** Track last sync timestamp to avoid re-scanning entire history

## Artifacts

### Code
- `spike/FitSync.Cli.FSharp/` - F# CLI with working sync
- `spike/FitSync.Cli.FSharp/Domain.fs` - Domain model
- `spike/FitSync.Cli.FSharp/TrainingPeaks.fs` - TP API client

### ADRs Created
- [ADR-012: Duplicate Detection Strategy](decisions/012-duplicate-detection-strategy.md)
- [ADR-013: Source Activity Domain Model](decisions/013-source-activity-domain-model.md)
- [ADR-014: Platform API Integration Patterns](decisions/014-platform-api-integration-patterns.md)
- [ADR-015: F# for CLI and Domain Logic](decisions/015-fsharp-for-cli-and-domain.md)

## Open Questions

1. **Token refresh automation:** Can we use Playwright/Selenium to automate token capture?
2. **Rate limits:** What are TrainingPeaks' rate limits? We haven't hit them yet.
3. **Gzip handling:** Are TP FIT files always gzipped? Need to detect and decompress.
