# Spike: Garmin Connect Data Export

## Time Box
**Allocated:** 4 hours  
**Status:** Not started

## Objective

### Questions to Answer
1. [ ] Can we use the web API without official partnership?
2. [ ] How does Garmin authenticate? (OAuth, SSO, cookies?)
3. [ ] Where are activities listed?
4. [ ] Can we download original FIT files (not just GPX)?
5. [ ] Is there anti-automation detection?

### Success Criteria
- [ ] Can list recent activities from Garmin Connect
- [ ] Can download at least one FIT file
- [ ] Understand auth flow for browser extension capture

## Background / Prior Art

### Known Information
- Garmin Connect: connect.garmin.com
- Official API exists but requires partnership (Health API, Wellness API)
- Many third-party tools work somehow (garmin-connect-export, etc.)
- Garmin uses complex SSO flow (SSO ticket system)
- Session cookies are HTTPOnly, harder to capture

### Potential Approaches

**Approach A: Browser API (connect.garmin.com)**
```
1. Log into connect.garmin.com
2. Dev tools → find activity list calls
3. Look for connectapi.garmin.com endpoints
4. Capture cookies/tokens
```

**Approach B: garth Library (Python)**
```
Existing Python library that handles Garmin auth:
https://github.com/matin/garth

Study how they:
- Handle SSO flow
- Get OAuth1/OAuth2 tokens
- Download FIT files
```

**Approach C: garmin-connect-export**
```
Another existing tool:
https://github.com/petergardfjall/garmin-connect-export

Python script that downloads all activities
Study auth mechanism
```

### Known Endpoints (from prior research)

```
SSO: sso.garmin.com/sso/signin
API Base: connectapi.garmin.com

List activities:
GET /activitylist-service/activities/search/activities
    ?start=0&limit=20

Download FIT:
GET /download-service/files/activity/{activityId}
```

### Auth Flow (Complex!)

```
1. GET sso.garmin.com/sso/signin → Get CSRF token + cookies
2. POST username/password with CSRF → Get SSO ticket
3. Exchange ticket for session cookies
4. Use cookies for API calls
```

This is harder than TrainingPeaks/MyWhoosh because:
- Multi-step SSO
- CSRF protection
- Cookies are HTTPOnly (can't read from JS easily)

## Research Notes

### Browser Extension Approach
Can we capture the final session cookie from browser?
- [ ] Check if Authorization header is used (easier)
- [ ] Check if cookies are HTTPOnly (harder)
- [ ] Look at requests to connectapi.garmin.com

### Garth Library Analysis
```python
# garth handles the complex auth flow
# Study: https://github.com/matin/garth/blob/main/garth/sso.py

# Key insight: They store OAuth tokens that can be refreshed
# If we can capture these, we don't need to re-auth often
```

## Risks

| Risk | Mitigation |
|------|------------|
| Complex SSO flow | Study garth library, replicate in F# |
| HTTPOnly cookies | May need to intercept at request level in extension |
| Rate limiting | Garmin is known to rate limit, go slow |
| Account lockout | Don't hammer during testing |
| SSO changes break us | Garmin updates auth frequently |

## Spike Output

### If Successful
- Add `Garmin.fs` to spike project
- Document SSO flow in ADR-014
- May need custom extension handling for HTTPOnly cookies

### If Blocked
- Many Garmin users also sync to Strava
- Can use Strava as intermediary source
- Document limitations

## References
- https://connect.garmin.com
- https://github.com/matin/garth (Python auth library)
- https://github.com/petergardfjall/garmin-connect-export
- https://github.com/cyberjunky/python-garminconnect
