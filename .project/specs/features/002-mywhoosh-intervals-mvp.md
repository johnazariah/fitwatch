# Feature: MyWhoosh → Intervals.icu Sync (MVP)

## Priority
P0 - This is the MVP

## Summary
Automatically sync cycling activities from MyWhoosh to Intervals.icu, with AI-generated summaries.

## Goals
1. Prove the end-to-end pattern works
2. Get a working product we can actually use
3. Keep it simple enough to build in a weekend

## User Flow

```
1. User completes ride in MyWhoosh
2. User triggers sync (manual for MVP, auto later)
3. FitSync downloads FIT file from MyWhoosh
4. FitSync parses FIT file, extracts cycling metrics
5. FitSync uploads to Intervals.icu
6. FitSync generates AI summary
7. User sees activity with summary in dashboard
```

## Scope

### In Scope (MVP)
- [ ] Download FIT files from MyWhoosh (method TBD - see spike)
- [ ] Parse FIT file to extract cycling metrics
- [ ] Upload FIT file to Intervals.icu
- [ ] Store activity metadata in Table Storage
- [ ] Store original FIT file in Blob Storage
- [ ] Generate AI summary using Azure OpenAI
- [ ] Simple Blazor dashboard showing activities
- [ ] Manual "sync now" button

### Out of Scope (Later)
- Automatic polling/webhooks
- Multiple users (single-user for now)
- Other sources (Garmin, Wahoo, etc.)
- Other sinks (Strava, TrainingPeaks, etc.)
- CLI tool
- Mobile-optimized UI

## Acceptance Criteria

### Core Sync
- [ ] Can authenticate with MyWhoosh (however that works)
- [ ] Can list available activities from MyWhoosh
- [ ] Can download FIT file for an activity
- [ ] FIT file is stored in Blob Storage with provenance
- [ ] Duplicate detection by file hash (don't re-download same file)

### FIT Parsing
- [ ] Extract: start time, duration, distance
- [ ] Extract: avg/max power, normalized power
- [ ] Extract: avg/max heart rate
- [ ] Extract: avg/max cadence
- [ ] Extract: TSS, IF (if available, or calculate)
- [ ] Extract: total ascent/descent
- [ ] Handle missing fields gracefully (indoor rides have no GPS)

### Intervals.icu Upload
- [ ] Authenticate with Intervals.icu API key
- [ ] Upload FIT file via API
- [ ] Handle duplicate detection (Intervals.icu may reject)
- [ ] Store sync status (pending/success/failed)

### AI Summary
- [ ] Generate 2-3 sentence summary after successful sync
- [ ] Include: workout type, key metrics, one observation
- [ ] Store summary with activity

### Dashboard
- [ ] List recent activities (last 20)
- [ ] Show: date, duration, distance, power, TSS
- [ ] Show: sync status for Intervals.icu
- [ ] Show: AI summary
- [ ] "Sync Now" button to trigger manual sync
- [ ] Settings page to enter API keys

## Technical Design

### Project Structure
```
FitSync.sln
├── FitSync.AppHost/
├── FitSync.ServiceDefaults/
├── FitSync.Web/              # Blazor dashboard
├── FitSync.Api/              # REST API
├── FitSync.Worker/           # Background processing
└── src/FitSync.Core/         # Shared logic
```

### Data Flow
```
[Sync Button] 
     │
     ▼
[API: POST /api/sync]
     │
     ▼
[Queue: sync-jobs]
     │
     ▼
[Worker: SyncJob]
     ├─► [MyWhoosh: List new activities]
     ├─► [MyWhoosh: Download FIT files]
     ├─► [Blob: Store FIT file]
     ├─► [Parse: Extract metrics]
     ├─► [Table: Store activity]
     ├─► [Intervals.icu: Upload FIT]
     ├─► [Azure OpenAI: Generate summary]
     └─► [Table: Update with summary]
```

### Configuration
```json
{
  "MyWhoosh": {
    "// TBD based on spike": ""
  },
  "IntervalsIcu": {
    "ApiKey": "xxx",
    "AthleteId": "i12345"
  },
  "AzureOpenAI": {
    "Endpoint": "https://xxx.openai.azure.com",
    "DeploymentName": "gpt-4o",
    "ApiKey": "xxx"
  }
}
```

### Key Interfaces

```csharp
public interface IActivitySource
{
    Task<List<ActivityReference>> ListNewActivitiesAsync(DateTime since);
    Task<Stream> DownloadFitFileAsync(string activityId);
}

public interface IActivitySink
{
    Task<UploadResult> UploadAsync(Stream fitFile, ActivityMetadata metadata);
}

public interface IFitParser
{
    CyclingActivity Parse(Stream fitFile);
}

public interface IWorkoutAnalyzer
{
    Task<string> GenerateSummaryAsync(CyclingActivity activity);
}
```

### Activity Entity (Table Storage)

```csharp
public class ActivityEntity : ITableEntity
{
    public string PartitionKey { get; set; }  // "default" (single user)
    public string RowKey { get; set; }        // ULID (time-sortable)
    
    // Timestamps
    public DateTime StartTime { get; set; }
    public int DurationSeconds { get; set; }
    
    // Cycling Metrics
    public double DistanceMeters { get; set; }
    public int? AvgPowerWatts { get; set; }
    public int? MaxPowerWatts { get; set; }
    public int? NormalizedPower { get; set; }
    public int? AvgHeartRate { get; set; }
    public int? MaxHeartRate { get; set; }
    public int? AvgCadence { get; set; }
    public int? Tss { get; set; }
    public double? IntensityFactor { get; set; }
    public int? TotalAscent { get; set; }
    
    // Source
    public string SourcePlatform { get; set; }     // "mywhoosh"
    public string SourceActivityId { get; set; }
    public string FitFileBlobPath { get; set; }
    public string FitFileHash { get; set; }
    
    // Sink Status
    public string IntervalsIcuStatus { get; set; }  // "pending", "synced", "failed"
    public string? IntervalsIcuActivityId { get; set; }
    public string? IntervalsIcuError { get; set; }
    
    // AI
    public string? AiSummary { get; set; }
    public DateTime? AiSummaryGeneratedAt { get; set; }
    
    // Table Storage required
    public ETag ETag { get; set; }
    public DateTimeOffset? Timestamp { get; set; }
}
```

## Test Cases

### Happy Path
- [ ] New activity in MyWhoosh → appears in dashboard with summary
- [ ] Activity shows as synced to Intervals.icu
- [ ] FIT file is preserved in Blob Storage

### Edge Cases
- [ ] Activity already synced (duplicate) → skip gracefully
- [ ] MyWhoosh auth expired → show error, allow re-auth
- [ ] Intervals.icu upload fails → mark as failed, allow retry
- [ ] FIT file has no power data → still works, metrics show as null
- [ ] AI service unavailable → activity syncs, summary generated later

## Dependencies

- **Spike: MyWhoosh FIT Access** - Need to know how to get files first!
- Azure subscription with Storage Account
- Azure OpenAI deployment (or skip AI for v0.1)
- Intervals.icu account with API key

## Estimates

| Task | Estimate |
|------|----------|
| Spike: MyWhoosh access | 2-4 hours |
| Project scaffolding | 1 hour |
| MyWhoosh source adapter | 2-4 hours (depends on spike) |
| FIT parser integration | 2 hours |
| Intervals.icu sink | 2 hours |
| Table/Blob storage | 2 hours |
| Basic Blazor dashboard | 3 hours |
| AI summary integration | 2 hours |
| Testing & polish | 3 hours |
| **Total** | **~20 hours** |

## Open Questions

- [ ] How do we get FIT files from MyWhoosh? (SPIKE REQUIRED)
- [ ] Do we need to calculate TSS/IF or does the FIT file contain it?
- [ ] What FTP value to use for IF calculation? (User setting?)

## Definition of Done

1. Can sync a ride from MyWhoosh to Intervals.icu
2. Activity visible in dashboard with metrics and AI summary
3. Original FIT file preserved
4. Works for at least 10 consecutive syncs without errors
