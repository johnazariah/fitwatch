# ADR-011: Cost-Optimized Azure Storage

## Status
Accepted

## Context
We need to choose storage solutions that are:
- Low cost for a hobby/side project
- Sufficient for the workload (not over-engineered)
- Simple to operate

Cosmos DB was initially considered but is expensive (~$25/month minimum for provisioned, or variable for serverless).

## Decision

**Use Azure Table Storage + Queue Storage + Blob Storage instead of Cosmos DB + Service Bus.**

### Cost Comparison

| Service | Cosmos DB / Service Bus | Table / Queue / Blob |
|---------|------------------------|----------------------|
| Database | ~$25-50/month (serverless) | ~$0.50/month (Table) |
| Messaging | ~$10/month (Service Bus Basic) | ~$0.01/month (Queue) |
| Files | Same (Blob) | Same (Blob) |
| **Total** | **~$35-60/month** | **~$1-5/month** |

### Storage Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                     Azure Storage Account                           │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌─────────────────────────────────────────────────────────────┐    │
│  │                    Blob Storage                              │    │
│  │  fit-files/                                                  │    │
│  │    {userId}/{activityId}.fit    ← Original FIT files        │    │
│  │  exports/                                                    │    │
│  │    {userId}/{date}-bulk.zip     ← User exports              │    │
│  └─────────────────────────────────────────────────────────────┘    │
│                                                                      │
│  ┌─────────────────────────────────────────────────────────────┐    │
│  │                    Table Storage                             │    │
│  │                                                               │    │
│  │  Activities                                                   │    │
│  │    PK: {userId}                                              │    │
│  │    RK: {activityId}                                          │    │
│  │    + summary fields, provenance, sync status                 │    │
│  │                                                               │    │
│  │  Connections                                                  │    │
│  │    PK: {userId}                                              │    │
│  │    RK: {provider}                                            │    │
│  │    + connection status (tokens in Key Vault)                 │    │
│  │                                                               │    │
│  │  SyncStatus                                                   │    │
│  │    PK: {userId}                                              │    │
│  │    RK: {activityId}_{sinkId}                                 │    │
│  │    + status, lastAttempt, error                              │    │
│  │                                                               │    │
│  │  UserSettings                                                 │    │
│  │    PK: "settings"                                            │    │
│  │    RK: {userId}                                              │    │
│  │    + preferences JSON                                        │    │
│  └─────────────────────────────────────────────────────────────┘    │
│                                                                      │
│  ┌─────────────────────────────────────────────────────────────┐    │
│  │                    Queue Storage                             │    │
│  │                                                               │    │
│  │  fit-parse-queue      ← New FIT file needs parsing          │    │
│  │  sync-queue           ← Activity needs uploading to sink    │    │
│  │  analysis-queue       ← Activity needs AI summary           │    │
│  └─────────────────────────────────────────────────────────────┘    │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### Table Storage Schema

#### Activities Table

```csharp
public class ActivityEntity : ITableEntity
{
    // Keys
    public string PartitionKey { get; set; }  // UserId
    public string RowKey { get; set; }        // ActivityId (ULID for sortability)
    public ETag ETag { get; set; }
    public DateTimeOffset? Timestamp { get; set; }
    
    // Activity Summary
    public DateTime StartTime { get; set; }
    public int DurationSeconds { get; set; }
    public double DistanceMeters { get; set; }
    public int? AveragePowerWatts { get; set; }
    public int? NormalizedPowerWatts { get; set; }
    public int? MaxPowerWatts { get; set; }
    public int? AverageHeartRate { get; set; }
    public int? MaxHeartRate { get; set; }
    public int? AverageCadence { get; set; }
    public double? IntensityFactor { get; set; }
    public int? TrainingStressScore { get; set; }
    public int? KilojoulesBurned { get; set; }
    public int? TotalAscent { get; set; }
    
    // Metadata
    public string Title { get; set; }
    public string ActivityType { get; set; }  // "outdoor_ride", "virtual_ride", etc.
    public string? Description { get; set; }
    
    // Provenance (JSON serialized for complex object)
    public string ProvenanceJson { get; set; }
    
    // AI Analysis
    public string? AiSummary { get; set; }
    public DateTime? AiSummaryGeneratedAt { get; set; }
    
    // Sync Status (denormalized for quick display)
    public string SyncStatusJson { get; set; }  // {"strava": "synced", "intervals": "pending"}
    
    // File Reference
    public string FitFileBlobPath { get; set; }
    public string FitFileHash { get; set; }
}
```

#### Why Table Storage Works Here

| Requirement | Table Storage Fit |
|-------------|-------------------|
| Query by user | ✅ PartitionKey = userId |
| List recent activities | ✅ RowKey = ULID (time-sortable) |
| Get single activity | ✅ Point query by PK+RK |
| Filter by date range | ✅ RowKey range query |
| Complex queries | ⚠️ Limited, but we don't need many |
| Transactions | ⚠️ Same partition only (fine for us) |

### Queue Storage vs Service Bus

| Feature | Service Bus | Queue Storage |
|---------|-------------|---------------|
| Cost | ~$10/month | ~$0.01/month |
| Message size | 256KB-100MB | 64KB |
| Dead letter | ✅ Built-in | ⚠️ Manual |
| Sessions/ordering | ✅ | ❌ |
| Peek | ✅ | ✅ |
| Our needs | Overkill | Sufficient |

For our workload (simple job queues), Queue Storage is plenty.

### Handling Queue Storage Limitations

**64KB message limit:**
```csharp
// Don't put data in queue - just references
public class ParseFitMessage
{
    public string UserId { get; set; }
    public string ActivityId { get; set; }
    public string BlobPath { get; set; }  // Worker fetches from Blob
}
```

**Dead letter (manual):**
```csharp
// After N failures, move to poison queue
if (dequeueCount > 5)
{
    await _poisonQueue.SendMessageAsync(message);
    await _queue.DeleteMessageAsync(message.MessageId, message.PopReceipt);
    return;
}
```

### Time-Series Data (Records)

The one thing Table Storage isn't great for is time-series data (the per-second records from FIT files). Options:

| Approach | Pros | Cons |
|----------|------|------|
| **A. Store in Blob as JSON** | Cheap, simple | Must load entire file |
| **B. Store sampled in Table** | Queryable | Limited to ~1MB entity |
| **C. Don't store at all** | Cheapest | Re-parse FIT file when needed |
| **D. Store in separate Table** | Queryable | Many rows per activity |

**Decision: Option A (Blob) + Option C (re-parse)**

- Store summary in Table (what users see 99% of time)
- Store original FIT file in Blob (source of truth)
- For charts/detailed view, re-parse FIT file on demand
- Cache parsed data in memory/local for session

```csharp
public async Task<List<Record>> GetActivityRecords(string activityId)
{
    // Check if already parsed this session
    if (_cache.TryGet(activityId, out var records))
        return records;
    
    // Fetch FIT file and parse
    var fitBlob = await _blobClient.DownloadAsync($"fit-files/{activityId}.fit");
    var parsed = await _fitParser.ParseAsync(fitBlob);
    
    _cache.Set(activityId, parsed.Records, TimeSpan.FromMinutes(30));
    return parsed.Records;
}
```

### Aspire Integration

```csharp
// AppHost/Program.cs
var builder = DistributedApplication.CreateBuilder(args);

// Use Azure Storage (emulated locally with Azurite)
var storage = builder.AddAzureStorage("storage")
    .RunAsEmulator();

var blobs = storage.AddBlobs("fit-files");
var tables = storage.AddTables("metadata");
var queues = storage.AddQueues("jobs");

var api = builder.AddProject<Projects.FitSync_Api>("api")
    .WithReference(blobs)
    .WithReference(tables);

var worker = builder.AddProject<Projects.FitSync_Worker>("worker")
    .WithReference(blobs)
    .WithReference(tables)
    .WithReference(queues);
```

## Consequences

### Positive
- ~95% cost reduction vs Cosmos DB + Service Bus
- Simpler to understand and debug
- Azurite provides perfect local emulation
- No capacity planning (fully serverless)

### Negative
- Limited query capabilities (no SQL-like queries)
- No automatic indexing (design around partition/row keys)
- 1MB entity limit (use Blob for large data)
- No change feed (poll or use Event Grid)

### When to Reconsider Cosmos DB
- Need complex queries across partitions
- Need global distribution
- Need change feed for real-time sync
- Scale to >10,000 users

## Cost Projection

| Scenario | Monthly Cost |
|----------|--------------|
| 1 user, 10 activities/week | < $1 |
| 10 users, 50 activities/week | ~$2 |
| 100 users, 500 activities/week | ~$5-10 |
| 1000 users | Consider Cosmos |

## Related Decisions
- ADR-006: Azure + .NET Aspire Architecture (updated)
- ADR-010: Cycling-First Architecture
