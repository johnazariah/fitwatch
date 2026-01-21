# ADR-015: F# for CLI and Domain Logic

## Status
Accepted

## Context

We initially built the FitSync CLI in C#, then rewrote it in F#. This documents why F# is the better choice for this domain.

## Decision Drivers

- **Domain modeling**: Fitness data has complex types (optional fields, discriminated unions)
- **Data transformation**: ETL pipeline from source → domain → sink
- **Correctness**: Type system should catch errors at compile time
- **Conciseness**: Less boilerplate = faster iteration during spikes
- **Interop**: Must work with .NET ecosystem (Azure, existing libraries)

## Decision

**Use F# for CLI, domain logic, and data transformation. Keep C# available for ASP.NET/Blazor if needed.**

### Why F# Fits This Domain

#### 1. Discriminated Unions for Activity Types

```fsharp
type Source = MyWhoosh | TrainingPeaks | Wahoo | Garmin | Zwift
type ActivityType = Ride | VirtualRide | Run | VirtualRun | Other of string
type DuplicateMatch = 
    | NoDuplicate 
    | ProbableDuplicate of SinkActivity * confidence: float 
    | ExactDuplicate of SinkActivity
```

C# equivalent would be verbose abstract classes or enums with switch statements.

#### 2. Option Types for Nullable Fields

```fsharp
type ActivityMetadata = {
    Title: string option          // Might not have a title
    Duration: TimeSpan option     // Might be missing
    TSS: float option             // Not all platforms calculate
}

// Safe access
let title = activity.Title |> Option.defaultValue "(untitled)"
```

No null reference exceptions, compiler enforces handling of missing data.

#### 3. Pipeline Operators for Data Transformation

```fsharp
workouts
|> List.filter (fun w -> w.Completed = Some true)
|> List.map TrainingPeaks.toMetadata
|> List.map (fun m -> m, DuplicateDetection.findDuplicate m existing)
|> List.filter (fun (_, dup) -> dup = NoDuplicate)
|> List.map fst
```

Clear, readable data flow from source to destination.

#### 4. Async Computation Expressions

```fsharp
let syncActivity workout = async {
    match! TrainingPeaks.fetchActivity token athleteId workout with
    | Some activity ->
        let! (success, _) = Intervals.uploadFitFile apiKey athleteId activity.FitData
        return success
    | None -> return false
}
```

Async is first-class, not bolted on.

#### 5. Pattern Matching for API Responses

```fsharp
match! httpClient.GetAsync(url) with
| response when response.IsSuccessStatusCode ->
    let! json = response.Content.ReadAsStringAsync()
    return Ok (deserialize json)
| response ->
    return Error $"HTTP {int response.StatusCode}"
```

### Project Structure

```
FitSync.Cli.FSharp/
├── Domain.fs          # Core types (Source, Sink, ActivityMetadata, DuplicateDetection)
├── Config.fs          # Configuration loading
├── MyWhoosh.fs        # MyWhoosh API client
├── TrainingPeaks.fs   # TrainingPeaks API client  
├── Intervals.fs       # Intervals.icu API client
└── Program.fs         # CLI commands
```

F# requires explicit file ordering, which enforces proper dependency direction.

## Alternatives Considered

### Option A: C# with Records and Pattern Matching
- **Pros**: C# 12 has records, pattern matching is improving
- **Cons**: Still verbose, nullable handling not as clean
- **Verdict**: Rejected - F# is more natural for this domain

### Option B: TypeScript/Node.js
- **Pros**: Fast iteration, JSON-native
- **Cons**: No .NET interop, weaker type system at runtime
- **Verdict**: Rejected - want .NET ecosystem for Azure integration

### Option C: Rust
- **Pros**: Performance, strong types, great for CLIs
- **Cons**: Learning curve, no .NET interop
- **Verdict**: Rejected - over-engineered for this use case

## Consequences

### Positive
- 40% less code than C# version for same functionality
- Compiler catches more errors (especially null handling)
- Domain model is self-documenting
- Easy to extend with new sources/sinks
- Full .NET 10 compatibility

### Negative
- Smaller talent pool if hiring
- Some Azure libraries have C#-first examples
- File ordering can be frustrating initially
- IDE support slightly behind C# (but VS Code + Ionide is good)

### Interop Strategy

If we need C# for specific components (e.g., Blazor UI):
1. Keep domain types in F# project
2. Reference F# project from C# project
3. F# records are visible as C# classes with properties
