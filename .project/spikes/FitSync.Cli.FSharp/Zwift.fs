module FitSync.Zwift

open System
open System.Net.Http
open System.Text.Json
open System.Text.Json.Serialization

// API Types - based on Zwift API discovery

type FitnessData = {
    [<JsonPropertyName("status")>] Status: string
    [<JsonPropertyName("fullDataUrl")>] FullDataUrl: string option
    [<JsonPropertyName("smallDataUrl")>] SmallDataUrl: string option
}

type ActivityProfile = {
    [<JsonPropertyName("id")>] Id: string  // API returns string, not int64
    [<JsonPropertyName("firstName")>] FirstName: string option
    [<JsonPropertyName("lastName")>] LastName: string option
    [<JsonPropertyName("imageSrc")>] ImageSrc: string option
}

type FeedActivity = {
    [<JsonPropertyName("id")>] Id: int64 option  // May not be present, use id_str
    [<JsonPropertyName("id_str")>] IdStr: string
    [<JsonPropertyName("profile")>] Profile: ActivityProfile option
    [<JsonPropertyName("name")>] Name: string option
    [<JsonPropertyName("sport")>] Sport: string option  // "CYCLING", "RUNNING"
    [<JsonPropertyName("startDate")>] StartDate: string option
    [<JsonPropertyName("endDate")>] EndDate: string option
    [<JsonPropertyName("distanceInMeters")>] DistanceInMeters: float option
    [<JsonPropertyName("avgWatts")>] AvgWatts: float option
    [<JsonPropertyName("avgHeartRate")>] AvgHeartRate: float option
    [<JsonPropertyName("calories")>] Calories: float option
    [<JsonPropertyName("totalElevation")>] TotalElevation: float option
    [<JsonPropertyName("fitnessData")>] FitnessData: FitnessData option
}

type ActivityFeedResponse = {
    [<JsonPropertyName("data")>] Data: FeedActivity list option
}

type ActivityDetail = {
    [<JsonPropertyName("id")>] Id: int64
    [<JsonPropertyName("id_str")>] IdStr: string option
    [<JsonPropertyName("name")>] Name: string option
    [<JsonPropertyName("sport")>] Sport: string option
    [<JsonPropertyName("startDate")>] StartDate: string option
    [<JsonPropertyName("endDate")>] EndDate: string option
    [<JsonPropertyName("distanceInMeters")>] DistanceInMeters: float option
    [<JsonPropertyName("movingTimeInMs")>] MovingTimeInMs: int64 option
    [<JsonPropertyName("avgWatts")>] AvgWatts: float option
    [<JsonPropertyName("avgHeartRate")>] AvgHeartRate: float option
    [<JsonPropertyName("totalElevation")>] TotalElevation: float option
    [<JsonPropertyName("fitnessData")>] FitnessData: FitnessData option
    // FIT file on S3
    [<JsonPropertyName("fitFileBucket")>] FitFileBucket: string option
    [<JsonPropertyName("fitFileKey")>] FitFileKey: string option
}

// API Endpoints
// Note: Regional relay - may need to handle different regions
let private apiBase = "https://us-or-rly101.zwift.com"

let private createClient (token: string) =
    let client = new HttpClient()
    client.DefaultRequestHeaders.Add("Authorization", $"Bearer {token}")
    client.DefaultRequestHeaders.Add("Accept", "application/json")
    client.DefaultRequestHeaders.Add("zwift-api-version", "2.5")
    client.DefaultRequestHeaders.Add("source", "my-zwift")
    client

/// List my activities from Zwift activity feed
/// Note: limit=30 was observed from the Zwift web app. 
/// TODO: Discover pagination mechanism (cursor? before/after timestamps?) for fetching all activities
let listActivities (token: string) (limit: int) = async {
    use client = createClient token
    // Try a higher limit - API might support more than the default 30 the web app uses
    let url = $"{apiBase}/api/activity-feed/feed/?limit={limit}&includeInProgress=false&feedType=JUST_ME"
    
    try
        let! response = client.GetAsync(url) |> Async.AwaitTask
        let! json = response.Content.ReadAsStringAsync() |> Async.AwaitTask
        
        if not response.IsSuccessStatusCode then
            printfn $"API Error: {response.StatusCode}"
            printfn $"Response: {json.Substring(0, min 500 json.Length)}"
            return []
        else
            let options = JsonSerializerOptions(PropertyNameCaseInsensitive = true)
            
            // The response might be an array directly or wrapped in { data: [...] }
            try
                // Try as direct array first
                let activities = JsonSerializer.Deserialize<FeedActivity list>(json, options)
                return activities
            with _ ->
                // Try as wrapped response
                try
                    let response = JsonSerializer.Deserialize<ActivityFeedResponse>(json, options)
                    return response.Data |> Option.defaultValue []
                with ex ->
                    printfn $"Parse error: {ex.Message}"
                    printfn $"JSON preview: {json.Substring(0, min 300 json.Length)}"
                    return []
    with ex ->
        printfn $"Error listing activities: {ex.Message}"
        return []
}

/// Get activity detail including FIT file URL
let getActivityDetail (token: string) (activityId: int64) = async {
    use client = createClient token
    let url = $"{apiBase}/api/activities/{activityId}"
    
    try
        let! response = client.GetAsync(url) |> Async.AwaitTask
        let! json = response.Content.ReadAsStringAsync() |> Async.AwaitTask
        
        if not response.IsSuccessStatusCode then
            printfn $"API Error: {response.StatusCode}"
            return None
        else
            let options = JsonSerializerOptions(PropertyNameCaseInsensitive = true)
            let detail = JsonSerializer.Deserialize<ActivityDetail>(json, options)
            return Some detail
    with ex ->
        printfn $"Error getting activity detail: {ex.Message}"
        return None
}

/// Decompress gzip bytes if needed
let private decompressIfGzip (bytes: byte[]) : byte[] =
    // Check for gzip magic bytes (0x1f 0x8b)
    if bytes.Length >= 2 && bytes.[0] = 0x1Fuy && bytes.[1] = 0x8Buy then
        use inputStream = new System.IO.MemoryStream(bytes)
        use gzipStream = new System.IO.Compression.GZipStream(inputStream, System.IO.Compression.CompressionMode.Decompress)
        use outputStream = new System.IO.MemoryStream()
        gzipStream.CopyTo(outputStream)
        outputStream.ToArray()
    else
        bytes

/// Download FIT file from the fullDataUrl
let downloadFitFile (token: string) (fitUrl: string) = async {
    use client = createClient token
    try
        let! bytes = client.GetByteArrayAsync(fitUrl) |> Async.AwaitTask
        // Zwift returns gzip-compressed files
        let decompressed = decompressIfGzip bytes
        // Debug: show first 20 bytes to identify file format
        let header = decompressed |> Array.take (min 20 decompressed.Length) |> Array.map (sprintf "%02X") |> String.concat " "
        printfn $"  Downloaded {bytes.Length} bytes, decompressed to {decompressed.Length} bytes"
        printfn $"  Header bytes: {header}"
        // FIT files start with header size byte (usually 12 or 14), then protocol version, then profile, then ".FIT" signature at bytes 8-11
        return Some decompressed
    with ex ->
        printfn $"Error downloading FIT: {ex.Message}"
        return None
}

/// Download FIT file for an activity (from S3)
let downloadActivity (token: string) (activityId: int64) = async {
    match! getActivityDetail token activityId with
    | Some detail ->
        // FIT file is on S3, not via the relay API
        match detail.FitFileBucket, detail.FitFileKey with
        | Some bucket, Some key ->
            let s3Url = $"https://{bucket}.s3.amazonaws.com/{key}"
            // S3 download doesn't need auth token
            use client = new HttpClient()
            try
                let! bytes = client.GetByteArrayAsync(s3Url) |> Async.AwaitTask
                let decompressed = decompressIfGzip bytes
                return Some decompressed
            with ex ->
                printfn $"  Error downloading from S3: {ex.Message}"
                return None
        | _ ->
            printfn "  No fitFileBucket/fitFileKey in activity detail"
            return None
    | None -> 
        return None
}

/// Parse Zwift activity type
let toActivityType (sport: string option) : Domain.ActivityType =
    match sport with
    | Some "CYCLING" -> Domain.ActivityType.VirtualRide
    | Some "RUNNING" -> Domain.ActivityType.VirtualRun
    | Some other -> Domain.ActivityType.Other other
    | None -> Domain.ActivityType.VirtualRide  // Default for Zwift

/// Parse ISO date string
let parseStartTime (s: string option) =
    match s with
    | Some dateStr ->
        match DateTimeOffset.TryParse(dateStr) with
        | true, dt -> dt
        | false, _ -> DateTimeOffset.MinValue
    | None -> DateTimeOffset.MinValue

/// Convert feed activity to domain ActivityMetadata
let toMetadata (activity: FeedActivity) : Domain.ActivityMetadata =
    let durationMs = 
        // Calculate from start/end if available
        match activity.StartDate, activity.EndDate with
        | Some s, Some e ->
            match DateTimeOffset.TryParse(s), DateTimeOffset.TryParse(e) with
            | (true, start), (true, finish) -> Some (finish - start)
            | _ -> None
        | _ -> None
    
    {
        SourceId = activity.IdStr
        Source = Domain.Source.Zwift
        Title = activity.Name
        ActivityType = toActivityType activity.Sport
        StartTime = parseStartTime activity.StartDate
        Duration = durationMs
        Distance = activity.DistanceInMeters
        TotalWork = None
        AveragePower = activity.AvgWatts
        NormalizedPower = None
        TSS = None
    }

/// Fetch an activity with its FIT file as a SourceActivity
let fetchActivity (token: string) (activity: FeedActivity) : Async<Domain.SourceActivity option> = async {
    let activityId = Int64.Parse(activity.IdStr)
    match! downloadActivity token activityId with
    | Some fitBytes ->
        return Some {
            Metadata = toMetadata activity
            FitData = fitBytes
        }
    | None -> return None
}
