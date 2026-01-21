module FitSync.TrainingPeaks

open System
open System.Net.Http
open System.Net.Http.Headers
open System.Net.Http.Json
open System.Text.Json
open System.Text.Json.Serialization

// TrainingPeaks API types
type Workout = {
    [<JsonPropertyName("workoutId")>] WorkoutId: int64
    [<JsonPropertyName("athleteId")>] AthleteId: int
    [<JsonPropertyName("workoutDay")>] WorkoutDay: string
    [<JsonPropertyName("title")>] Title: string option
    [<JsonPropertyName("workoutType")>] WorkoutType: string option
    [<JsonPropertyName("totalTime")>] TotalTime: float option
    [<JsonPropertyName("totalDistance")>] TotalDistance: float option
    [<JsonPropertyName("totalDistanceCustom")>] TotalDistanceCustom: float option
    [<JsonPropertyName("tssActual")>] TssActual: float option
    [<JsonPropertyName("ifActual")>] IfActual: float option
    [<JsonPropertyName("normalizedPower")>] NormalizedPower: float option
    [<JsonPropertyName("completed")>] Completed: bool option
}

type WorkoutFileInfo = {
    [<JsonPropertyName("fileId")>] FileId: int64
    [<JsonPropertyName("fileName")>] FileName: string
}

type WorkoutDetails = {
    [<JsonPropertyName("workoutId")>] WorkoutId: int64
    [<JsonPropertyName("workoutDeviceFileInfos")>] WorkoutDeviceFileInfos: WorkoutFileInfo list option
}

// TrainingPeaks uses OAuth, but there's also a "fitness" subdomain API
// Let's try the approach of using browser cookies like MyWhoosh

let private baseUrl = "https://www.trainingpeaks.com"
let private apiUrl = "https://tpapi.trainingpeaks.com"

let private createClient (accessToken: string) =
    let client = new HttpClient()
    client.DefaultRequestHeaders.Authorization <- AuthenticationHeaderValue("Bearer", accessToken)
    client.DefaultRequestHeaders.Add("Accept", "application/json")
    client

/// List completed workouts from TrainingPeaks
let listWorkouts (accessToken: string) (athleteId: int) (startDate: DateTime) (endDate: DateTime) = async {
    use client = createClient accessToken
    
    let startStr = startDate.ToString("yyyy-MM-dd")
    let endStr = endDate.ToString("yyyy-MM-dd")
    
    // TrainingPeaks API endpoint for workouts (v6) - path includes date range
    let url = $"{apiUrl}/fitness/v6/athletes/{athleteId}/workouts/{startStr}/{endStr}"
    
    try
        let! response = client.GetAsync(url) |> Async.AwaitTask
        let! json = response.Content.ReadAsStringAsync() |> Async.AwaitTask
        
        if response.IsSuccessStatusCode then
            // Debug: dump first workout JSON to see field names
            use doc = JsonDocument.Parse(json)
            let arr = doc.RootElement
            if arr.GetArrayLength() > 0 then
                let first = arr.[0]
                printfn "DEBUG: First workout raw JSON fields:"
                for prop in first.EnumerateObject() do
                    let valueStr = 
                        if prop.Value.ValueKind = JsonValueKind.String then prop.Value.GetString()
                        elif prop.Value.ValueKind = JsonValueKind.Number then prop.Value.GetRawText()
                        elif prop.Value.ValueKind = JsonValueKind.Null then "null"
                        else 
                            let raw = prop.Value.GetRawText()
                            raw.Substring(0, min 50 raw.Length)
                    printfn "  %s = %s" prop.Name valueStr
            
            let options = JsonSerializerOptions(PropertyNameCaseInsensitive = true)
            let workouts = JsonSerializer.Deserialize<Workout list>(json, options)
            // Filter to completed workouts only (have actual data)
            let completed = workouts |> List.filter (fun w -> w.TotalTime.IsSome || w.TotalDistance.IsSome || w.TssActual.IsSome)
            return Ok completed
        else
            return Error $"HTTP {int response.StatusCode}: {json.Substring(0, min 200 json.Length)}"
    with ex ->
        return Error ex.Message
}

/// Get workout details to find the file ID
let getWorkoutDetails (accessToken: string) (athleteId: int) (workoutId: int64) = async {
    use client = createClient accessToken
    
    let url = $"{apiUrl}/fitness/v6/athletes/{athleteId}/workouts/{workoutId}/details"
    
    try
        let! response = client.GetAsync(url) |> Async.AwaitTask
        let! json = response.Content.ReadAsStringAsync() |> Async.AwaitTask
        
        if response.IsSuccessStatusCode then
            let options = JsonSerializerOptions(PropertyNameCaseInsensitive = true)
            let details = JsonSerializer.Deserialize<WorkoutDetails>(json, options)
            return Ok details
        else
            return Error $"HTTP {int response.StatusCode}"
    with ex ->
        return Error ex.Message
}

/// Download a workout file from TrainingPeaks
let downloadWorkoutFile (accessToken: string) (athleteId: int) (workoutId: int64) (fileId: int64) = async {
    use client = createClient accessToken
    
    // v6 endpoint with fileId from details
    let url = $"{apiUrl}/fitness/v6/athletes/{athleteId}/workouts/{workoutId}/rawfiledata/{fileId}"
    
    try
        let! response = client.GetAsync(url) |> Async.AwaitTask
        
        if response.IsSuccessStatusCode then
            let! bytes = response.Content.ReadAsByteArrayAsync() |> Async.AwaitTask
            return Some bytes
        else
            let! error = response.Content.ReadAsStringAsync() |> Async.AwaitTask
            printfn "Download error: %s" error
            return None
    with ex ->
        printfn "Exception: %s" ex.Message
        return None
}

open FitSync.Domain

/// Convert TrainingPeaks workout type to domain ActivityType
let private toActivityType (workoutType: string option) =
    match workoutType with
    | Some "Bike" -> VirtualRide  // TP virtual workouts
    | Some "Ride" -> Ride
    | Some "Run" -> Run
    | Some other -> Other other
    | None -> Other "Unknown"

/// Convert a TrainingPeaks Workout to domain ActivityMetadata
let toMetadata (workout: Workout) : ActivityMetadata =
    {
        SourceId = string workout.WorkoutId
        Source = Source.TrainingPeaks
        Title = workout.Title
        ActivityType = toActivityType workout.WorkoutType
        StartTime = DateTimeOffset.Parse(workout.WorkoutDay)
        Duration = workout.TotalTime |> Option.map (fun h -> TimeSpan.FromHours(h))
        Distance = workout.TotalDistance  // Already in meters (check this)
        TotalWork = None  // Not directly available
        AveragePower = None  // Would need to calculate from NP/IF
        NormalizedPower = workout.NormalizedPower
        TSS = workout.TssActual
    }

/// Fetch a workout with its FIT file as a SourceActivity
let fetchActivity (accessToken: string) (athleteId: int) (workout: Workout) : Async<SourceActivity option> = async {
    // Get workout details to find file ID
    match! getWorkoutDetails accessToken athleteId workout.WorkoutId with
    | Ok details ->
        match details.WorkoutDeviceFileInfos with
        | Some (fileInfo :: _) ->
            match! downloadWorkoutFile accessToken athleteId workout.WorkoutId fileInfo.FileId with
            | Some fitBytes ->
                return Some {
                    Metadata = toMetadata workout
                    FitData = fitBytes
                }
            | None -> return None
        | _ -> return None  // No file attached
    | Error _ -> return None
}

/// Get athlete info (needed to get athleteId)
let getAthleteInfo (accessToken: string) = async {
    use client = createClient accessToken
    
    let url = $"{apiUrl}/fitness/v1/athletes/self"
    
    try
        let! response = client.GetAsync(url) |> Async.AwaitTask
        let! json = response.Content.ReadAsStringAsync() |> Async.AwaitTask
        
        if response.IsSuccessStatusCode then
            use doc = JsonDocument.Parse(json)
            let root = doc.RootElement
            let athleteId = root.GetProperty("athleteId").GetInt32()
            return Ok athleteId
        else
            return Error $"HTTP {int response.StatusCode}: {json}"
    with ex ->
        return Error ex.Message
}
