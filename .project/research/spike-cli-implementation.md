# Spike: CLI Tool for MyWhoosh → Intervals.icu

## Goal
Build a simple .NET CLI tool that proves we can:
1. Authenticate with MyWhoosh
2. Download FIT files
3. Upload to Intervals.icu

## Resources

- **MyWhoosh API:** https://github.com/mywhoosh-community/mywhoosh-api
- **Intervals.icu API:** https://forum.intervals.icu/t/intervals-icu-api-integration-cookbook/80090

## CLI Commands

```bash
# Configure credentials
fitsync config set mywhoosh:email "your@email.com"
fitsync config set mywhoosh:password "yourpassword"
fitsync config set intervals:apikey "your-api-key"
fitsync config set intervals:athleteid "i12345"

# List recent activities from MyWhoosh
fitsync list

# Download a specific activity
fitsync download <activity-id> --output ./downloads/

# Upload to Intervals.icu
fitsync upload <file.fit>

# Sync: download from MyWhoosh, upload to Intervals
fitsync sync --since 2026-01-01
```

## Project Structure

```
spike/
├── FitSync.Cli/
│   ├── Program.cs
│   ├── Commands/
│   │   ├── ConfigCommand.cs
│   │   ├── ListCommand.cs
│   │   ├── DownloadCommand.cs
│   │   ├── UploadCommand.cs
│   │   └── SyncCommand.cs
│   ├── Services/
│   │   ├── MyWhooshClient.cs
│   │   ├── IntervalsIcuClient.cs
│   │   └── ConfigStore.cs
│   └── FitSync.Cli.csproj
└── FitSync.Cli.sln
```

## Implementation Plan

### Step 1: Scaffold CLI (30 min)
```bash
dotnet new console -n FitSync.Cli
cd FitSync.Cli
dotnet add package System.CommandLine
dotnet add package Microsoft.Extensions.Http
dotnet add package System.Text.Json
```

### Step 2: Investigate MyWhoosh API (1 hour)
- Clone/read https://github.com/mywhoosh-community/mywhoosh-api
- Understand auth flow
- Find endpoints for:
  - Login
  - List activities
  - Download FIT file

### Step 3: Build MyWhooshClient (1 hour)
```csharp
public class MyWhooshClient
{
    public async Task<bool> LoginAsync(string email, string password);
    public async Task<List<Activity>> ListActivitiesAsync(DateTime? since = null);
    public async Task<Stream> DownloadFitAsync(string activityId);
}
```

### Step 4: Build IntervalsIcuClient (1 hour)
```csharp
public class IntervalsIcuClient
{
    // API Key auth via Basic Auth: "API_KEY" as username, api key as password
    public async Task<UploadResult> UploadFitAsync(string athleteId, Stream fitFile, string filename);
}
```

Based on cookbook:
```
POST https://intervals.icu/api/v1/athlete/{id}/activities
Authorization: Basic base64(API_KEY:{api_key})
Content-Type: multipart/form-data
```

### Step 5: Wire up CLI commands (30 min)
- `list` → MyWhooshClient.ListActivitiesAsync
- `download` → MyWhooshClient.DownloadFitAsync
- `upload` → IntervalsIcuClient.UploadFitAsync
- `sync` → Combine download + upload

### Step 6: Test end-to-end (30 min)
- Complete one full sync cycle
- Verify activity appears in Intervals.icu

## Configuration Storage

Simple JSON file in user profile:
```
~/.fitsync/config.json

{
  "mywhoosh": {
    "email": "...",
    "password": "..."
  },
  "intervals": {
    "apiKey": "...",
    "athleteId": "i12345"
  }
}
```

## Success Criteria

- [ ] Can login to MyWhoosh
- [ ] Can list activities from MyWhoosh
- [ ] Can download FIT file from MyWhoosh
- [ ] Can upload FIT file to Intervals.icu
- [ ] Complete sync of one activity works end-to-end

## Notes

### MyWhoosh API Investigation

[Fill in as you explore the repo]

**Auth method:**
```
[Document how auth works]
```

**List activities endpoint:**
```
[Document endpoint and response]
```

**Download FIT endpoint:**
```
[Document endpoint]
```

### Intervals.icu API Notes

From the cookbook:
- Base URL: `https://intervals.icu`
- Auth: Basic auth with `API_KEY` as username
- Upload: `POST /api/v1/athlete/{id}/activities` with multipart form

```bash
# Example curl
curl -X POST "https://intervals.icu/api/v1/athlete/i12345/activities" \
  -u "API_KEY:your-api-key-here" \
  -F "file=@activity.fit"
```

## Learnings

[Document what you learn as you build]

## Next Steps After Spike

If successful:
1. Move CLI code into proper solution structure
2. Add Azure Storage for persistence
3. Add AI summary generation
4. Build Blazor web UI
