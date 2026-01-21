# ADR-012: Duplicate Detection Strategy

## Status
Accepted

## Context

When syncing activities between platforms (TrainingPeaks → Intervals.icu, MyWhoosh → Intervals.icu), we frequently encounter activities that already exist in the destination. Without proper duplicate detection:

1. Users get duplicate activities cluttering their training log
2. Analytics become inaccurate (double-counted TSS, volume, etc.)
3. Manual cleanup is tedious and error-prone

We discovered during the TrainingPeaks spike that most workouts already existed in Intervals.icu (synced via Garmin/Wahoo). We need a robust strategy to detect and handle duplicates.

## Decision Drivers

- **Accuracy**: Minimize false positives (incorrectly marking distinct activities as duplicates)
- **Recall**: Minimize false negatives (missing actual duplicates)
- **Performance**: Detection should be fast even with thousands of activities
- **Transparency**: Users should understand why duplicates were detected
- **Source diversity**: Same activity may have different timestamps/metadata from different sources

## Decision

### Multi-Factor Confidence-Based Matching

Use a weighted confidence score based on multiple factors, not just a single field match.

```fsharp
type DuplicateMatch =
    | NoDuplicate
    | ProbableDuplicate of SinkActivity * confidence: float
    | ExactDuplicate of SinkActivity
```

### Matching Factors

| Factor | Weight | Tolerance | Rationale |
|--------|--------|-----------|-----------|
| Start time | 0.5 (required) | ±5 minutes | Clock drift between devices |
| Duration | 0.3 | ±10% | Different pause handling |
| Distance | 0.2 | ±10% | GPS accuracy, calibration |

### Confidence Thresholds

| Score | Classification | Action |
|-------|---------------|--------|
| ≥ 0.8 | ExactDuplicate | Skip upload |
| 0.5 - 0.8 | ProbableDuplicate | Skip upload, flag for review |
| < 0.5 | NoDuplicate | Upload |

### Implementation

```fsharp
let matchConfidence (source: ActivityMetadata) (sink: SinkActivity) =
    let mutable score = 0.0
    
    // Time match is essential gate
    if abs (source.StartTime - sink.StartTime).TotalMinutes < 5.0 then
        score <- 0.5
        
        // Duration adds confidence
        if durationWithinTolerance source.Duration sink.Duration 0.10 then
            score <- score + 0.3
        
        // Distance adds confidence  
        if distanceWithinTolerance source.Distance sink.Distance 0.10 then
            score <- score + 0.2
    
    score
```

## Alternatives Considered

### Option A: Exact Timestamp Match
- **Pros**: Simple, deterministic
- **Cons**: Fails with clock drift, timezone issues, different source precision
- **Verdict**: Rejected - too brittle

### Option B: Activity ID / External ID
- **Pros**: Deterministic if platforms share IDs
- **Cons**: IDs are platform-specific, not preserved across syncs
- **Verdict**: Rejected - not portable

### Option C: File Hash
- **Pros**: Exact duplicate detection
- **Cons**: Same activity from different sources produces different FIT files
- **Verdict**: Rejected - only works for exact same file

### Option D: Date-Only Match (Initial Implementation)
- **Pros**: Simple, catches most duplicates
- **Cons**: Multiple activities per day (AM/PM workouts) causes false positives
- **Verdict**: Rejected after testing - too coarse

## Consequences

### Positive
- Handles clock drift and timezone differences gracefully
- Multiple activities per day correctly distinguished
- Transparent confidence scores for debugging
- Extensible - can add more factors (power data, route signature, etc.)

### Negative
- More complex than simple matching
- Requires fetching metadata from both source and sink
- Thresholds may need tuning based on real-world data

### Future Enhancements
- **Route signature**: Hash of lat/lng points for outdoor activities
- **Power curve matching**: Compare 1s/5s/1m power for indoor activities
- **User feedback loop**: Learn from manual duplicate resolutions
