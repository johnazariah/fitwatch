open System
open System.IO
open System.Diagnostics
open FitSync

// ============================================================================
// CLI Commands
// ============================================================================

let printHelp () =
    printfn """
FitSync - Sync your fitness data 🚴

Sources:
  MyWhoosh       → Indoor cycling platform
  TrainingPeaks  → Training platform
  iGPSport       → GPS bike computer data
  Zwift          → Virtual cycling/running

Sinks:
  Intervals.icu  → Analytics platform

Commands:
  login                   Authenticate with MyWhoosh (opens browser)
  login-tp                Authenticate with TrainingPeaks (opens browser)
  login-igp               Authenticate with iGPSport (opens browser)
  login-zwift             Authenticate with Zwift (opens browser)
  list                    List recent activities from MyWhoosh
  list-tp                 List recent workouts from TrainingPeaks
  list-igp                List recent activities from iGPSport
  list-zwift              List recent activities from Zwift
  list-intervals          List recent activities from Intervals.icu
  verify                  Compare MyWhoosh vs Intervals.icu
  download <id> [path]    Download a FIT file from MyWhoosh
  download-all [path]     Download all FIT files from MyWhoosh
  sync                    Sync MyWhoosh → Intervals.icu
  sync-tp                 Sync TrainingPeaks → Intervals.icu
  sync-igp                Sync iGPSport → Intervals.icu
  sync-zwift              Sync Zwift → Intervals.icu
  config set <key> <val>  Set configuration
  config show             Show current configuration

Examples:
  fitsync login-zwift
  fitsync list-zwift
  fitsync sync-zwift
"""

let requireToken () =
    let config = Config.Config.load()
    match config.mywhooshToken with
    | Some token -> token
    | None -> 
        printfn "❌ Not authenticated. Run: fitsync login"
        exit 1

let requireIntervals () =
    let config = Config.Config.load()
    match config.intervalsApiKey, config.intervalsAthleteId with
    | Some apiKey, Some athleteId -> apiKey, athleteId
    | _ ->
        printfn "❌ Intervals.icu not configured. Run:"
        printfn "   fitsync config set intervals:apikey <key>"
        printfn "   fitsync config set intervals:athleteid <id>"
        exit 1

let requireTrainingPeaks () =
    let config = Config.Config.load()
    match config.trainingPeaksToken, config.trainingPeaksAthleteId with
    | Some token, Some athleteId -> token, athleteId
    | _ ->
        printfn "❌ TrainingPeaks not configured. Run: fitsync login-tp"
        exit 1

let requireIGPSport () =
    let config = Config.Config.load()
    match config.igpsportToken with
    | Some token -> token
    | None -> 
        printfn "❌ iGPSport not configured. Run: fitsync login-igp"
        exit 1

let requireZwift () =
    let config = Config.Config.load()
    match config.zwiftToken with
    | Some token -> token
    | None -> 
        printfn "❌ Zwift not configured. Run: fitsync login-zwift"
        exit 1

// ============================================================================
// Login Command - Browser-based auth
// ============================================================================

let login () =
    printfn "🌐 Opening MyWhoosh login page..."
    printfn ""
    printfn "1. Log in to MyWhoosh in your browser"
    printfn "2. Open DevTools (F12) → Application → Cookies"
    printfn "3. Copy the 'whoosh_token' and 'whoosh_uuid' values"
    printfn ""
    
    // Open browser
    let url = "https://event.mywhoosh.com/auth/login"
    Process.Start(ProcessStartInfo(url, UseShellExecute = true)) |> ignore
    
    printf "Paste whoosh_token: "
    let token = Console.ReadLine().Trim()
    
    printf "Paste whoosh_uuid: "
    let whooshId = Console.ReadLine().Trim()
    
    if String.IsNullOrEmpty(token) || String.IsNullOrEmpty(whooshId) then
        printfn "❌ Invalid token or UUID"
        exit 1
    
    Config.Config.setToken token whooshId |> ignore
    printfn "✅ Saved! You're authenticated."

// ============================================================================
// TrainingPeaks Login - Browser-based auth
// ============================================================================

let loginTrainingPeaks () = async {
    printfn "🌐 Opening TrainingPeaks login page..."
    printfn ""
    printfn "1. Log in to TrainingPeaks in your browser"
    printfn "2. Open DevTools (F12) → Network tab"
    printfn "3. Look for any request to tpapi.trainingpeaks.com"
    printfn "4. Copy the 'authorization: Bearer ...' header value (without 'Bearer ')"
    printfn "5. Find your athlete ID in the URL path (e.g., /athletes/5365138/...)"
    printfn ""
    
    // Open browser
    let url = "https://app.trainingpeaks.com/login"
    Process.Start(ProcessStartInfo(url, UseShellExecute = true)) |> ignore
    
    printf "Paste access_token: "
    let token = Console.ReadLine().Trim()
    
    printf "Paste athlete ID: "
    let athleteIdStr = Console.ReadLine().Trim()
    
    if String.IsNullOrEmpty(token) then
        printfn "❌ Invalid token"
        exit 1
    
    match Int32.TryParse(athleteIdStr) with
    | true, athleteId ->
        Config.Config.update (fun c -> 
            { c with trainingPeaksToken = Some token; trainingPeaksAthleteId = Some athleteId }) |> ignore
        printfn "✅ Saved! Athlete ID: %d" athleteId
    | _ ->
        printfn "❌ Invalid athlete ID"
        exit 1
}

// ============================================================================
// iGPSport Login - Browser-based auth
// ============================================================================

let loginIGPSport () = async {
    printfn "🌐 Opening iGPSport login page..."
    printfn ""
    printfn "1. Log in to iGPSport in your browser"
    printfn "2. Open DevTools (F12) → Network tab"
    printfn "3. Look for any request to prod.en.igpsport.com"
    printfn "4. Copy the 'authorization: Bearer ...' header value (just the token, not 'Bearer ')"
    printfn ""
    
    // Open browser
    let url = "https://app.igpsport.com/sport/history/list?lang=en"
    Process.Start(ProcessStartInfo(url, UseShellExecute = true)) |> ignore
    
    printf "Paste Bearer token: "
    let token = Console.ReadLine().Trim()
    
    if String.IsNullOrEmpty(token) then
        printfn "❌ Invalid token"
        exit 1
    
    Config.Config.update (fun c -> { c with igpsportToken = Some token }) |> ignore
    printfn "✅ Saved! iGPSport token stored."
    
    // Quick test
    printfn "🔍 Testing connection..."
    let! (activities, _) = IGPSport.listActivities token 1 5
    printfn "✅ Found %d activities. You're connected!" (List.length activities)
}

// ============================================================================
// Zwift Login - Browser-based auth
// ============================================================================

let loginZwift () = async {
    printfn "🌐 Opening Zwift login page..."
    printfn ""
    printfn "1. Log in to Zwift in your browser"
    printfn "2. Go to your activity feed"
    printfn "3. Open DevTools (F12) → Network tab"
    printfn "4. Look for requests to us-or-rly101.zwift.com (or similar regional)"
    printfn "5. Copy the 'authorization: Bearer ...' header value (just the token, not 'Bearer ')"
    printfn ""
    printfn "⚠️  Note: Zwift tokens expire in ~6 hours"
    printfn ""
    
    // Open browser
    let url = "https://www.zwift.com/feed"
    Process.Start(ProcessStartInfo(url, UseShellExecute = true)) |> ignore
    
    printf "Paste Bearer token: "
    let token = Console.ReadLine().Trim()
    
    if String.IsNullOrEmpty(token) then
        printfn "❌ Invalid token"
        exit 1
    
    Config.Config.update (fun c -> { c with zwiftToken = Some token }) |> ignore
    printfn "✅ Saved! Zwift token stored."
    
    // Quick test
    printfn "🔍 Testing connection..."
    let! activities = Zwift.listActivities token 5
    printfn "✅ Found %d activities. You're connected!" (List.length activities)
}

/// Debug: download a single Zwift activity to inspect file format
let debugZwiftDownload (activityId: string) = async {
    let token = requireZwift()
    let id = Int64.Parse(activityId)
    printfn $"Downloading activity {id}..."
    match! Zwift.downloadActivity token id with
    | Some bytes ->
        let path = $"zwift_debug_{id}.bin"
        System.IO.File.WriteAllBytes(path, bytes)
        printfn $"Saved {bytes.Length} bytes to {path}"
        // Show header
        let header = bytes |> Array.take (min 32 bytes.Length) |> Array.map (sprintf "%02X") |> String.concat " "
        printfn $"Header: {header}"
        // Try to interpret as text if it starts with printable chars
        if bytes.Length > 0 && bytes.[0] >= 32uy && bytes.[0] < 127uy then
            let text = System.Text.Encoding.UTF8.GetString(bytes, 0, min 200 bytes.Length)
            printfn $"Text: {text}"
    | None ->
        printfn "Failed to download"
}

// ============================================================================
// List Command
// ============================================================================

let list () = async {
    let token = requireToken()
    printfn "📋 Fetching activities from MyWhoosh..."
    
    let! (activities, totalPages) = MyWhoosh.listActivities token 1
    
    printfn ""
    printfn "%-24s %-30s %8s %6s %5s" "Date" "Route" "Distance" "Elev" "Watts"
    printfn "%s" (String.replicate 80 "-")
    
    activities
    |> List.iter (fun a ->
        let date = DateTimeOffset.FromUnixTimeSeconds(a.Date).LocalDateTime.ToString("yyyy-MM-dd HH:mm")
        printfn "%-24s %-30s %7.1fkm %5.0fm %5.0fW" date (a.RouteName.Substring(0, min 30 a.RouteName.Length)) a.Distance a.Elevation a.Watt
    )
    
    printfn ""
    printfn "Page 1 of %d. Showing %d activities." totalPages (List.length activities)
}

// ============================================================================
// List Intervals.icu Command
// ============================================================================

let listIntervals () = async {
    let apiKey, athleteId = requireIntervals()
    printfn "📋 Fetching activities from Intervals.icu..."
    
    match! Intervals.listActivities apiKey athleteId None None with
    | Ok activities ->
        // Filter to VirtualRide only (MyWhoosh activities)
        let rides = activities |> List.filter (fun a -> a.Type = "VirtualRide")
        
        printfn ""
        printfn "%-24s %-30s %8s %6s %5s" "Date" "Name" "Distance" "Elev" "Watts"
        printfn "%s" (String.replicate 80 "-")
        
        rides
        |> List.truncate 20
        |> List.iter (fun a ->
            let name = a.Name |> Option.defaultValue "(unnamed)"
            let dist = a.Distance |> Option.map (fun d -> d / 1000.0) |> Option.defaultValue 0.0
            let elev = a.ElevationGain |> Option.defaultValue 0.0
            let watts = a.AverageWatts |> Option.defaultValue 0.0
            printfn "%-24s %-30s %7.1fkm %5.0fm %5.0fW" a.StartDateLocal (name.Substring(0, min 30 name.Length)) dist elev watts
        )
        
        printfn ""
        printfn "Showing %d of %d VirtualRide activities." (min 20 rides.Length) rides.Length
    | Error err ->
        printfn "❌ Failed to fetch activities: %s" err
}

// ============================================================================
// List TrainingPeaks Command
// ============================================================================

let listTrainingPeaks () = async {
    let token, athleteId = requireTrainingPeaks()
    printfn "📋 Fetching workouts from TrainingPeaks..."
    
    let endDate = DateTime.Now
    let startDate = endDate.AddYears(-2)  // Go back 2 years to find data
    
    match! TrainingPeaks.listWorkouts token athleteId startDate endDate with
    | Ok workouts ->
        printfn ""
        printfn "%-12s %-35s %8s %6s %5s" "Date" "Title" "Duration" "TSS" "NP"
        printfn "%s" (String.replicate 80 "-")
        
        workouts
        |> List.truncate 20
        |> List.iter (fun w ->
            let title = w.Title |> Option.defaultValue "(no title)"
            let duration = w.TotalTime |> Option.map (fun t -> sprintf "%.0fmin" (t / 60.0)) |> Option.defaultValue "-"
            let tss = w.TssActual |> Option.map (sprintf "%.0f") |> Option.defaultValue "-"
            let np = w.NormalizedPower |> Option.map (sprintf "%.0f") |> Option.defaultValue "-"
            printfn "%-12s %-35s %8s %6s %5s" w.WorkoutDay (title.Substring(0, min 35 title.Length)) duration tss np
        )
        
        printfn ""
        printfn "Showing %d of %d completed workouts (last 3 months)." (min 20 workouts.Length) workouts.Length
    | Error err ->
        printfn "❌ Failed to fetch workouts: %s" err
}

// ============================================================================
// List iGPSport Command
// ============================================================================

let listIGPSport () = async {
    let token = requireIGPSport()
    printfn "📋 Fetching activities from iGPSport..."
    
    let! (activities, totalPages) = IGPSport.listActivities token 1 20
    
    printfn ""
    printfn "%-12s %-30s %8s %8s %6s" "Date" "Name" "Distance" "Duration" "Elev"
    printfn "%s" (String.replicate 75 "-")
    
    activities
    |> List.iter (fun a ->
        let name = a.Title |> Option.defaultValue "Outdoor Cycling"
        let dist = a.RideDistance / 1000.0
        let duration = TimeSpan.FromSeconds(a.TotalMovingTime).ToString(@"h\:mm")
        let elev = a.TotalAscent |> Option.map (sprintf "%.0fm") |> Option.defaultValue "-"
        let displayName = if name.Length > 28 then name.Substring(0, 28) + ".." else name
        printfn "%-12s %-30s %7.1fkm %8s %6s" a.StartTime displayName dist duration elev
    )
    
    printfn ""
    printfn "Page 1 of %d. Showing %d activities." totalPages (List.length activities)
}

// ============================================================================
// List Zwift Command
// ============================================================================

let listZwift () = async {
    let token = requireZwift()
    printfn "📋 Fetching activities from Zwift..."
    
    let! activities = Zwift.listActivities token 50
    
    printfn "   Returned %d activities" (List.length activities)
    printfn ""
    printfn "%-20s %-25s %8s %8s %5s" "Date" "Name" "Distance" "Duration" "Watts"
    printfn "%s" (String.replicate 75 "-")
    
    activities
    |> List.iter (fun a ->
        let name = a.Name |> Option.defaultValue "Zwift Ride"
        let dist = a.DistanceInMeters |> Option.map (fun d -> d / 1000.0) |> Option.defaultValue 0.0
        let duration = 
            match a.StartDate, a.EndDate with
            | Some s, Some e ->
                match DateTimeOffset.TryParse(s), DateTimeOffset.TryParse(e) with
                | (true, start), (true, finish) -> (finish - start).ToString(@"h\:mm")
                | _ -> "-"
            | _ -> "-"
        let watts = a.AvgWatts |> Option.map (sprintf "%.0fW") |> Option.defaultValue "-"
        let date = a.StartDate |> Option.map (fun s -> s.Substring(0, min 19 s.Length)) |> Option.defaultValue "-"
        let displayName = if name.Length > 23 then name.Substring(0, 23) + ".." else name
        printfn "%-20s %-25s %7.1fkm %8s %5s" date displayName dist duration watts
    )
    
    printfn ""
    printfn "Showing %d activities." (List.length activities)
}

// ============================================================================
// Sync iGPSport → Intervals.icu Command
// ============================================================================

let syncIGPSport () = async {
    let token = requireIGPSport()
    let apiKey, intervalsAthleteId = requireIntervals()
    
    printfn "🔄 Syncing iGPSport → Intervals.icu..."
    printfn ""
    
    // First, fetch existing Intervals.icu activities for duplicate detection
    printfn "📋 Fetching existing activities from Intervals.icu..."
    let! existingResult = Intervals.listSinkActivities apiKey intervalsAthleteId None None
    
    let existingActivities = 
        match existingResult with
        | Ok activities ->
            printfn "   Found %d existing activities" activities.Length
            activities
        | Error err ->
            printfn "   ⚠️  Could not fetch existing activities: %s" err
            []
    
    // Get activities from iGPSport
    printfn ""
    printfn "📋 Fetching activities from iGPSport..."
    let! activities = IGPSport.listAllActivities token
    printfn "   Found %d activities" (List.length activities)
    
    // Convert to metadata and check for duplicates
    let activitiesWithStatus =
        activities
        |> List.map (fun a ->
            let metadata = IGPSport.toMetadata a
            let dupMatch = Domain.DuplicateDetection.findDuplicate metadata existingActivities
            (a, metadata, dupMatch))
    
    let duplicates = activitiesWithStatus |> List.filter (fun (_, _, m) -> match m with Domain.NoDuplicate -> false | _ -> true)
    let newActivities = activitiesWithStatus |> List.filter (fun (_, _, m) -> match m with Domain.NoDuplicate -> true | _ -> false)
    
    printfn ""
    printfn "🔍 Duplicate detection results:"
    printfn "   ✓ %d exact/probable duplicates (skipping)" duplicates.Length
    printfn "   🆕 %d new activities to sync" newActivities.Length
    
    // Show duplicate details
    if not (List.isEmpty duplicates) && duplicates.Length <= 10 then
        printfn ""
        printfn "Duplicates found:"
        duplicates |> List.iter (fun (a, meta, dupMatch) ->
            let title = meta.Title |> Option.defaultValue $"Ride {a.Id}"
            let date = meta.StartTime.ToString("yyyy-MM-dd")
            match dupMatch with
            | Domain.ExactDuplicate sink ->
                printfn "   ⏭️  %s - %s → matches %s (exact)" date title sink.SinkId
            | Domain.ProbableDuplicate (sink, conf) ->
                printfn "   ⏭️  %s - %s → matches %s (%.0f%% confidence)" date title sink.SinkId (conf * 100.0)
            | Domain.NoDuplicate -> ())
    
    if List.isEmpty newActivities then
        printfn ""
        printfn "✅ No new activities to sync - all activities already in Intervals.icu!"
    else
        printfn ""
        printfn "Syncing new activities..."
        let mutable syncedCount = 0
        let mutable noFileCount = 0
        
        for (activity, meta, _) in newActivities do
            let title = meta.Title |> Option.defaultValue $"Ride {activity.Id}"
            let date = meta.StartTime.ToString("yyyy-MM-dd")
            let duration = meta.Duration |> Option.map Domain.Display.formatDuration |> Option.defaultValue "?"
            
            // Fetch the activity with FIT file
            match! IGPSport.fetchActivity token activity with
            | Some sourceActivity ->
                // Upload to Intervals.icu
                let fileName = $"igpsport_{activity.Id}.fit"
                let success, response = Intervals.uploadFitFile apiKey intervalsAthleteId fileName sourceActivity.FitData meta.Title |> Async.RunSynchronously
                
                if success then
                    printfn "   ✅ %s - %s (%s) - synced!" date title duration
                    syncedCount <- syncedCount + 1
                else
                    printfn "   ❌ %s - %s - upload failed: %s" date title response
            | None ->
                printfn "   ⚠️  %s - %s - no FIT file available" date title
                noFileCount <- noFileCount + 1
        
        printfn ""
        printfn "📊 Sync complete!"
        printfn "   ✅ Synced:        %d" syncedCount
        if noFileCount > 0 then
            printfn "   ⚠️  No file:      %d" noFileCount
}

// ============================================================================
// Sync Zwift → Intervals.icu Command
// ============================================================================

let syncZwift () = async {
    let token = requireZwift()
    let apiKey, intervalsAthleteId = requireIntervals()
    
    printfn "🔄 Syncing Zwift → Intervals.icu..."
    printfn ""
    
    // First, fetch existing Intervals.icu activities for duplicate detection
    printfn "📋 Fetching existing activities from Intervals.icu..."
    let! existingResult = Intervals.listSinkActivities apiKey intervalsAthleteId None None
    
    let existingActivities = 
        match existingResult with
        | Ok activities ->
            printfn "   Found %d existing activities" activities.Length
            activities
        | Error err ->
            printfn "   ⚠️  Could not fetch existing activities: %s" err
            []
    
    // Get activities from Zwift
    // TODO: Discover pagination mechanism when API returns <200 or has cursor in response
    printfn ""
    printfn "📋 Fetching activities from Zwift..."
    let! activities = Zwift.listActivities token 200
    printfn "   Found %d activities (may need pagination for more)" (List.length activities)
    
    // Convert to metadata and check for duplicates
    let activitiesWithStatus =
        activities
        |> List.map (fun a ->
            let metadata = Zwift.toMetadata a
            let dupMatch = Domain.DuplicateDetection.findDuplicate metadata existingActivities
            (a, metadata, dupMatch))
    
    let duplicates = activitiesWithStatus |> List.filter (fun (_, _, m) -> match m with Domain.NoDuplicate -> false | _ -> true)
    let newActivities = activitiesWithStatus |> List.filter (fun (_, _, m) -> match m with Domain.NoDuplicate -> true | _ -> false)
    
    printfn ""
    printfn "🔍 Duplicate detection results:"
    printfn "   ✓ %d exact/probable duplicates (skipping)" duplicates.Length
    printfn "   🆕 %d new activities to sync" newActivities.Length
    
    if List.isEmpty newActivities then
        printfn ""
        printfn "✅ No new activities to sync - all activities already in Intervals.icu!"
    else
        printfn ""
        printfn "Syncing new activities..."
        let mutable syncedCount = 0
        let mutable noFileCount = 0
        
        for (activity, meta, _) in newActivities do
            let title = meta.Title |> Option.defaultValue "Zwift Ride"
            let date = meta.StartTime.ToString("yyyy-MM-dd HH:mm")
            let duration = meta.Duration |> Option.map Domain.Display.formatDuration |> Option.defaultValue "?"
            
            // Fetch the activity with FIT file
            match! Zwift.fetchActivity token activity with
            | Some sourceActivity ->
                // Upload to Intervals.icu
                let fileName = $"zwift_{activity.Id}.fit"
                let success, response = Intervals.uploadFitFile apiKey intervalsAthleteId fileName sourceActivity.FitData meta.Title |> Async.RunSynchronously
                
                if success then
                    printfn "   ✅ %s - %s (%s) - synced!" date title duration
                    syncedCount <- syncedCount + 1
                else
                    printfn "   ❌ %s - %s - upload failed: %s" date title response
            | None ->
                printfn "   ⚠️  %s - %s - no FIT file available" date title
                noFileCount <- noFileCount + 1
        
        printfn ""
        printfn "📊 Sync complete!"
        printfn "   ✅ Synced:        %d" syncedCount
        if noFileCount > 0 then
            printfn "   ⚠️  No file:      %d" noFileCount
}

// ============================================================================
// Update Zwift Names in Intervals.icu
// ============================================================================

let updateZwiftNames () = async {
    let token = requireZwift()
    let apiKey, intervalsAthleteId = requireIntervals()
    
    printfn "🏷️  Updating Zwift activity names in Intervals.icu..."
    
    // Get existing Intervals.icu activities
    printfn ""
    printfn "📋 Fetching existing activities from Intervals.icu..."
    let! intervalsResult = Intervals.listActivities apiKey intervalsAthleteId None None
    let existingActivities = 
        match intervalsResult with
        | Ok activities -> activities
        | Error err ->
            printfn "❌ Failed to fetch Intervals.icu activities: %s" err
            exit 1
    printfn "   Found %d existing activities" (List.length existingActivities)
    
    // Get Zwift activities
    printfn ""
    printfn "📋 Fetching activities from Zwift..."
    let! zwiftActivities = Zwift.listActivities token 200
    printfn "   Found %d activities" (List.length zwiftActivities)
    
    // Convert both to domain types for matching
    let existingSinks = existingActivities |> List.map Intervals.toSinkActivity
    
    // Match Zwift activities to Intervals.icu activities
    let mutable updatedCount = 0
    let mutable skippedCount = 0
    let mutable noMatchCount = 0
    
    printfn ""
    printfn "Matching and updating names..."
    
    for activity in zwiftActivities do
        let meta = Zwift.toMetadata activity
        let zwiftName = meta.Title |> Option.defaultValue "Zwift Ride"
        let date = meta.StartTime.ToString("yyyy-MM-dd HH:mm")
        
        match Domain.DuplicateDetection.findDuplicate meta existingSinks with
        | Domain.ExactDuplicate sink | Domain.ProbableDuplicate (sink, _) ->
            // Check if name already matches
            let intervalsName = sink.Name |> Option.defaultValue ""
            if intervalsName = zwiftName then
                skippedCount <- skippedCount + 1
            else
                let! (success, body) = Intervals.updateActivityName apiKey intervalsAthleteId sink.SinkId zwiftName
                if success then
                    printfn "   ✅ %s: '%s' → '%s'" date intervalsName zwiftName
                    updatedCount <- updatedCount + 1
                else
                    printfn "   ❌ %s: Failed to update: %s" date body
        | Domain.NoDuplicate ->
            noMatchCount <- noMatchCount + 1
    
    printfn ""
    printfn "📊 Update complete!"
    printfn "   ✅ Updated:       %d" updatedCount
    printfn "   ⏭️  Already named: %d" skippedCount
    printfn "   ❓ No match:      %d" noMatchCount
}

// ============================================================================
// Verify Command - Compare sources
// ============================================================================

let verify () = async {
    let token = requireToken()
    let apiKey, athleteId = requireIntervals()
    
    printfn "🔍 Comparing MyWhoosh vs Intervals.icu..."
    printfn ""
    
    // Get all from both sources
    let! mywhooshActivities = MyWhoosh.listAllActivities token
    let! intervalsResult = Intervals.listActivities apiKey athleteId None None
    
    match intervalsResult with
    | Ok intervalsActivities ->
        let virtualRides = intervalsActivities |> List.filter (fun a -> a.Type = "VirtualRide")
        
        printfn "📊 Summary:"
        printfn "   MyWhoosh:      %d activities" mywhooshActivities.Length
        printfn "   Intervals.icu: %d VirtualRide activities" virtualRides.Length
        printfn ""
        
        if mywhooshActivities.Length = virtualRides.Length then
            printfn "✅ Counts match! All activities synced."
        elif mywhooshActivities.Length > virtualRides.Length then
            printfn "⚠️  %d activities in MyWhoosh not in Intervals.icu" (mywhooshActivities.Length - virtualRides.Length)
            printfn "   Run 'fitsync sync' to upload missing activities."
        else
            printfn "ℹ️  More activities in Intervals.icu than MyWhoosh"
            printfn "   (You may have activities from other sources)"
    | Error err ->
        printfn "❌ Failed to fetch from Intervals.icu: %s" err
}

// ============================================================================
// Download Commands
// ============================================================================

let download (fileId: string) (outputPath: string option) = async {
    let token = requireToken()
    printfn "⬇️  Downloading FIT file %s..." fileId
    
    match! MyWhoosh.downloadActivity token fileId with
    | Some bytes ->
        let path = outputPath |> Option.defaultValue $"{fileId}.fit"
        File.WriteAllBytes(path, bytes)
        printfn "✅ Saved to %s (%d bytes)" path bytes.Length
    | None ->
        printfn "❌ Failed to download FIT file"
}

let downloadAll (outputDir: string) = async {
    let token = requireToken()
    printfn "⬇️  Downloading ALL FIT files to %s..." outputDir
    
    Directory.CreateDirectory(outputDir) |> ignore
    
    let! activities = MyWhoosh.listAllActivities token
    
    printfn "Found %d activities. Downloading..." (List.length activities)
    
    let! results =
        activities
        |> List.map (fun a -> async {
            let! fitData = MyWhoosh.downloadActivity token a.ActivityFileId
            match fitData with
            | Some bytes ->
                let date = DateTimeOffset.FromUnixTimeSeconds(a.Date).ToString("yyyy-MM-dd")
                let safeName = a.RouteName.Replace(" ", "_").Replace("/", "-")
                let fileName = $"{date}_{safeName}.fit"
                let path = Path.Combine(outputDir, fileName)
                File.WriteAllBytes(path, bytes)
                return Some (a.RouteName, path)
            | None ->
                return None
        })
        |> Async.Parallel
    
    let successes = results |> Array.choose id
    printfn ""
    printfn "✅ Downloaded %d/%d FIT files" successes.Length activities.Length
    successes |> Array.iter (fun (name, path) -> printfn "   %s" path)
}

// ============================================================================
// Sync Command - The beautiful pipeline!
// ============================================================================

let sync () = async {
    let config = Config.Config.load()
    
    let token = 
        match config.mywhooshToken with
        | Some t -> t
        | None -> 
            printfn "❌ Not authenticated. Run: fitsync login"
            exit 1
    
    let apiKey, athleteId = 
        match config.intervalsApiKey, config.intervalsAthleteId with
        | Some k, Some a -> k, a
        | _ ->
            printfn "❌ Intervals.icu not configured. Run:"
            printfn "   fitsync config set intervals:apikey <key>"
            printfn "   fitsync config set intervals:athleteid <id>"
            exit 1
    
    printfn "🔄 Syncing MyWhoosh → Intervals.icu..."
    
    // Get activities since last sync
    let! activities = MyWhoosh.listAllActivities token
    
    let newActivities = 
        match config.lastSync with
        | Some lastSync ->
            activities 
            |> List.filter (fun a -> 
                DateTimeOffset.FromUnixTimeSeconds(a.Date) > lastSync)
        | None -> 
            activities
    
    printfn "Found %d new activities to sync" (List.length newActivities)
    
    if List.isEmpty newActivities then
        printfn "✅ Already up to date!"
    else
        // The beautiful pipeline! 🦊
        let! results =
            newActivities
            |> List.map (fun activity -> async {
                printfn "  ⬇️  %s..." activity.RouteName
                
                match! MyWhoosh.downloadActivity token activity.ActivityFileId with
                | Some fitBytes ->
                    let fileName = $"{activity.Id}.fit"
                    let! (success, _) = Intervals.uploadFitFile apiKey athleteId fileName fitBytes (Some activity.RouteName)
                    
                    if success then
                        printfn "  ✅ %s uploaded!" activity.RouteName
                        return Some activity.RouteName
                    else
                        printfn "  ❌ %s failed to upload" activity.RouteName
                        return None
                | None ->
                    printfn "  ❌ %s failed to download" activity.RouteName
                    return None
            })
            |> Async.Sequential
        
        let synced = results |> Array.choose id
        Config.Config.updateLastSync() |> ignore
        
        printfn ""
        printfn "✅ Synced %d/%d activities" synced.Length newActivities.Length

}

// ============================================================================
// Sync TrainingPeaks → Intervals.icu
// ============================================================================

let syncTrainingPeaks () = async {
    let token, tpAthleteId = requireTrainingPeaks()
    let apiKey, intervalsAthleteId = requireIntervals()
    
    printfn "🔄 Syncing TrainingPeaks → Intervals.icu..."
    printfn ""
    
    // Get last 2 years of workouts
    let endDate = DateTime.Now
    let startDate = endDate.AddYears(-2)
    
    // First, fetch existing Intervals.icu activities for duplicate detection
    printfn "📋 Fetching existing activities from Intervals.icu..."
    let! existingResult = Intervals.listSinkActivities apiKey intervalsAthleteId (Some (DateTimeOffset startDate)) (Some (DateTimeOffset endDate))
    
    let existingActivities = 
        match existingResult with
        | Ok activities ->
            printfn "   Found %d existing activities" activities.Length
            activities
        | Error err ->
            printfn "   ⚠️  Could not fetch existing activities: %s" err
            []
    
    // Fetch TrainingPeaks workouts
    printfn ""
    printfn "📋 Fetching workouts from TrainingPeaks..."
    match! TrainingPeaks.listWorkouts token tpAthleteId startDate endDate with
    | Ok workouts ->
        printfn "   Found %d completed workouts" workouts.Length
        
        // Convert to metadata and check for duplicates
        let workoutsWithStatus =
            workouts
            |> List.map (fun w ->
                let metadata = TrainingPeaks.toMetadata w
                let dupMatch = Domain.DuplicateDetection.findDuplicate metadata existingActivities
                (w, metadata, dupMatch))
        
        let duplicates = workoutsWithStatus |> List.filter (fun (_, _, m) -> match m with Domain.NoDuplicate -> false | _ -> true)
        let newWorkouts = workoutsWithStatus |> List.filter (fun (_, _, m) -> match m with Domain.NoDuplicate -> true | _ -> false)
        
        printfn ""
        printfn "🔍 Duplicate detection results:"
        printfn "   ✓ %d exact/probable duplicates (skipping)" duplicates.Length
        printfn "   🆕 %d new workouts to sync" newWorkouts.Length
        
        // Show duplicate details
        if not (List.isEmpty duplicates) then
            printfn ""
            printfn "Duplicates found:"
            duplicates |> List.iter (fun (w, meta, dupMatch) ->
                let title = meta.Title |> Option.defaultValue "(untitled)"
                let date = meta.StartTime.ToString("yyyy-MM-dd")
                match dupMatch with
                | Domain.ExactDuplicate sink ->
                    printfn "   ⏭️  %s - %s → matches %s (exact)" date title sink.SinkId
                | Domain.ProbableDuplicate (sink, conf) ->
                    printfn "   ⏭️  %s - %s → matches %s (%.0f%% confidence)" date title sink.SinkId (conf * 100.0)
                | Domain.NoDuplicate -> ())
        
        if List.isEmpty newWorkouts then
            printfn ""
            printfn "✅ No new workouts to sync - all activities already in Intervals.icu!"
        else
            printfn ""
            printfn "Syncing new workouts..."
            let mutable syncedCount = 0
            let mutable noFileCount = 0
            
            for (workout, meta, _) in newWorkouts do
                let title = meta.Title |> Option.defaultValue "(untitled)"
                let date = meta.StartTime.ToString("yyyy-MM-dd")
                let duration = meta.Duration |> Option.map Domain.Display.formatDuration |> Option.defaultValue "?"
                
                // Fetch the activity with FIT file
                match! TrainingPeaks.fetchActivity token tpAthleteId workout with
                | Some sourceActivity ->
                    printfn "  ⬇️  %s - %s (%s, %d bytes)" date title duration sourceActivity.FitData.Length
                    
                    let fileName = $"tp_{meta.SourceId}.fit"
                    let! (success, body) = Intervals.uploadFitFile apiKey intervalsAthleteId fileName sourceActivity.FitData meta.Title
                    
                    if success then
                        printfn "  ✅ Uploaded!"
                        syncedCount <- syncedCount + 1
                    else
                        printfn "  ❌ Upload failed: %s" body
                | None ->
                    printfn "  ⏭️  %s - %s (no FIT file)" date title
                    noFileCount <- noFileCount + 1
            
            printfn ""
            printfn "✅ Synced %d new workouts" syncedCount
            if noFileCount > 0 then
                printfn "   ⏭️  %d workouts had no FIT files attached" noFileCount
    | Error err ->
        printfn "❌ Failed to fetch workouts: %s" err
}

// ============================================================================
// Config Commands
// ============================================================================

let configSet (key: string) (value: string) =
    let config = Config.Config.load()
    let updated = 
        match key.ToLower() with
        | "mywhoosh:token" -> { config with mywhooshToken = Some value }
        | "mywhoosh:whooshid" -> { config with mywhooshWhooshId = Some value }
        | "intervals:apikey" -> { config with intervalsApiKey = Some value }
        | "intervals:athleteid" -> { config with intervalsAthleteId = Some value }
        | "trainingpeaks:token" -> { config with trainingPeaksToken = Some value }
        | "trainingpeaks:athleteid" -> 
            match Int32.TryParse(value) with
            | true, id -> { config with trainingPeaksAthleteId = Some id }
            | _ -> 
                printfn "Invalid athlete ID (must be a number)"
                config
        | "igpsport:token" -> { config with igpsportToken = Some value }
        | "zwift:token" -> { config with zwiftToken = Some value }
        | _ -> 
            printfn "Unknown config key: %s" key
            config
    Config.Config.save updated |> ignore
    printfn "✅ Set %s" key

let configShow () =
    let config = Config.Config.load()
    printfn "Configuration (~/.fitsync/config.json):"
    printfn ""
    printfn "  [MyWhoosh]"
    printfn "  token              = %s" (config.mywhooshToken |> Option.map (fun t -> t.Substring(0, min 20 t.Length) + "...") |> Option.defaultValue "(not set)")
    printfn "  whooshid           = %s" (config.mywhooshWhooshId |> Option.defaultValue "(not set)")
    printfn ""
    printfn "  [TrainingPeaks]"
    printfn "  token              = %s" (config.trainingPeaksToken |> Option.map (fun t -> t.Substring(0, min 20 t.Length) + "...") |> Option.defaultValue "(not set)")
    printfn "  athleteid          = %s" (config.trainingPeaksAthleteId |> Option.map string |> Option.defaultValue "(not set)")
    printfn ""
    printfn "  [iGPSport]"
    printfn "  token              = %s" (config.igpsportToken |> Option.map (fun t -> t.Substring(0, min 20 t.Length) + "...") |> Option.defaultValue "(not set)")
    printfn ""
    printfn "  [Zwift]"
    printfn "  token              = %s" (config.zwiftToken |> Option.map (fun t -> t.Substring(0, min 20 t.Length) + "...") |> Option.defaultValue "(not set)")
    printfn ""
    printfn "  [Intervals.icu]"
    printfn "  apikey             = %s" (config.intervalsApiKey |> Option.map (fun _ -> "***") |> Option.defaultValue "(not set)")
    printfn "  athleteid          = %s" (config.intervalsAthleteId |> Option.defaultValue "(not set)")
    printfn ""
    printfn "  lastSync           = %s" (config.lastSync |> Option.map string |> Option.defaultValue "(never)")

// ============================================================================
// Main
// ============================================================================

[<EntryPoint>]
let main args =
    match args |> Array.toList with
    | [] | ["help"] | ["--help"] | ["-h"] -> 
        printHelp()
        0
    
    | ["login"] -> 
        login()
        0
    
    | ["login-tp"] -> 
        loginTrainingPeaks() |> Async.RunSynchronously
        0
    
    | ["login-igp"] -> 
        loginIGPSport() |> Async.RunSynchronously
        0
    
    | ["login-zwift"] -> 
        loginZwift() |> Async.RunSynchronously
        0
    
    | ["debug-zwift"; activityId] ->
        debugZwiftDownload activityId |> Async.RunSynchronously
        0
    
    | ["list"] -> 
        list() |> Async.RunSynchronously
        0
    
    | ["list-tp"] -> 
        listTrainingPeaks() |> Async.RunSynchronously
        0
    
    | ["list-igp"] -> 
        listIGPSport() |> Async.RunSynchronously
        0
    
    | ["list-zwift"] -> 
        listZwift() |> Async.RunSynchronously
        0
    
    | ["list-intervals"] -> 
        listIntervals() |> Async.RunSynchronously
        0
    
    | ["verify"] -> 
        verify() |> Async.RunSynchronously
        0
    
    | ["download"; fileId] -> 
        download fileId None |> Async.RunSynchronously
        0
    
    | ["download"; fileId; path] -> 
        download fileId (Some path) |> Async.RunSynchronously
        0
    
    | ["download-all"] -> 
        downloadAll "./fits" |> Async.RunSynchronously
        0
    
    | ["download-all"; path] -> 
        downloadAll path |> Async.RunSynchronously
        0
    
    | ["sync"] -> 
        sync() |> Async.RunSynchronously
        0
    
    | ["sync-tp"] -> 
        syncTrainingPeaks() |> Async.RunSynchronously
        0
    
    | ["sync-igp"] -> 
        syncIGPSport() |> Async.RunSynchronously
        0
    
    | ["sync-zwift"] -> 
        syncZwift() |> Async.RunSynchronously
        0
    
    | ["update-zwift-names"] ->
        updateZwiftNames() |> Async.RunSynchronously
        0
    
    | ["config"; "set"; key; value] -> 
        configSet key value
        0
    
    | ["config"; "show"] -> 
        configShow()
        0
    
    | _ -> 
        printfn "Unknown command. Run: fitsync --help"
        1
