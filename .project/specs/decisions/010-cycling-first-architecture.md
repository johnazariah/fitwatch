# ADR-010: Cycling-First, Sport-Sharded Architecture

## Status
Accepted

## Context
We need to decide whether to build a generic multi-sport platform or focus on a single sport initially.

## Decision

**Build for cycling only in MVP. Design for sport-sharding later.**

### Rationale

| Factor | Multi-Sport | Cycling-Only |
|--------|-------------|--------------|
| Data model complexity | High (run/swim/etc. have different fields) | Low (power, cadence, HR, GPS) |
| LLM prompts | Generic, less useful | Cycling coach persona, specific advice |
| Metrics | TSS varies by sport | Power-based TSS is well-defined |
| Target users | Diffuse | Clear: cyclists who care about data |
| Development speed | Slower | Faster to MVP |
| Competition | Crowded generic space | Focused niche |

### What "Cycling-Only" Means

**In scope:**
- Road cycling
- Indoor cycling (Zwift, MyWhoosh, trainer)
- Gravel/mountain biking
- Virtual racing

**Out of scope (for now):**
- Running
- Swimming
- Triathlon (multisport files)
- Strength training
- Other sports

### Sport-Sharding for Future

Design the architecture so we *could* add running later:

```
fitsync.cycling.app    ← MVP
fitsync.running.app    ← Future
fitsync.triathlon.app  ← Future (combines both)
```

Each "shard" has:
- Sport-specific data model
- Sport-specific AI prompts
- Sport-specific metrics
- Shared infrastructure (auth, storage patterns)

### Implementation Notes

```csharp
// Activity model is cycling-specific
public class CyclingActivity
{
    // Universal
    public Guid Id { get; set; }
    public DateTime StartTime { get; set; }
    public TimeSpan Duration { get; set; }
    public double DistanceMeters { get; set; }
    
    // Cycling-specific (not optional, expected)
    public int? AveragePowerWatts { get; set; }
    public int? NormalizedPowerWatts { get; set; }
    public int? AverageCadenceRpm { get; set; }
    public double? IntensityFactor { get; set; }
    public int? TrainingStressScore { get; set; }
    public int? FunctionalThresholdPower { get; set; }  // User's FTP at time
    
    // Not "speed" - cycling cares about power
    public double? AverageSpeedKph { get; set; }  // Secondary metric
}
```

```csharp
// AI prompts are cycling-specific
public static class CyclingPrompts
{
    public const string WorkoutSummary = """
        You are an experienced cycling coach analyzing a ride.
        Focus on power-based metrics (NP, IF, TSS) over speed.
        Use cycling terminology: intervals, threshold, sweet spot, Z2, etc.
        
        Workout data:
        {{$workoutData}}
        """;
}
```

## Consequences

### Positive
- Faster MVP
- Better AI advice (cycling-specific)
- Clearer value proposition
- Simpler data model

### Negative
- Excludes runners initially
- Multisport athletes may look elsewhere
- Need to refactor for multi-sport later

### Marketing
- "The cyclist's data freedom platform"
- "Own your watts"
- "Break free from Strava — for cyclists"

## Related Decisions
- ADR-011: Cost-Optimized Storage (Table/Queue/Blob)
