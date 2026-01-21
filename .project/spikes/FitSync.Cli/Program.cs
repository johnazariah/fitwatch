using FitSync.Cli.Services;

// Simple CLI for spike - not using System.CommandLine to avoid API churn
var configStore = new ConfigStore();
var myWhooshClient = new MyWhooshWebClient(configStore);  // Use web API instead of game API
var intervalsClient = new IntervalsIcuClient(configStore);

if (args.Length == 0)
{
    PrintHelp();
    return 0;
}

var command = args[0].ToLower();

switch (command)
{
    case "config":
        await HandleConfig(args);
        break;
    case "login":
        await HandleLogin(args);
        break;
    case "list":
        await HandleList(args);
        break;
    case "download":
        await HandleDownload(args);
        break;
    case "upload":
        await HandleUpload(args);
        break;
    case "sync":
        await HandleSync(args);
        break;
    default:
        Console.WriteLine($"Unknown command: {command}");
        PrintHelp();
        return 1;
}

return 0;

async Task HandleLogin(string[] args)
{
    bool force = args.Contains("--force") || args.Contains("-f");
    
    if (force)
    {
        // Clear cached token
        configStore.Set("mywhoosh:web_token", "");
        configStore.Set("mywhoosh:whoosh_id", "");
        Console.WriteLine("Cleared cached credentials.");
    }
    
    await myWhooshClient.LoginAsync();
}

void PrintHelp()
{
    Console.WriteLine("""
        FitSync - Sync cycling activities from MyWhoosh to Intervals.icu
        
        Usage: fitsync <command> [options]
        
        Commands:
          login [--force]            Authenticate with MyWhoosh (opens browser)
          config set <key> <value>   Set a configuration value
          config list                List all configuration values
          list                       List activities from MyWhoosh
          download <activity-id>     Download FIT file from MyWhoosh
          upload <file.fit>          Upload FIT file to Intervals.icu
          upload test                Test Intervals.icu connection
          sync [--since <date>]      Sync activities from MyWhoosh to Intervals.icu
        
        Configuration keys:
          intervals:apikey           Your Intervals.icu API key
          intervals:athleteid        Your Intervals.icu athlete ID (e.g., i12345)
        
        Examples:
          fitsync login              Log in to MyWhoosh via browser
          fitsync login --force      Clear cached token and re-login
          fitsync list               List your recent rides
          fitsync sync               Sync all rides to Intervals.icu
          fitsync sync --since 2026-01-01
        """);
}

async Task HandleConfig(string[] args)
{
    if (args.Length < 2)
    {
        Console.WriteLine("Usage: fitsync config <set|list> [key] [value]");
        return;
    }

    var subCommand = args[1].ToLower();
    
    if (subCommand == "list")
    {
        configStore.List();
    }
    else if (subCommand == "set" && args.Length >= 4)
    {
        configStore.Set(args[2], args[3]);
    }
    else if (subCommand == "get" && args.Length >= 3)
    {
        var value = configStore.Get(args[2]);
        Console.WriteLine(value ?? "(not set)");
    }
    else
    {
        Console.WriteLine("Usage: fitsync config <set|get|list> [key] [value]");
    }
}

async Task HandleList(string[] args)
{
    Console.WriteLine("Fetching activities from MyWhoosh...");
    var activities = await myWhooshClient.ListActivitiesAsync();

    if (activities.Count == 0)
    {
        Console.WriteLine("No activities found.");
        return;
    }

    Console.WriteLine($"Found {activities.Count} activities:");
    Console.WriteLine(new string('-', 80));
    
    foreach (var activity in activities)
    {
        var date = DateTimeOffset.FromUnixTimeSeconds(activity.Date).LocalDateTime;
        Console.WriteLine($"  {activity.ActivityFileId} | {date:yyyy-MM-dd HH:mm} | {activity.Title} | {activity.Distance:F1}km | {activity.Watt}W | {activity.RideDuration}");
    }
}

async Task HandleDownload(string[] args)
{
    if (args.Length < 2)
    {
        Console.WriteLine("Usage: fitsync download <activity-id> [--output <dir>]");
        return;
    }

    var activityId = args[1];
    var outputDir = Directory.GetCurrentDirectory();
    
    for (int i = 2; i < args.Length - 1; i++)
    {
        if (args[i] == "--output")
        {
            outputDir = args[i + 1];
        }
    }
    
    Console.WriteLine($"Downloading activity {activityId}...");
    
    var stream = await myWhooshClient.DownloadFitAsync(activityId);
    if (stream == null)
    {
        Console.WriteLine("Download failed.");
        return;
    }

    Directory.CreateDirectory(outputDir);
    var filePath = Path.Combine(outputDir, $"{activityId}.fit");
    
    using var fileStream = File.Create(filePath);
    await stream.CopyToAsync(fileStream);
    
    Console.WriteLine($"Downloaded: {filePath}");
}

async Task HandleUpload(string[] args)
{
    if (args.Length < 2)
    {
        Console.WriteLine("Usage: fitsync upload <file.fit> | fitsync upload test");
        return;
    }

    if (args[1].ToLower() == "test")
    {
        await intervalsClient.TestConnectionAsync();
        return;
    }

    var filePath = args[1];
    if (!File.Exists(filePath))
    {
        Console.WriteLine($"File not found: {filePath}");
        return;
    }

    Console.WriteLine($"Uploading {Path.GetFileName(filePath)} to Intervals.icu...");
    
    using var stream = File.OpenRead(filePath);
    var result = await intervalsClient.UploadFitAsync(stream, Path.GetFileName(filePath));

    if (result.Success)
    {
        Console.WriteLine("Upload complete!");
    }
    else
    {
        Console.WriteLine($"Upload failed: {result.Error}");
    }
}

async Task HandleSync(string[] args)
{
    DateTime? since = null;
    var dryRun = args.Contains("--dry-run");
    
    for (int i = 1; i < args.Length - 1; i++)
    {
        if (args[i] == "--since" && DateTime.TryParse(args[i + 1], out var d))
        {
            since = d;
        }
    }
    
    Console.WriteLine("Starting sync...");
    
    // 1. Get activities from MyWhoosh
    Console.WriteLine("Fetching activities from MyWhoosh...");
    var activities = await myWhooshClient.ListActivitiesAsync();
    
    // Filter by since date if provided
    if (since.HasValue)
    {
        var sinceUnix = new DateTimeOffset(since.Value).ToUnixTimeSeconds();
        activities = activities.Where(a => a.Date >= sinceUnix).ToList();
    }

    if (activities.Count == 0)
    {
        Console.WriteLine("No activities to sync.");
        return;
    }

    Console.WriteLine($"Found {activities.Count} activities to sync.");

    if (dryRun)
    {
        Console.WriteLine("[DRY RUN] Would sync:");
        foreach (var a in activities)
        {
            Console.WriteLine($"  - {a}");
        }
        return;
    }

    // 2. For each activity, download and upload
    var success = 0;
    var failed = 0;

    foreach (var activity in activities)
    {
        Console.WriteLine($"\nProcessing: {activity.Title ?? activity.Id}");
        
        // Download FIT
        var fitStream = await myWhooshClient.DownloadFitAsync(activity.ActivityFileId ?? activity.Id ?? "");
        if (fitStream == null)
        {
            Console.WriteLine($"  ✗ Failed to download");
            failed++;
            continue;
        }

        // Upload to Intervals.icu
        var filename = $"{activity.ActivityFileId ?? activity.Id}.fit";
        var result = await intervalsClient.UploadFitAsync(fitStream, filename);
        
        if (result.Success)
        {
            Console.WriteLine($"  ✓ Synced to Intervals.icu");
            success++;
        }
        else
        {
            Console.WriteLine($"  ✗ Failed to upload: {result.Error}");
            failed++;
        }
    }

    Console.WriteLine($"\nSync complete: {success} succeeded, {failed} failed");
}
