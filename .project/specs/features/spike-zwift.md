# Spike: Zwift Data Export

## Time Box
**Allocated:** 4 hours  
**Status:** 🔄 In Progress

## Objective

### Questions to Answer
1. [x] How does Zwift authenticate users? → Bearer JWT via Keycloak
2. [ ] Where are activities listed? (need to find profile activities endpoint)
3. [ ] Can we download FIT files? What format?
4. [x] What's the token lifetime? → ~6 hours (short!)
5. [ ] Any rate limiting or anti-bot measures?

### Success Criteria
- [ ] Can list recent activities from Zwift
- [ ] Can download at least one FIT file
- [x] Understand auth flow for browser extension capture

## Background / Prior Art

### Known Information
- Zwift is primarily a game client (Windows/Mac/iOS/Android)
- Web interface at my.zwift.com shows activity history
- Zwift Connect mobile app also shows activities
- No official public API
- Third-party tools exist (ZwiftGPS, etc.) - how do they do it?

### Potential Approaches

**Approach A: my.zwift.com Web API**
```
1. Log into my.zwift.com
2. Open dev tools, find API calls
3. Look for /api/activities or similar
4. Capture bearer token from requests
```

**Approach B: Zwift Game API**
```
1. Zwift game makes API calls during/after rides
2. Use proxy (Fiddler/Charles) to intercept
3. May require game client running
```

**Approach C: Strava Integration**
```
1. Most Zwift users auto-upload to Strava
2. Pull from Strava instead (official OAuth)
3. Downside: Extra hop, user must have Strava connected
```

### Third-Party Tools to Research
- **ZwiftGPS** - exports routes, how do they auth?
- **ZwiftPower** - race results, uses Zwift data
- **WTRL** - team racing, accesses Zwift data
- **Sauce for Zwift** - browser extension for live data

## Research Notes

### API Endpoints - COMPLETE! ✅

```
Auth Server: https://secure.zwift.com/auth/realms/zwift (Keycloak)
API Base: https://us-or-rly101.zwift.com (regional relay - US Oregon)
         May also be: eu-*, ap-* for other regions

List My Activities:
GET /api/activity-feed/feed/?limit=30&includeInProgress=false&feedType=JUST_ME

Activity Detail:
GET /api/activities/{activityId}

FIT File Download:
GET /api/activities/{activityId}/file/{fileId}
  - fileId comes from activity detail: fitnessData.fullDataUrl
  - fullDataUrl = full FIT file
  - smallDataUrl = smaller version (probably GPS only?)

Auth: Bearer JWT
Headers:
  - zwift-api-version: 2.5
  - source: my-zwift
  - Origin: https://www.zwift.com

Token lifetime: ~6 HOURS (short! need refresh token strategy)
```

### Activity Detail Response Structure
```json
{
  "fitnessData": {
    "status": "AVAILABLE",
    "fullDataUrl": "https://us-or-rly101.zwift.com/api/activities/{id}/file/{fileId}",
    "smallDataUrl": "https://us-or-rly101.zwift.com/api/activities/{id}/file/{fileId2}"
  }
}
```

### JWT Token Analysis
```json
{
  "iss": "https://secure.zwift.com/auth/realms/zwift",
  "sub": "a7063135-e736-4f13-8e88-0f23565a946c",  // User UUID
  "exp": 1768983842,  // ~6 hour lifetime
  "aud": ["Zwift REST API -- production", ...],
  "realm_access": { "roles": ["trial-subscriber", "beta-tester"] },
  "name": "John Az",
  "preferred_username": "john.azariah.float@gmail.com"
}
```

### Token Capture Points
- [x] zwift.com - Authorization header ✅
- [ ] Zwift Companion app - Bearer token?
- [ ] Desktop client - Local files/API calls?

## Risks

| Risk | Mitigation |
|------|------------|
| Zwift actively blocks scrapers | May need to rate limit, use realistic user-agent |
| Auth requires game client | Fallback to Strava source |
| Token expires very quickly | Need frequent re-auth via extension |
| Legal/ToS concerns | Users are exporting their own data |

## Spike Output

### If Successful
- Add `Zwift.fs` to spike project
- Document endpoints in ADR-014
- Update browser extension permissions

### If Blocked
- Document why
- Recommend Strava as fallback source for Zwift activities
- Consider reaching out to Zwift for API partnership (long shot)

## References
- https://my.zwift.com
- https://zwiftpower.com
- Reddit r/Zwift threads on data export
