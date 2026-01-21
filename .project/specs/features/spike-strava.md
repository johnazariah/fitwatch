# Spike: Strava Data Export

## Time Box
**Allocated:** 2 hours (should be straightforward - official API)
**Status:** Not started

## Objective

### Questions to Answer
1. [ ] What OAuth scopes do we need for activity read + FIT download?
2. [ ] Can we get original FIT files or only Strava's processed data?
3. [ ] What are the rate limits?
4. [ ] How long do tokens last? Refresh token flow?
5. [ ] Any restrictions on what we can do with the data?

### Success Criteria
- [ ] Register a Strava API application
- [ ] Complete OAuth flow
- [ ] List recent activities
- [ ] Download activity data (FIT or streams)

## Background / Prior Art

### Known Information
- Strava has a well-documented public API
- OAuth 2.0 authentication
- Rate limits: 100 requests/15 minutes, 1000/day (per app)
- Need to register app at strava.com/settings/api
- **This is the "easy" integration** - official API, proper OAuth

### API Documentation
```
Base URL: https://www.strava.com/api/v3

Key Endpoints:
- GET /athlete/activities - List activities
- GET /activities/{id} - Activity details
- GET /activities/{id}/streams - Detailed data streams

OAuth:
- Authorize: https://www.strava.com/oauth/authorize
- Token: https://www.strava.com/oauth/token
```

### Scopes
```
read              - Read public data
read_all          - Read private activities
activity:read     - Read activity data
activity:read_all - Read all activities including private
```

We need: `activity:read_all` for full access.

## Research Notes

### FIT File Access - The Catch

**Strava does NOT provide original FIT files via API.**

Options:
```
1. Export streams as pseudo-FIT
   - Get latlng, heartrate, power, cadence streams
   - Reconstruct FIT-like data
   - Lose some metadata

2. Original upload preserved?
   - Users can manually download original from web
   - API endpoint may exist but undocumented
   - Check: GET /activities/{id}/export_original

3. Use Strava as metadata only
   - List activities from Strava
   - Match to original source (Garmin, Wahoo)
   - Download FIT from original source
```

### Rate Limiting Strategy
```
100 requests / 15 minutes = ~6.6/min

For a user with 500 activities:
- List: 5 requests (100 per page)
- Details: 500 requests → 75 minutes to backfill

Strategy:
- Incremental sync (only new activities)
- Cache activity IDs we've seen
- Respect rate limits with delays
```

### Webhook Support
```
Strava supports webhooks for new activities!
- Register subscription for athlete
- Get notified when new activity uploaded
- No polling needed for incremental sync

This is ideal for FitBridge real-time sync.
```

## Implementation Notes

### OAuth Flow for FitBridge

```
1. User clicks "Connect Strava" in FitBridge
2. Redirect to Strava OAuth with our client_id
3. User authorizes
4. Strava redirects to our callback with code
5. Exchange code for access_token + refresh_token
6. Store tokens in Key Vault

No browser extension needed - proper OAuth!
```

### Token Refresh
```
Access tokens expire in 6 hours
Refresh tokens are long-lived

On API call:
- If 401, refresh token
- Retry request
```

## Risks

| Risk | Mitigation |
|------|------------|
| No original FIT files | Reconstruct from streams, or use as metadata source |
| Rate limits during backfill | Slow backfill, prioritize recent activities |
| App approval required? | Start with personal app, apply for approval if scaling |
| Terms of Service | Read carefully - can we store/redistribute data? |

## Spike Output

### If Successful
- Register Strava app (dev credentials)
- Add `Strava.fs` to spike project
- Document OAuth flow
- Test webhook subscription

### Strava as "Easy Win"
This should be our first fully "proper" integration:
- Official API
- Standard OAuth
- Webhooks for real-time
- No browser extension gymnastics

Good reference implementation for how integrations *should* work.

## References
- https://developers.strava.com
- https://developers.strava.com/docs/reference/
- https://developers.strava.com/docs/webhooks/
- https://www.strava.com/settings/api (app registration)
