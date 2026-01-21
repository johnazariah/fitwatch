module FitSync.IGPSport

open System
open System.Net.Http
open System.Net.Http.Json
open System.Text.Json
open System.Text.Json.Serialization

// API Types - based on actual API response
type ActivitySummary = {
    [<JsonPropertyName("id")>] Id: string  // MongoDB ObjectId string
    [<JsonPropertyName("rideId")>] RideId: int64  // The actual activity ID for detail requests
    [<JsonPropertyName("startTime")>] StartTime: string  // "2025.10.14" format
    [<JsonPropertyName("title")>] Title: string option
    [<JsonPropertyName("rideDistance")>] RideDistance: float  // meters
    [<JsonPropertyName("totalMovingTime")>] TotalMovingTime: float  // seconds
    [<JsonPropertyName("avgSpeed")>] AvgSpeed: float option
    [<JsonPropertyName("avgPower")>] AvgPower: float option
    [<JsonPropertyName("avgHeartRate")>] AvgHeartRate: float option
    [<JsonPropertyName("totalAscent")>] TotalAscent: float option
}

type ActivityListData = {
    [<JsonPropertyName("rows")>] Rows: ActivitySummary list
    [<JsonPropertyName("totalRows")>] TotalRows: int
    [<JsonPropertyName("totalPage")>] TotalPage: int
    [<JsonPropertyName("pageNo")>] PageNo: int
    [<JsonPropertyName("pageSize")>] PageSize: int
}

type ActivityListResponse = {
    [<JsonPropertyName("code")>] Code: int
    [<JsonPropertyName("message")>] Message: string option
    [<JsonPropertyName("data")>] Data: ActivityListData option
}

type ActivityDetail = {
    [<JsonPropertyName("id")>] Id: int64
    [<JsonPropertyName("fitUrl")>] FitUrl: string option
    [<JsonPropertyName("startTime")>] StartTime: string
    [<JsonPropertyName("rideName")>] RideName: string option
    [<JsonPropertyName("distance")>] Distance: float
    [<JsonPropertyName("duration")>] Duration: int64
}

type ActivityDetailResponse = {
    [<JsonPropertyName("code")>] Code: int
    [<JsonPropertyName("message")>] Message: string option
    [<JsonPropertyName("data")>] Data: ActivityDetail option
}

// API Endpoints
let private baseUrl = "https://prod.en.igpsport.com"
let private webGateway = $"{baseUrl}/service/web-gateway/web-analyze/activity"

let private createClient (token: string) =
    let client = new HttpClient()
    client.DefaultRequestHeaders.Add("Authorization", $"Bearer {token}")
    client.DefaultRequestHeaders.Add("Accept", "application/json")
    client.DefaultRequestHeaders.Add("Origin", "https://app.igpsport.com")
    client.DefaultRequestHeaders.Add("Referer", "https://app.igpsport.com/")
    client.DefaultRequestHeaders.Add("qiwu-app-version", "1.0.0")
    client

/// List activities from iGPSport, page by page
/// reqType: 0 = all, sort: 1 = newest first
let listActivities (token: string) (page: int) (pageSize: int) = async {
    use client = createClient token
    let url = $"{webGateway}/queryMyActivity?pageNo={page}&pageSize={pageSize}&reqType=0&sort=1"
    
    try
        let! response = client.GetAsync(url) |> Async.AwaitTask
        let! json = response.Content.ReadAsStringAsync() |> Async.AwaitTask
        
        if not response.IsSuccessStatusCode then
            printfn $"API Error: {response.StatusCode}"
            printfn $"Response: {json}"
            return [], 0
        else
            let options = JsonSerializerOptions(PropertyNameCaseInsensitive = true)
            let result = JsonSerializer.Deserialize<ActivityListResponse>(json, options)
            
            match result.Data with
            | Some data -> 
                return data.Rows, data.TotalPage
            | None -> 
                printfn $"No data in response"
                return [], 0
    with ex ->
        printfn $"Error listing activities: {ex.Message}"
        return [], 0
}

/// List ALL activities (all pages)
let listAllActivities (token: string) = async {
    let pageSize = 20
    let rec loop page acc = async {
        let! (activities, totalPages) = listActivities token page pageSize
        let all = acc @ activities
        if page < totalPages && not (List.isEmpty activities) then
            return! loop (page + 1) all
        else
            return all
    }
    return! loop 1 []
}

/// Get activity detail including FIT URL
let getActivityDetail (token: string) (activityId: int64) = async {
    use client = createClient token
    let url = $"{webGateway}/queryActivityDetail/{activityId}"
    
    try
        let! response = client.GetAsync(url) |> Async.AwaitTask
        let! json = response.Content.ReadAsStringAsync() |> Async.AwaitTask
        
        if not response.IsSuccessStatusCode then
            printfn $"API Error: {response.StatusCode}"
            return None
        else
            // Parse dynamically to handle nested structure
            use doc = JsonDocument.Parse(json)
            let root = doc.RootElement
            
            let mutable dataElement = Unchecked.defaultof<JsonElement>
            if root.TryGetProperty("data", &dataElement) then
                let mutable fitUrlElement = Unchecked.defaultof<JsonElement>
                if dataElement.TryGetProperty("fitUrl", &fitUrlElement) then
                    let fitUrl = fitUrlElement.GetString()
                    return Some fitUrl
                else
                    printfn "No fitUrl in response"
                    return None
            else
                printfn "No data in response"
                return None
    with ex ->
        printfn $"Error getting detail: {ex.Message}"
        return None
}

/// Download FIT file bytes from Aliyun OSS URL (no auth needed)
let downloadFitFile (url: string) = async {
    use client = new HttpClient()
    try
        let! bytes = client.GetByteArrayAsync(url) |> Async.AwaitTask
        return Some bytes
    with ex ->
        printfn $"Error downloading FIT: {ex.Message}"
        return None
}

/// Download a FIT file by activity ID (uses rideId, not MongoDB id)
let downloadActivity (token: string) (rideId: int64) = async {
    match! getActivityDetail token rideId with
    | Some fitUrl -> 
        return! downloadFitFile fitUrl
    | None -> 
        return None
}

/// Parse iGPSport activity type
let toActivityType (activity: ActivitySummary) : Domain.ActivityType =
    // iGPSport is primarily for cycling
    Domain.ActivityType.Ride

/// Parse date from "2025.10.14" format
let parseStartTime (s: string) =
    // Try multiple formats
    let formats = [| "yyyy.MM.dd"; "yyyy-MM-dd"; "yyyy/MM/dd" |]
    match DateTimeOffset.TryParseExact(s, formats, null, System.Globalization.DateTimeStyles.None) with
    | true, dt -> dt
    | false, _ -> 
        match DateTimeOffset.TryParse(s) with
        | true, dt -> dt
        | false, _ -> DateTimeOffset.MinValue

/// Convert to domain ActivityMetadata
let toMetadata (activity: ActivitySummary) : Domain.ActivityMetadata =
    {
        SourceId = string activity.RideId
        Source = Domain.Source.IGPSport
        Title = activity.Title
        ActivityType = toActivityType activity
        StartTime = parseStartTime activity.StartTime
        Duration = Some (TimeSpan.FromSeconds(activity.TotalMovingTime))
        Distance = Some activity.RideDistance  // in meters
        TotalWork = None
        AveragePower = activity.AvgPower
        NormalizedPower = None
        TSS = None
    }

/// Fetch an activity with its FIT file as a SourceActivity
let fetchActivity (token: string) (activity: ActivitySummary) : Async<Domain.SourceActivity option> = async {
    match! downloadActivity token activity.RideId with
    | Some fitBytes ->
        return Some {
            Metadata = toMetadata activity
            FitData = fitBytes
        }
    | None -> return None
}
