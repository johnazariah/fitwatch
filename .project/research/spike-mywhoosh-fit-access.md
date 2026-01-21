# Research Spike: MyWhoosh FIT File Access

## Status: IN PROGRESS

## Time Box
**Allocated:** 4 hours  
**Started:** 2026-01-20  
**Completed:** [Fill in when done]

## Key Discoveries

### MyWhoosh API
**Found:** https://github.com/mywhoosh-community/mywhoosh-api

Community-maintained API client. Need to investigate:
- Authentication method
- Available endpoints
- FIT file download capability

### Intervals.icu API
**Found:** https://forum.intervals.icu/t/intervals-icu-api-integration-cookbook/80090

Official cookbook with examples. Well-documented API with API key auth.

## Objective

Build a working CLI tool that downloads FIT files from MyWhoosh and uploads to Intervals.icu.

### Questions to Answer

1. [ ] Does MyWhoosh have a public API?
2. [ ] Can we download FIT files via the website (scraping)?
3. [ ] Does the MyWhoosh desktop/mobile app save files locally?
4. [ ] Is there OAuth or another auth mechanism?
5. [ ] Are there rate limits or terms of service concerns?
6. [ ] What data is in the FIT files? (power, HR, cadence, GPS?)

### Success Criteria

- [ ] Can authenticate programmatically (or know it's impossible)
- [ ] Can list activities for a user
- [ ] Can download a FIT file for a specific activity
- [ ] Have working proof-of-concept code
- [ ] Know the limitations/risks

## Research Areas

### 1. Official API

**Check:**
- [ ] MyWhoosh developer documentation
- [ ] API references in their website/app
- [ ] Developer programs or partner access

**URLs to investigate:**
- https://mywhoosh.com (main site)
- Look for /api/, /developer/, /partners/ pages
- Check mobile app for API calls (proxy/network inspection)

**Notes:**
```
[Document findings here]
```

### 2. Website Analysis

**Check:**
- [ ] Login flow (what auth mechanism?)
- [ ] Activity list page (is there one?)
- [ ] Download buttons (can you export FIT?)
- [ ] Network requests when viewing activities

**Tools:**
- Browser DevTools (Network tab)
- Look for XHR/fetch requests to APIs

**Notes:**
```
[Document findings here]
```

### 3. Mobile/Desktop App

**Check:**
- [ ] Does MyWhoosh have a companion app?
- [ ] Where are ride files saved locally?
- [ ] Does it sync to a folder we can watch?
- [ ] Any Garmin/Wahoo integration that exports FIT?

**Common paths:**
- Windows: `%USERPROFILE%\Documents\MyWhoosh\`
- macOS: `~/Documents/MyWhoosh/`

**Notes:**
```
[Document findings here]
```

### 4. Third-Party Integrations

**Check:**
- [ ] Does MyWhoosh connect to Strava? (We could reverse-engineer)
- [ ] Does it connect to Garmin Connect?
- [ ] Any Zapier/IFTTT integrations?
- [ ] Community tools/scripts on GitHub?

**Search:**
```
site:github.com mywhoosh
site:reddit.com mywhoosh api
site:reddit.com mywhoosh export
```

**Notes:**
```
[Document findings here]
```

### 5. Community Knowledge

**Check:**
- [ ] MyWhoosh forums/Discord
- [ ] Reddit r/MyWhoosh or r/cycling
- [ ] Facebook groups
- [ ] Zwift/indoor cycling forums

**Notes:**
```
[Document findings here]
```

## Findings

### Question 1: Does MyWhoosh have a public API?

**Answer:** [Yes/No/Partial]

**Details:**
```
[Document what you found]
```

---

### Question 2: Can we download FIT files via website?

**Answer:** [Yes/No/Partial]

**Details:**
```
[Steps to download, or why it's not possible]
```

---

### Question 3: Local file storage?

**Answer:** [Yes/No/Partial]

**Details:**
```
[Where files are saved, format, etc.]
```

---

### Question 4: Authentication mechanism?

**Answer:** [OAuth/Cookie/API Key/Unknown]

**Details:**
```
[How auth works, token format, expiry, etc.]
```

---

### Question 5: Rate limits / TOS?

**Answer:** [Known limits/TOS concerns]

**Details:**
```
[Any restrictions we should be aware of]
```

---

### Question 6: FIT file contents?

**Answer:** [What data is available]

**Details:**
```
[Power, HR, cadence, GPS, etc.]
```

---

## Proof of Concept

### What We Built

[Description of any test code]

### Code Location

`.project/prototypes/mywhoosh-spike/`

### Sample Code

```python
# or C#, whatever works for testing
# Paste working code here
```

### Sample Response

```json
{
  "// paste example API response or file contents": ""
}
```

## Recommendations

Based on findings, recommend one of:

### Option A: [Name the approach]
**How it works:**
**Pros:**
**Cons:**
**Effort:**

### Option B: [Alternative approach]
**How it works:**
**Pros:**
**Cons:**
**Effort:**

### Recommended Approach

[Which option and why]

## Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| No API exists | ? | High | Fall back to local file watch |
| Auth is complex | ? | Medium | Cookie-based session? |
| TOS prohibits scraping | ? | High | Contact MyWhoosh, or manual export |

## Action Items

Coming out of this spike:

- [ ] [Update feature spec with findings]
- [ ] [Create MyWhoosh source adapter based on approach]
- [ ] [Document any credentials/setup needed]

## References

- [MyWhoosh Website](https://mywhoosh.com)
- [Link to any API docs found]
- [Link to community discussions]
- [Link to any GitHub repos found]
