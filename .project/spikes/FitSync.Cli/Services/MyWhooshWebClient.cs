using System.Net.Http.Json;
using System.Text.Json;

namespace FitSync.Cli.Services;

/// <summary>
/// Client for MyWhoosh Web API (event.mywhoosh.com)
/// Uses the web login flow which doesn't conflict with game sessions
/// </summary>
public class MyWhooshWebClient
{
    private readonly ConfigStore _config;
    private readonly HttpClient _http;
    private string? _accessToken;
    private string? _whooshId;

    // Web API endpoints (different from game API)
    private const string AuthApiUrl = "https://event.mywhoosh.com/api/auth";
    private const string Service14Url = "https://service14.mywhoosh.com/v2";

    public MyWhooshWebClient(ConfigStore config)
    {
        _config = config;
        _http = new HttpClient();
        _http.DefaultRequestHeaders.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36");
        _http.DefaultRequestHeaders.Add("Accept", "application/json");
        _http.DefaultRequestHeaders.Add("Origin", "https://event.mywhoosh.com");
    }

    public async Task<bool> LoginAsync()
    {
        // Check for cached token first
        var cachedToken = _config.Get("mywhoosh:web_token");
        var cachedWhooshId = _config.Get("mywhoosh:whoosh_id");
        if (!string.IsNullOrEmpty(cachedToken) && !string.IsNullOrEmpty(cachedWhooshId))
        {
            _accessToken = cachedToken;
            _whooshId = cachedWhooshId;
            Console.WriteLine($"Using cached token for WhooshId: {_whooshId}");
            return true;
        }

        // Browser-based login - open browser and ask user to paste token
        Console.WriteLine("\n=== MyWhoosh Authentication ===");
        Console.WriteLine("The MyWhoosh game app blocks API logins, so we need to use browser auth.");
        Console.WriteLine("\nOpening browser to MyWhoosh login page...");
        Console.WriteLine("After logging in, you'll need to copy your token from browser cookies.\n");
        
        // Open browser to login page
        try
        {
            var psi = new System.Diagnostics.ProcessStartInfo
            {
                FileName = "https://event.mywhoosh.com/auth/login",
                UseShellExecute = true
            };
            System.Diagnostics.Process.Start(psi);
        }
        catch
        {
            Console.WriteLine("Could not open browser. Please navigate to:");
            Console.WriteLine("  https://event.mywhoosh.com/auth/login");
        }

        Console.WriteLine("After logging in:");
        Console.WriteLine("  1. Press F12 to open Developer Tools");
        Console.WriteLine("  2. Go to Application tab → Cookies → event.mywhoosh.com");
        Console.WriteLine("  3. Find 'whoosh_token' and copy its value");
        Console.WriteLine("  4. Also copy the 'whoosh_uuid' value\n");
        
        Console.Write("Paste your whoosh_token: ");
        var token = Console.ReadLine()?.Trim();
        
        if (string.IsNullOrEmpty(token))
        {
            Console.WriteLine("No token provided.");
            return false;
        }
        
        Console.Write("Paste your whoosh_uuid: ");
        var whooshId = Console.ReadLine()?.Trim();
        
        if (string.IsNullOrEmpty(whooshId))
        {
            Console.WriteLine("No whoosh_uuid provided.");
            return false;
        }

        // Save and use the token
        _accessToken = token;
        _whooshId = whooshId;
        _config.Set("mywhoosh:web_token", token);
        _config.Set("mywhoosh:whoosh_id", whooshId);
        
        Console.WriteLine($"\nToken saved! WhooshId: {_whooshId}");
        return true;
    }

    public async Task<List<ActivitySummary>> ListActivitiesAsync(int page = 1)
    {
        if (_accessToken == null && !await LoginAsync())
            return new List<ActivitySummary>();

        _http.DefaultRequestHeaders.Authorization = 
            new System.Net.Http.Headers.AuthenticationHeaderValue("Bearer", _accessToken);

        try
        {
            var request = new { sortDate = "DESC", page };
            var json = JsonSerializer.Serialize(request);
            var content = new StringContent(json, System.Text.Encoding.UTF8, "application/json");
            
            Console.WriteLine($"Fetching activities (page {page})...");
            var response = await _http.PostAsync($"{Service14Url}/rider/profile/activities", content);
            var responseBody = await response.Content.ReadAsStringAsync();
            
            if (!response.IsSuccessStatusCode)
            {
                Console.WriteLine($"Failed to list activities: {response.StatusCode}");
                Console.WriteLine(responseBody);
                return new List<ActivitySummary>();
            }

            var result = JsonSerializer.Deserialize<ActivitiesResponse>(responseBody, new JsonSerializerOptions
            {
                PropertyNameCaseInsensitive = true
            });
            
            return result?.Data?.Results ?? new List<ActivitySummary>();
        }
        catch (Exception ex)
        {
            Console.WriteLine($"Error listing activities: {ex.Message}");
            return new List<ActivitySummary>();
        }
    }

    public async Task<Stream?> DownloadFitAsync(string activityFileId)
    {
        if (_accessToken == null && !await LoginAsync())
            return null;

        _http.DefaultRequestHeaders.Authorization = 
            new System.Net.Http.Headers.AuthenticationHeaderValue("Bearer", _accessToken);

        try
        {
            // First get the S3 presigned URL
            var request = new { fileId = activityFileId };
            var json = JsonSerializer.Serialize(request);
            var content = new StringContent(json, System.Text.Encoding.UTF8, "application/json");
            
            var response = await _http.PostAsync($"{Service14Url}/rider/profile/download-activity-file", content);
            var responseBody = await response.Content.ReadAsStringAsync();
            
            if (!response.IsSuccessStatusCode)
            {
                Console.WriteLine($"Failed to get download URL: {response.StatusCode}");
                Console.WriteLine(responseBody);
                return null;
            }

            var result = JsonSerializer.Deserialize<DownloadResponse>(responseBody, new JsonSerializerOptions
            {
                PropertyNameCaseInsensitive = true
            });
            
            if (string.IsNullOrEmpty(result?.Data))
            {
                Console.WriteLine("No download URL in response");
                return null;
            }

            // Download from S3
            Console.WriteLine("Downloading FIT file from S3...");
            var fitResponse = await _http.GetAsync(result.Data);
            
            if (!fitResponse.IsSuccessStatusCode)
            {
                Console.WriteLine($"Failed to download FIT: {fitResponse.StatusCode}");
                return null;
            }

            return await fitResponse.Content.ReadAsStreamAsync();
        }
        catch (Exception ex)
        {
            Console.WriteLine($"Error downloading FIT: {ex.Message}");
            return null;
        }
    }
}

// Response DTOs for Web API
public class WebLoginResponse
{
    public string? Token { get; set; }
    public WebUser? User { get; set; }
}

public class WebUser
{
    public string? WhooshId { get; set; }
    public string? Email { get; set; }
}

public class ActivitiesResponse
{
    public bool Error { get; set; }
    public int Code { get; set; }
    public string? Message { get; set; }
    public ActivitiesData? Data { get; set; }
}

public class ActivitiesData
{
    public List<ActivitySummary>? Results { get; set; }
}

public class ActivitySummary
{
    public string? Id { get; set; }
    public long Date { get; set; }
    public string? Type { get; set; }
    public string? Title { get; set; }
    public string? SportType { get; set; }
    public string? RouteName { get; set; }
    public double Distance { get; set; }
    public int Elevation { get; set; }
    public int Watt { get; set; }
    public int TotalWorkout { get; set; }
    public double WattPerKg { get; set; }
    public int Heartrate { get; set; }
    public string? RideDuration { get; set; }
    public string? StartDatetime { get; set; }
    public string? CreatedAt { get; set; }
    public string? ActivityFileId { get; set; }
}

public class DownloadResponse
{
    public bool Error { get; set; }
    public int Code { get; set; }
    public string? Data { get; set; }  // The S3 presigned URL
}
