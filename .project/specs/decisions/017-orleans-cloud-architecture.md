# ADR-017: Orleans Cloud Architecture

## Status
**Superseded by [ADR-019: Serverless Multi-Tenant Architecture](019-serverless-multi-tenant-architecture.md)**

> **Reason for supersession**: Orleans requires always-on infrastructure ($100+/month minimum) which is excessive for a stealth-mode SaaS with ~100 users. Durable Functions provides similar actor-like state management with scale-to-zero economics (~$5/month).

## Context

FitBridge needs a cloud backend that can:
1. Handle multiple users with isolated state (tokens, sync history)
2. Schedule and execute syncs on behalf of users
3. Scale efficiently (mostly idle, bursts during sync)
4. Store FIT files durably
5. Run LLM analysis on activities

We previously considered Azure Functions (ADR-006) but need more sophisticated state management and scheduling than Functions easily provides.

## Decision

**Use Microsoft Orleans on Azure Container Apps for the backend, with each user represented as a grain.**

### Why Orleans?

| Requirement | Orleans Solution |
|-------------|------------------|
| Per-user state | Virtual actor (grain) per user |
| Token storage | Grain state persisted to Azure Storage |
| Sync scheduling | Grain timers / reminders |
| Horizontal scale | Silo auto-scaling on Container Apps |
| Long-running syncs | Grain activations survive across requests |
| Fault tolerance | Automatic grain reactivation |

### Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  Azure Container Apps Environment                                            │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────────┐│
│  │ Orleans Silo (auto-scaled)                                               ││
│  │                                                                          ││
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐       ││
│  │  │ UserGrain   │ │ UserGrain   │ │ UserGrain   │ │ UserGrain   │       ││
│  │  │ user-001    │ │ user-002    │ │ user-003    │ │ user-004    │       ││
│  │  └──────┬──────┘ └──────┬──────┘ └──────┬──────┘ └──────┬──────┘       ││
│  │         │               │               │               │               ││
│  │  ┌──────▼──────┐ ┌──────▼──────┐ ┌──────▼──────┐ ┌──────▼──────┐       ││
│  │  │ ProviderGrain│ │ ProviderGrain│ │ ProviderGrain│ │ ProviderGrain│   ││
│  │  │ user-001/tp │ │ user-002/mw │ │ user-003/tp │ │ user-004/mw │       ││
│  │  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘       ││
│  └─────────────────────────────────────────────────────────────────────────┘│
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────────┐│
│  │ Shared Infrastructure                                                    ││
│  │                                                                          ││
│  │  Azure Storage    Key Vault       Azure OpenAI     SignalR               ││
│  │  - Grain state    - Tokens        - Analysis       - Real-time updates   ││
│  │  - FIT files      - API keys      - Summaries      - Sync progress       ││
│  │  - Reminders      - Secrets       - Insights                             ││
│  └─────────────────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────────────┘
```

### Grain Design

```fsharp
/// User grain - orchestrates all user operations
[<GenerateSerializer>]
type IUserGrain =
    inherit IGrainWithStringKey  // Key = user ID (from Entra)
    
    // Provider connections
    abstract member ConnectProvider: provider: string * token: string -> Task<unit>
    abstract member DisconnectProvider: provider: string -> Task<unit>
    abstract member GetConnectedProviders: unit -> Task<ProviderStatus list>
    
    // Sync operations
    abstract member TriggerSync: provider: string -> Task<SyncResult>
    abstract member GetSyncHistory: limit: int -> Task<SyncRecord list>
    abstract member SetSyncSchedule: provider: string * interval: TimeSpan -> Task<unit>
    
    // Activity queries
    abstract member GetActivities: startDate: DateTime * endDate: DateTime -> Task<Activity list>
    abstract member GetActivity: id: string -> Task<Activity option>

/// Provider grain - handles sync for one user+provider combination
[<GenerateSerializer>]
type IProviderGrain =
    inherit IGrainWithStringKey  // Key = "userId/provider" e.g. "user-001/trainingpeaks"
    
    abstract member SetToken: token: string -> Task<unit>
    abstract member GetTokenStatus: unit -> Task<TokenStatus>
    abstract member SyncNow: unit -> Task<SyncResult>
    abstract member GetLastSync: unit -> Task<SyncRecord option>

/// Analysis grain - handles LLM analysis for activities
[<GenerateSerializer>]
type IAnalysisGrain =
    inherit IGrainWithStringKey  // Key = activity ID
    
    abstract member Analyze: unit -> Task<AnalysisResult>
    abstract member GetInsights: unit -> Task<Insight list>
```

### Grain State

```fsharp
[<GenerateSerializer>]
type UserGrainState = {
    UserId: string
    Email: string
    Providers: Map<string, ProviderConnection>
    SyncSchedules: Map<string, TimeSpan>
    CreatedAt: DateTimeOffset
    LastActiveAt: DateTimeOffset
}

[<GenerateSerializer>]
type ProviderGrainState = {
    Provider: string
    TokenHash: string  // Actual token in Key Vault
    TokenExpiresAt: DateTimeOffset option
    LastSyncAt: DateTimeOffset option
    LastSyncStatus: SyncStatus
    SyncedActivityIds: Set<string>  // For deduplication
}
```

### Sync Flow

```fsharp
// In ProviderGrain
member this.SyncNow() = task {
    let! token = keyVault.GetSecretAsync($"token-{grainKey}")
    
    // Get existing activities from sink for dedup
    let! existing = intervalsClient.ListActivities(startDate, endDate)
    let existingSink = existing |> List.map Intervals.toSinkActivity
    
    // Fetch from source
    let! sourceWorkouts = 
        match state.Provider with
        | "trainingpeaks" -> TrainingPeaks.listWorkouts token athleteId startDate endDate
        | "mywhoosh" -> MyWhoosh.listActivities token whooshId
        | _ -> Task.FromResult []
    
    // Detect duplicates
    let toSync = 
        sourceWorkouts
        |> List.map (fun w -> w, DuplicateDetection.findDuplicate (toMetadata w) existingSink)
        |> List.filter (fun (_, dup) -> dup = NoDuplicate)
        |> List.map fst
    
    // Download and upload
    let mutable synced = 0
    for workout in toSync do
        let! activity = fetchActivity token workout
        match activity with
        | Some a ->
            let! success = Intervals.uploadFitFile a.FitData
            if success then synced <- synced + 1
        | None -> ()
    
    // Update state
    state <- { state with 
        LastSyncAt = Some DateTimeOffset.UtcNow
        LastSyncStatus = SyncStatus.Success
        SyncedActivityIds = state.SyncedActivityIds |> Set.union (toSync |> List.map id |> Set.ofList)
    }
    
    return { SyncedCount = synced; TotalCount = sourceWorkouts.Length }
}
```

### Scheduling with Reminders

```fsharp
// Orleans reminders persist across silo restarts
member this.SetSyncSchedule(interval: TimeSpan) = task {
    do! this.RegisterOrUpdateReminder("sync", interval, interval)
    state <- { state with SyncSchedule = Some interval }
}

// Called by Orleans when reminder fires
member this.ReceiveReminder(reminderName, status) = task {
    if reminderName = "sync" then
        do! this.SyncNow() |> ignore
}
```

### Container Apps Configuration

```yaml
# container-app.yaml
properties:
  configuration:
    ingress:
      external: true
      targetPort: 8080
    secrets:
      - name: storage-connection
        keyVaultUrl: https://fitbridge-kv.vault.azure.net/secrets/storage-connection
  template:
    containers:
      - name: fitbridge-silo
        image: fitbridge.azurecr.io/silo:latest
        resources:
          cpu: 0.5
          memory: 1Gi
        env:
          - name: ORLEANS_CLUSTER_ID
            value: fitbridge-prod
          - name: AZURE_STORAGE_CONNECTION
            secretRef: storage-connection
    scale:
      minReplicas: 1
      maxReplicas: 10
      rules:
        - name: cpu-scaling
          custom:
            type: cpu
            metadata:
              type: Utilization
              value: "70"
```

## Alternatives Considered

### Option A: Azure Functions + Durable Functions
- **Pros**: Serverless, pay-per-use
- **Cons**: Durable entities less mature than Orleans, cold starts
- **Verdict**: Orleans is more natural for actor-based user state

### Option B: Traditional Web API + Database
- **Pros**: Simple, well-understood
- **Cons**: Manual state management, harder to scale per-user isolation
- **Verdict**: Orleans grains are a better abstraction for this domain

### Option C: Azure Kubernetes Service (AKS)
- **Pros**: Full control
- **Cons**: Operational overhead, more expensive for small scale
- **Verdict**: Container Apps is simpler and sufficient

## Consequences

### Positive
- Natural per-user isolation with grains
- Built-in persistence and fault tolerance
- Reminders for reliable scheduling
- Scales from 1 to many users efficiently
- F# works great with Orleans

### Negative
- Learning curve for Orleans concepts
- Debugging distributed grains is harder than monolith
- Need to design grain boundaries carefully

### Cost Estimate (Small Scale)

| Resource | Config | Est. Monthly Cost |
|----------|--------|-------------------|
| Container Apps | 1 replica, 0.5 CPU, 1GB | ~$15 |
| Azure Storage | 10GB + transactions | ~$5 |
| Key Vault | 1000 operations/mo | ~$1 |
| Azure OpenAI | 100K tokens/mo | ~$5 |
| **Total** | | **~$26/mo** |

Scales up as users increase, but base cost is low.
