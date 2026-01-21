# ADR-014: Platform API Integration Patterns

## Status
Accepted

## Context

During the TrainingPeaks and MyWhoosh integration spikes, we discovered that fitness platform APIs share common patterns but have significant differences in authentication, endpoint structure, and data access. Documenting these patterns helps us integrate new platforms faster.

## Decision

### Common Integration Pattern

Each platform integration follows a three-step discovery process:

1. **Authentication capture** - Get tokens from browser dev tools
2. **API endpoint discovery** - Intercept browser requests to find endpoints
3. **Data structure mapping** - Map platform types to domain model

### Platform-Specific Learnings

#### MyWhoosh

| Aspect | Details |
|--------|---------|
| Auth | Cookie-based, capture `accessToken` from browser |
| Base URL | `https://service14.mywhoosh.com` |
| List endpoint | `GET /webapi/Activity/GetActivitiesFromStart?whooshId={id}&start=0&count=100` |
| Download endpoint | `GET /webapi/Activity/GetFitFile?activityId={id}` |
| Gotcha | FIT files are gzip-compressed, need decompression |

#### TrainingPeaks

| Aspect | Details |
|--------|---------|
| Auth | Bearer token, capture from browser Authorization header |
| Base URL | `https://tpapi.trainingpeaks.com` |
| List endpoint | `GET /fitness/v6/athletes/{athleteId}/workouts/{startDate}/{endDate}` |
| Details endpoint | `GET /fitness/v6/athletes/{athleteId}/workouts/{workoutId}/details` |
| Download endpoint | `GET /fitness/v6/athletes/{athleteId}/workouts/{workoutId}/rawfiledata/{fileId}` |
| Gotcha | File ID not in list response - must fetch `/details` first to get `workoutDeviceFileInfos[0].fileId` |

#### Intervals.icu

| Aspect | Details |
|--------|---------|
| Auth | Basic auth with `API_KEY:{key}` |
| Base URL | `https://intervals.icu/api/v1` |
| List endpoint | `GET /athlete/{athleteId}/activities?oldest={date}&newest={date}` |
| Upload endpoint | `POST /athlete/{athleteId}/activities` (multipart form) |
| Gotcha | None - well-documented public API |

### Token Refresh Strategy

| Platform | Token Lifetime | Refresh Strategy |
|----------|---------------|------------------|
| MyWhoosh | ~hours | Re-capture from browser |
| TrainingPeaks | ~hours | Re-capture from browser |
| Intervals.icu | Permanent | API key doesn't expire |

### File ID Indirection Pattern

TrainingPeaks taught us that some platforms require multiple API calls to download a file:

```
1. List activities → [activityId, activityId, ...]
2. Get details(activityId) → { fileId, fileName, ... }
3. Download(activityId, fileId) → bytes
```

This is common when platforms store multiple files per activity (original + processed).

### API Response Debugging

Always log raw JSON for first integration attempt:

```fsharp
if response.IsSuccessStatusCode then
    use doc = JsonDocument.Parse(json)
    let first = doc.RootElement.[0]
    for prop in first.EnumerateObject() do
        printfn "  %s = %s" prop.Name (prop.Value.GetRawText().Substring(0, 50))
```

This saved hours on TrainingPeaks when discovering the `workoutDeviceFileInfos` field.

## Consequences

### Positive
- Documented patterns accelerate new integrations
- Known gotchas prevent wasted debugging time
- Clear separation: auth capture vs API discovery vs data mapping

### Negative
- Token capture from browser is manual and fragile
- No refresh tokens means periodic re-authentication
- Undocumented APIs may change without notice

### Future Platforms - Expected Patterns

| Platform | Expected Auth | Expected Gotchas |
|----------|--------------|------------------|
| Garmin Connect | OAuth 2.0 (official API) | Rate limiting, subscription required |
| Strava | OAuth 2.0 (official API) | Rate limiting, read/write scopes |
| Wahoo | Bearer token | Cloud vs local sync differences |
| Zwift | Cookie-based | No official API, may need scraping |

### Mitigation for Token Expiry

For MVP: Manual token refresh is acceptable.

Post-MVP options:
1. Implement OAuth flows with refresh tokens (requires app registration)
2. Build browser extension to capture tokens automatically
3. Selenium/Playwright automation for token capture
