using System.Net.Http.Headers;
using System.Text;

namespace FitSync.Cli.Services;

/// <summary>
/// Client for Intervals.icu API
/// Based on: https://forum.intervals.icu/t/intervals-icu-api-integration-cookbook/80090
/// </summary>
public class IntervalsIcuClient
{
    private readonly ConfigStore _config;
    private readonly HttpClient _http;

    private const string BaseUrl = "https://intervals.icu/api/v1";

    public IntervalsIcuClient(ConfigStore config)
    {
        _config = config;
        _http = new HttpClient();
    }

    private bool Configure()
    {
        var apiKey = _config.Get("intervals:apikey");
        var athleteId = _config.Get("intervals:athleteid");

        if (string.IsNullOrEmpty(apiKey) || string.IsNullOrEmpty(athleteId))
        {
            Console.WriteLine("Error: Intervals.icu not configured.");
            Console.WriteLine("Run: fitsync config set intervals:apikey <your-api-key>");
            Console.WriteLine("Run: fitsync config set intervals:athleteid <your-athlete-id>");
            Console.WriteLine();
            Console.WriteLine("Get your API key from: https://intervals.icu/settings");
            Console.WriteLine("Your athlete ID is shown on your profile (e.g., i12345)");
            return false;
        }

        // Intervals.icu uses Basic auth with "API_KEY" as username and the actual key as password
        var authValue = Convert.ToBase64String(Encoding.UTF8.GetBytes($"API_KEY:{apiKey}"));
        _http.DefaultRequestHeaders.Authorization = new AuthenticationHeaderValue("Basic", authValue);
        
        return true;
    }

    public async Task<UploadResult> UploadFitAsync(Stream fitFile, string filename)
    {
        if (!Configure())
            return new UploadResult { Success = false, Error = "Not configured" };

        var athleteId = _config.Get("intervals:athleteid")!;

        try
        {
            using var content = new MultipartFormDataContent();
            using var fileContent = new StreamContent(fitFile);
            fileContent.Headers.ContentType = new MediaTypeHeaderValue("application/octet-stream");
            content.Add(fileContent, "file", filename);

            var response = await _http.PostAsync($"{BaseUrl}/athlete/{athleteId}/activities", content);

            if (response.IsSuccessStatusCode)
            {
                var result = await response.Content.ReadAsStringAsync();
                Console.WriteLine($"Uploaded successfully: {filename}");
                return new UploadResult { Success = true, Response = result };
            }
            else
            {
                var error = await response.Content.ReadAsStringAsync();
                Console.WriteLine($"Upload failed ({response.StatusCode}): {error}");
                return new UploadResult { Success = false, Error = error };
            }
        }
        catch (Exception ex)
        {
            Console.WriteLine($"Upload error: {ex.Message}");
            return new UploadResult { Success = false, Error = ex.Message };
        }
    }

    public async Task<bool> TestConnectionAsync()
    {
        if (!Configure())
            return false;

        var athleteId = _config.Get("intervals:athleteid")!;

        try
        {
            var response = await _http.GetAsync($"{BaseUrl}/athlete/{athleteId}");
            
            if (response.IsSuccessStatusCode)
            {
                Console.WriteLine("Intervals.icu connection OK");
                return true;
            }
            else
            {
                Console.WriteLine($"Intervals.icu connection failed: {response.StatusCode}");
                return false;
            }
        }
        catch (Exception ex)
        {
            Console.WriteLine($"Connection error: {ex.Message}");
            return false;
        }
    }
}

public class UploadResult
{
    public bool Success { get; set; }
    public string? Response { get; set; }
    public string? Error { get; set; }
}
