module FitSync.MyWhoosh

open System
open System.Net.Http
open System.Net.Http.Json
open System.Text.Json
open System.Text.Json.Serialization

// API Types
type Activity = {
    [<JsonPropertyName("id")>] Id: string
    [<JsonPropertyName("date")>] Date: int64
    [<JsonPropertyName("title")>] Title: string
    [<JsonPropertyName("routeName")>] RouteName: string
    [<JsonPropertyName("distance")>] Distance: float
    [<JsonPropertyName("elevation")>] Elevation: float
    [<JsonPropertyName("watt")>] Watt: float
    [<JsonPropertyName("heartrate")>] Heartrate: float
    [<JsonPropertyName("rideDuration")>] RideDuration: string
    [<JsonPropertyName("activityFileId")>] ActivityFileId: string
    [<JsonPropertyName("startDatetime")>] StartDatetime: string
}

type ActivitiesResponse = {
    [<JsonPropertyName("error")>] Error: bool
    [<JsonPropertyName("code")>] Code: int
    [<JsonPropertyName("data")>] Data: {| results: Activity list; totalPages: int |}
}

type DownloadData = {
    [<JsonPropertyName("fileUrl")>] FileUrl: string
}

type DownloadResponse = {
    [<JsonPropertyName("error")>] Error: bool
    [<JsonPropertyName("data")>] Data: DownloadData
}

// API Endpoints
let private service14 = "https://service14.mywhoosh.com"
let private eventApi = "https://event.mywhoosh.com"

let private createClient (token: string) =
    let client = new HttpClient()
    client.DefaultRequestHeaders.Add("Authorization", $"Bearer {token}")
    client.DefaultRequestHeaders.Add("User-Agent", "Mozilla/5.0 FitSync/1.0")
    client.DefaultRequestHeaders.Add("Accept", "application/json")
    client

/// List activities from MyWhoosh, page by page
let listActivities (token: string) (page: int) = async {
    use client = createClient token
    let url = $"{service14}/v2/rider/profile/activities"
    let body = {| sortDate = "DESC"; page = page |}
    
    let! response = 
        client.PostAsJsonAsync(url, body) 
        |> Async.AwaitTask
    
    let! result = 
        response.Content.ReadFromJsonAsync<ActivitiesResponse>() 
        |> Async.AwaitTask
    
    return result.Data.results, result.Data.totalPages
}

/// List ALL activities (all pages)
let listAllActivities (token: string) = async {
    let rec loop page acc = async {
        let! (activities, totalPages) = listActivities token page
        let all = acc @ activities
        if page < totalPages then
            return! loop (page + 1) all
        else
            return all
    }
    return! loop 1 []
}

/// Get download URL for a FIT file
let getDownloadUrl (token: string) (fileId: string) = async {
    use client = createClient token
    let url = $"{service14}/v2/rider/profile/download-activity-file"
    let body = {| fileId = fileId |}
    
    let! response = 
        client.PostAsJsonAsync(url, body) 
        |> Async.AwaitTask
    
    let! json = response.Content.ReadAsStringAsync() |> Async.AwaitTask
    
    // Parse dynamically to handle the response structure
    use doc = JsonDocument.Parse(json)
    let root = doc.RootElement
    
    try
        let errorProp = root.GetProperty("error")
        if errorProp.GetBoolean() then
            return None
        else
            // data is the URL string directly, not an object
            let fileUrl = root.GetProperty("data").GetString()
            return Some fileUrl
    with ex ->
        printfn "Parse error: %s" ex.Message
        return None
}

/// Download FIT file bytes from S3 URL
let downloadFitFile (url: string) = async {
    use client = new HttpClient()
    let! bytes = client.GetByteArrayAsync(url) |> Async.AwaitTask
    return bytes
}

/// Download a FIT file by activity file ID
let downloadActivity (token: string) (fileId: string) = async {
    match! getDownloadUrl token fileId with
    | Some url -> 
        let! bytes = downloadFitFile url
        return Some bytes
    | None -> 
        return None
}
