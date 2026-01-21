module FitSync.Domain

open System

/// Represents the source platform an activity came from
type Source =
    | MyWhoosh
    | TrainingPeaks
    | IGPSport
    | Wahoo
    | Garmin
    | Zwift

/// Represents the sink/destination platform
type Sink =
    | IntervalsIcu
    | Strava
    | TrainingPeaks

/// Activity type (cycling-focused for now)
type ActivityType =
    | Ride
    | VirtualRide
    | Run
    | VirtualRun
    | Other of string

/// Core activity metadata - common across all sources
type ActivityMetadata = {
    SourceId: string              // Unique ID in the source system
    Source: Source
    Title: string option
    ActivityType: ActivityType
    StartTime: DateTimeOffset
    Duration: TimeSpan option     // Moving time
    Distance: float option        // Meters
    TotalWork: float option       // kJ
    AveragePower: float option    // Watts
    NormalizedPower: float option // Watts
    TSS: float option
}

/// A downloaded activity with its FIT file and metadata
type SourceActivity = {
    Metadata: ActivityMetadata
    FitData: byte[]
}

/// An existing activity in a sink (for duplicate detection)
type SinkActivity = {
    SinkId: string                // Unique ID in the sink system
    Sink: Sink
    StartTime: DateTimeOffset
    Duration: TimeSpan option
    Distance: float option
    Name: string option
    Source: string option         // Where the sink thinks it came from (if known)
}

/// Result of comparing a source activity against existing sink activities
type DuplicateMatch =
    | NoDuplicate
    | ProbableDuplicate of SinkActivity * confidence: float
    | ExactDuplicate of SinkActivity

/// Duplicate detection strategies
module DuplicateDetection =
    
    /// Check if two activities are on the same day
    let sameDayMatch (source: ActivityMetadata) (sink: SinkActivity) =
        source.StartTime.Date = sink.StartTime.Date
    
    /// Check if two activities are likely the same based on timing (within 5 mins)
    let timeMatch (source: ActivityMetadata) (sink: SinkActivity) =
        let timeDiff = abs (source.StartTime - sink.StartTime).TotalMinutes
        timeDiff < 5.0  // Within 5 minutes
    
    /// Check duration similarity (within 10%)
    let durationMatch (source: ActivityMetadata) (sink: SinkActivity) =
        match source.Duration, sink.Duration with
        | Some sd, Some ed ->
            let diff = abs (sd.TotalSeconds - ed.TotalSeconds)
            let tolerance = sd.TotalSeconds * 0.1
            diff <= tolerance
        | _ -> false
    
    /// Check distance similarity (within 10%)
    let distanceMatch (source: ActivityMetadata) (sink: SinkActivity) =
        match source.Distance, sink.Distance with
        | Some sd, Some ed when sd > 0.0 && ed > 0.0 ->
            let diff = abs (sd - ed)
            let tolerance = sd * 0.1
            diff <= tolerance
        | _ -> false
    
    /// Calculate match confidence (0.0 to 1.0)
    /// Handles both precise timestamps and date-only sources (like iGPSport)
    let matchConfidence (source: ActivityMetadata) (sink: SinkActivity) =
        let mutable score = 0.0
        
        // Check for exact time match first
        if timeMatch source sink then
            score <- score + 0.5
            if durationMatch source sink then score <- score + 0.3
            if distanceMatch source sink then score <- score + 0.2
        // Fall back to same-day + duration + distance match (for date-only sources)
        elif sameDayMatch source sink && durationMatch source sink && distanceMatch source sink then
            score <- 0.9  // High confidence if same day with matching duration and distance
        elif sameDayMatch source sink && durationMatch source sink then
            score <- 0.7  // Medium-high if same day with matching duration
        elif sameDayMatch source sink && distanceMatch source sink then
            score <- 0.6  // Medium if same day with matching distance
        
        score
    
    /// Find duplicates for a source activity
    let findDuplicate (source: ActivityMetadata) (sinkActivities: SinkActivity list) : DuplicateMatch =
        let candidates = 
            sinkActivities
            |> List.map (fun sink -> sink, matchConfidence source sink)
            |> List.filter (fun (_, conf) -> conf > 0.0)
            |> List.sortByDescending snd
        
        match candidates with
        | [] -> NoDuplicate
        | (sink, conf) :: _ when conf >= 0.8 -> ExactDuplicate sink
        | (sink, conf) :: _ -> ProbableDuplicate (sink, conf)

/// Pretty print helpers
module Display =
    
    let sourceToString (s: Source) =
        match s with
        | Source.MyWhoosh -> "MyWhoosh"
        | Source.TrainingPeaks -> "TrainingPeaks"
        | Source.IGPSport -> "iGPSport"
        | Source.Wahoo -> "Wahoo"
        | Source.Garmin -> "Garmin"
        | Source.Zwift -> "Zwift"
    
    let sinkToString (s: Sink) =
        match s with
        | Sink.IntervalsIcu -> "Intervals.icu"
        | Sink.Strava -> "Strava"
        | Sink.TrainingPeaks -> "TrainingPeaks"
    
    let activityTypeToString = function
        | Ride -> "Ride"
        | VirtualRide -> "VirtualRide"
        | Run -> "Run"
        | VirtualRun -> "VirtualRun"
        | Other s -> s
    
    let formatDuration (ts: TimeSpan) =
        if ts.TotalHours >= 1.0 then
            sprintf "%d:%02d:%02d" (int ts.TotalHours) ts.Minutes ts.Seconds
        else
            sprintf "%d:%02d" ts.Minutes ts.Seconds
    
    let formatDistance (meters: float) =
        if meters >= 1000.0 then
            sprintf "%.1f km" (meters / 1000.0)
        else
            sprintf "%.0f m" meters
