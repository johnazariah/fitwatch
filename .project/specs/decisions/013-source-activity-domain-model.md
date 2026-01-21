# ADR-013: Source Activity Domain Model

## Status
Accepted

## Context

When syncing FIT files between platforms, we initially treated activities as just byte arrays. This led to problems:

1. No metadata for duplicate detection (had to re-parse FIT files)
2. No visibility into what was being synced (just "downloading...")
3. No consistent way to handle activities from different sources
4. Each source API returns different field names and formats

We need a unified domain model that captures activity metadata alongside the raw FIT data.

## Decision Drivers

- **Consistency**: Same representation regardless of source platform
- **Duplicate detection**: Metadata must be available without parsing FIT
- **Observability**: Clear logging of what's being synced
- **Extensibility**: Easy to add new sources (Wahoo, Garmin, Zwift)
- **Simplicity**: Don't over-engineer for hypothetical requirements

## Decision

### Core Domain Types

```fsharp
/// Source platform an activity came from
type Source = MyWhoosh | TrainingPeaks | Wahoo | Garmin | Zwift

/// Sink/destination platform  
type Sink = IntervalsIcu | Strava | TrainingPeaks

/// Activity type (cycling-focused for MVP)
type ActivityType = Ride | VirtualRide | Run | VirtualRun | Other of string

/// Core metadata - common across all sources
type ActivityMetadata = {
    SourceId: string              // ID in source system
    Source: Source
    Title: string option
    ActivityType: ActivityType
    StartTime: DateTimeOffset
    Duration: TimeSpan option     // Moving time
    Distance: float option        // Meters
    TotalWork: float option       // kJ
    AveragePower: float option    // Watts
    NormalizedPower: float option
    TSS: float option
}

/// Downloaded activity with FIT file
type SourceActivity = {
    Metadata: ActivityMetadata
    FitData: byte[]
}

/// Existing activity in destination (for matching)
type SinkActivity = {
    SinkId: string
    Sink: Sink
    StartTime: DateTimeOffset
    Duration: TimeSpan option
    Distance: float option
    Name: string option
    Source: string option
}
```

### Source Adapters

Each source module provides a function to convert its API response to the domain model:

```fsharp
// TrainingPeaks.fs
let toMetadata (workout: Workout) : ActivityMetadata = ...
let fetchActivity (token: string) (athleteId: int) (workout: Workout) : Async<SourceActivity option> = ...

// MyWhoosh.fs  
let toMetadata (activity: Activity) : ActivityMetadata = ...
let fetchActivity (token: string) (whooshId: string) (activityId: int64) : Async<SourceActivity option> = ...
```

### Sink Adapters

Each sink module provides conversion from its API response:

```fsharp
// Intervals.fs
let toSinkActivity (activity: Activity) : SinkActivity = ...
let listSinkActivities (apiKey: string) (athleteId: string) (...) : Async<Result<SinkActivity list, string>> = ...
```

## Alternatives Considered

### Option A: Parse FIT Files for Metadata
- **Pros**: Single source of truth
- **Cons**: Slow (must download entire file), complex parsing
- **Verdict**: Rejected - API metadata is faster and sufficient

### Option B: Platform-Specific Types Throughout
- **Pros**: No translation layer
- **Cons**: Duplicate logic everywhere, hard to add sources
- **Verdict**: Rejected - doesn't scale

### Option C: Generic Dictionary-Based Model
- **Pros**: Infinitely flexible
- **Cons**: No type safety, easy to misuse
- **Verdict**: Rejected - F# strength is in types

## Consequences

### Positive
- Type-safe, self-documenting code
- Easy to add new sources (implement `toMetadata` + `fetchActivity`)
- Duplicate detection works on metadata without FIT parsing
- Clear separation: domain model vs API types
- Good logging/display with `Display` module helpers

### Negative
- Some data loss in translation (source-specific fields not captured)
- Must maintain mapping for each source

### Fields Intentionally Omitted
- GPS data (would require FIT parsing, not needed for duplicate detection)
- HR zones, power zones (can be derived from FIT later)
- Equipment (bike, shoes - nice to have, not MVP)

### Future Enhancements
- Add `RouteSignature: byte[] option` for outdoor activity matching
- Add `Equipment: Equipment option` for bike/shoe tracking
- Add `Weather: Weather option` from external APIs
