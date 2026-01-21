# Spike: iGPSport Data Export

## Time Box
**Allocated:** 3 hours  
**Status:** ✅ Complete

## Objective

### Questions to Answer
1. [x] Does iGPSport have a web interface? → YES: https://app.igpsport.com
2. [x] What authentication method do they use? → Bearer JWT (~7 day lifetime)
3. [x] Can we download FIT files directly from their cloud? → YES: via Aliyun OSS URL in activity detail
4. [x] Is data only accessible via mobile app? → NO, full web API available
5. [x] Any undocumented API we can discover? → YES, full REST API discovered!

### Success Criteria
- [x] Can list recent activities from iGPSport → `list-igp` command
- [x] Can download at least one FIT file → `fetchActivity` function
- [x] Understand the data flow → device → app → cloud (Aliyun OSS) → web API

## Background / Prior Art

### The Problem
```
iGPSport bike computers (iGS630, iGS530, etc.):
- Sync to iGPSport mobile app
- Can auto-upload to Strava
- That's it. No other export options.
- No manual FIT file download from app/web
- User is locked into Strava as the only destination
```

### User Pain Points
- Can't sync to Intervals.icu directly
- Can't sync to TrainingPeaks directly
- Must go through Strava (if they even have Strava)
- Strava free tier has limitations
- Original FIT files trapped in iGPSport ecosystem

### iGPSport Ecosystem
```
Devices:
- iGS630 (flagship GPS computer)
- iGS530
- iGS320
- Various other models

Software:
- iGPSport mobile app (iOS/Android)
- Possibly web interface?

Known Integrations:
- Strava (one-way upload)
- TrainingPeaks (maybe? need to verify)
- Komoot (maybe?)
```

## Research Notes

### Web Interface ✅ CONFIRMED
```
URL: https://app.igpsport.com/sport/history/list?lang=en

This is great news - there IS a web interface!
Now we need to:
1. Log in with dev tools open
2. Capture API calls
3. Find activity list + download endpoints
```

### API Endpoints ✅ DISCOVERED

```
Base URL: https://prod.en.igpsport.com
Auth: Bearer JWT (same pattern as TrainingPeaks!)

List Activities:
GET /service/web-gateway/web-analyze/activity/queryMyActivity
    ?pageNo=1
    &pageSize=20
    &reqType=0
    &sort=1          (1 = newest first)

Activity Detail:
GET /service/web-gateway/web-analyze/activity/queryActivityDetail/{activityId}

FIT Download:
✅ URL is in the activity detail response!
Field: "fitUrl"
Example: https://qw20191008.oss-us-west-1.aliyuncs.com/{guid}
Storage: Aliyun OSS (Alibaba Cloud) - US West region
Auth: None needed - direct download from cloud storage
```

### JWT Token Analysis
```
Issuer: http://authserver
Lifetime: ~7 days (exp - nbf)
Client: qiwu.mobile
Scopes: activity.api, device.api, mobile.api, user.api, offline_access
MemberId: extracted from token claims (user identifier)

Key claim: "sub" or "memberid" = user ID
```

### Request Headers
```
Accept: application/json
Authorization: Bearer <jwt>
Origin: https://app.igpsport.com
Referer: https://app.igpsport.com/
qiwu-app-version: 1.0.0
```

### API Gateway Pattern
```
Interesting: They use a service gateway architecture
- /service/web-gateway/web-analyze/activity/...

This suggests microservices behind the gateway.
"qiwu" appears to be the internal codename.
```

## Risks

| Risk | Mitigation |
|------|------------|
| No web interface at all | Intercept mobile app API |
| API requires mobile app signature | May need to reverse engineer |
| Chinese-first company | API docs (if any) may be in Chinese |
| Small market = less prior art | Less community knowledge to draw from |
| Device direct access only | Document USB/file workflow as fallback |

## Spike Output

### ✅ Implementation Complete!

Added `IGPSport.fs` to spike project with:
- `listActivities` - paginated activity list
- `listAllActivities` - fetch all activities  
- `getActivityDetail` - get detail including FIT URL
- `downloadFitFile` - download from Aliyun OSS
- `fetchActivity` - full SourceActivity with FIT data
- `toMetadata` - convert to domain model

CLI Commands:
- `login-igp` - Browser-based auth flow
- `list-igp` - List recent activities
- `sync-igp` - Sync to Intervals.icu

Config:
- `config set igpsport:token <jwt>` - Set token manually

### Browser Extension Domain

Add to extension's monitored domains:
```javascript
{
  domain: "prod.en.igpsport.com",
  captureHeader: "Authorization", 
  platform: "igpsport"
}
```

## Fallback: The Strava Route

If direct integration fails, current workaround:
```
iGPSport → Strava → FitBridge → Intervals.icu
```

Problems:
- Requires Strava account
- Strava doesn't give us original FIT files
- Extra hop adds latency
- Strava free tier limitations

This is exactly why we want direct integration.

## USB/Direct File Access (Backup Plan)

If all else fails, document the manual workflow:
```
1. Connect iGPSport device via USB
2. Device mounts as mass storage
3. Navigate to /Activities or /FIT folder
4. Copy .fit files
5. Upload to FitBridge via web interface

FitBridge provides:
- Drag-and-drop upload page
- Bulk upload support
- Same deduplication logic
- Sync to configured sinks
```

Not as slick as automatic sync, but:
- User owns their data
- Works with any device that stores FIT files
- No dependency on iGPSport's cloud

## Priority

**HIGH** - This is a real user pain point you've experienced. 

If we can crack iGPSport, it validates FitBridge for the "forgotten" device ecosystem—not just the big players.

## References
- https://www.igpsport.com
- iGPSport app on iOS/Android stores
- Reddit/forums for iGPSport users
- DC Rainmaker reviews (may have technical details)
