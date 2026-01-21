# Spike: Polar Flow Data Export

## Time Box
**Allocated:** 3 hours  
**Status:** Not started

## Objective

### Questions to Answer
1. [ ] Does Polar have a public API? (AccessLink?)
2. [ ] Can we use the web interface API if no public API?
3. [ ] What authentication method?
4. [ ] Can we download FIT/TCX files?
5. [ ] Any third-party tools we can learn from?

### Success Criteria
- [ ] Can list recent activities from Polar Flow
- [ ] Can download at least one activity file
- [ ] Understand auth flow

## Background / Prior Art

### Polar Ecosystem
```
Polar devices:
- Polar Vantage (V, V2, V3)
- Polar Grit X
- Polar Pacer
- Polar Ignite
- Polar H10 heart rate strap

Polar Flow:
- https://flow.polar.com
- Web and mobile apps
- Training analysis, sleep, etc.
```

### Known API Options

**Option A: Polar AccessLink API (Official)**
```
https://www.polar.com/accesslink-api

- Official partner API
- Requires registration
- May require partnership for full access
- OAuth 2.0

Endpoints:
- /v3/users/{user-id}/exercise-transactions
- /v3/exercises/{exercise-id}
- /v3/exercises/{exercise-id}/fit (FIT download!)
```

**Option B: Polar Flow Web API (Unofficial)**
```
1. Log into flow.polar.com
2. Intercept API calls
3. Same pattern as TrainingPeaks/MyWhoosh
```

### AccessLink API Details
```
Registration: https://admin.polaraccesslink.com

OAuth Flow:
- Authorize: https://flow.polar.com/oauth2/authorization
- Token: https://polaraccesslink.com/v2/oauth2/token

Scopes:
- accesslink.read_all

Rate Limits:
- Unknown, need to test
```

## Research Notes

### Polar AccessLink Registration
```
1. Go to admin.polaraccesslink.com
2. Register as developer
3. Create application
4. Get client_id, client_secret
5. May need approval for production

Check if there's a "hobby" or "personal" tier
```

### FIT File Access
```
According to docs, AccessLink provides:
- GET /v3/exercises/{id}/fit → FIT file download
- GET /v3/exercises/{id}/gpx → GPX file
- GET /v3/exercises/{id}/tcx → TCX file

This is better than Strava - they actually provide original files!
```

### User Flow
```
If official API works:
1. User clicks "Connect Polar" in FitBridge
2. OAuth redirect to Polar
3. User authorizes
4. We get tokens
5. Sync activities

Similar to Strava - proper OAuth, no extension needed.
```

## Risks

| Risk | Mitigation |
|------|------------|
| API requires partnership | Try registering, see what access we get |
| Limited to certain device types | Document which devices supported |
| Lower user base than Garmin/Strava | Lower priority, but good to have |
| API changes | Monitor Polar developer updates |

## Spike Output

### If Official API Works
- Register AccessLink app
- Add `Polar.fs` to spike project
- Document OAuth flow
- Use as example of "proper" integration alongside Strava

### If Blocked
- Fall back to web API scraping (like TrainingPeaks)
- Or mark as unsupported initially

## Priority

Lower priority than Zwift/Garmin/Wahoo, but:
- If AccessLink is easy to register, could be quick win
- Proper OAuth means no extension complexity
- Worth checking registration process early

## References
- https://www.polar.com/accesslink-api
- https://admin.polaraccesslink.com
- https://flow.polar.com
- https://github.com/polarofficial/accesslink-example-python
