module FitSync.Intervals

open System
open System.Net.Http
open System.Net.Http.Headers
open System.Text
open System.Text.Json
open System.Text.Json.Serialization

// Activity type from Intervals.icu
type Activity = {
    [<JsonPropertyName("id")>] Id: string
    [<JsonPropertyName("start_date_local")>] StartDateLocal: string
    [<JsonPropertyName("name")>] Name: string option
    [<JsonPropertyName("type")>] Type: string
    [<JsonPropertyName("moving_time")>] MovingTime: int option
    [<JsonPropertyName("distance")>] Distance: float option
    [<JsonPropertyName("total_elevation_gain")>] ElevationGain: float option
    [<JsonPropertyName("average_watts")>] AverageWatts: float option
}

let private createClient (apiKey: string) =
    let client = new HttpClient()
    let authBytes = Encoding.UTF8.GetBytes($"API_KEY:{apiKey}")
    let authHeader = Convert.ToBase64String(authBytes)
    client.DefaultRequestHeaders.Authorization <- AuthenticationHeaderValue("Basic", authHeader)
    client

/// List activities from Intervals.icu
let listActivities (apiKey: string) (athleteId: string) (oldest: DateTimeOffset option) (newest: DateTimeOffset option) = async {
    use client = createClient apiKey
    
    let oldestStr = oldest |> Option.map (fun d -> d.ToString("yyyy-MM-dd")) |> Option.defaultValue "2020-01-01"
    let newestStr = newest |> Option.map (fun d -> d.ToString("yyyy-MM-dd")) |> Option.defaultValue (DateTime.Now.ToString("yyyy-MM-dd"))
    
    let url = $"https://intervals.icu/api/v1/athlete/{athleteId}/activities?oldest={oldestStr}&newest={newestStr}"
    
    let! response = client.GetAsync(url) |> Async.AwaitTask
    let! json = response.Content.ReadAsStringAsync() |> Async.AwaitTask
    
    if response.IsSuccessStatusCode then
        let options = JsonSerializerOptions(PropertyNameCaseInsensitive = true)
        let activities = JsonSerializer.Deserialize<Activity list>(json, options)
        return Ok activities
    else
        return Error json
}

open FitSync.Domain

/// Convert Intervals.icu Activity to domain SinkActivity
let toSinkActivity (activity: Activity) : SinkActivity =
    {
        SinkId = activity.Id
        Sink = Sink.IntervalsIcu
        StartTime = DateTimeOffset.Parse(activity.StartDateLocal)
        Duration = activity.MovingTime |> Option.map (fun s -> TimeSpan.FromSeconds(float s))
        Distance = activity.Distance
        Name = activity.Name
        Source = None  // Could parse from activity metadata if available
    }

/// List activities as domain SinkActivity objects
let listSinkActivities (apiKey: string) (athleteId: string) (oldest: DateTimeOffset option) (newest: DateTimeOffset option) = async {
    match! listActivities apiKey athleteId oldest newest with
    | Ok activities -> return Ok (activities |> List.map toSinkActivity)
    | Error e -> return Error e
}

/// Upload a FIT file to Intervals.icu
let uploadFitFile (apiKey: string) (athleteId: string) (fileName: string) (fitData: byte[]) (activityName: string option) = async {
    use client = createClient apiKey
    
    // Create multipart form with the FIT file
    use content = new MultipartFormDataContent()
    use fileContent = new ByteArrayContent(fitData)
    fileContent.Headers.ContentType <- MediaTypeHeaderValue("application/octet-stream")
    content.Add(fileContent, "file", fileName)
    
    // Add activity name if provided
    match activityName with
    | Some name -> 
        use nameContent = new StringContent(name)
        content.Add(nameContent, "name")
    | None -> ()
    
    let url = $"https://intervals.icu/api/v1/athlete/{athleteId}/activities"
    
    let! response = client.PostAsync(url, content) |> Async.AwaitTask
    let! body = response.Content.ReadAsStringAsync() |> Async.AwaitTask
    
    return response.IsSuccessStatusCode, body
}

/// Update an activity's name in Intervals.icu
let updateActivityName (apiKey: string) (athleteId: string) (activityId: string) (name: string) = async {
    use client = createClient apiKey
    
    let payload = JsonSerializer.Serialize({| name = name |})
    use content = new StringContent(payload, Encoding.UTF8, "application/json")
    
    // Intervals.icu API uses PUT on the activity endpoint
    // But we need to send to the correct endpoint format
    let url = $"https://intervals.icu/api/v1/activity/{activityId}"
    
    let! response = client.PutAsync(url, content) |> Async.AwaitTask
    let! body = response.Content.ReadAsStringAsync() |> Async.AwaitTask
    
    return response.IsSuccessStatusCode, body
}
