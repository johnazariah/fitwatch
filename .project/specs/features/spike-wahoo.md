# Spike: Wahoo Data Export

## Time Box
**Allocated:** 4 hours  
**Status:** Not started

## Objective

### Questions to Answer
1. [ ] Where does Wahoo store workout data? (Cloud, device, both?)
2. [ ] How does Wahoo cloud authenticate?
3. [ ] Can we access SYSTM workouts? ELEMNT ride data?
4. [ ] What format are files in? (FIT, proprietary?)
5. [ ] Is there a web interface or only mobile app?

### Success Criteria
- [ ] Can list recent activities from Wahoo cloud
- [ ] Can download at least one FIT file
- [ ] Understand data sources (SYSTM vs ELEMNT vs KICKR)

## Background / Prior Art

### Wahoo Ecosystem
```
Wahoo has multiple products:

1. ELEMNT bike computers (BOLT, ROAM)
   - Sync to Wahoo app
   - Auto-upload to Strava/TP if configured
   - FIT files on device

2. KICKR trainers
   - Controlled by apps (Wahoo, Zwift, TrainerRoad, etc.)
   - FIT files if using Wahoo app

3. SYSTM training app (formerly Sufferfest)
   - Structured workouts
   - Stores workout history
   - Subscription service
```

### Known Information
- Web interface: https://cloud.wahoofitness.com (?)
- SYSTM: https://systm.wahoofitness.com
- Most users sync ELEMNT → Strava automatically
- No known public API

### Potential Approaches

**Approach A: Wahoo Cloud Web**
```
1. Find web login (cloud.wahoofitness.com?)
2. Log in, intercept API calls
3. Capture auth token
4. List and download activities
```

**Approach B: SYSTM Web**
```
1. Log into systm.wahoofitness.com
2. Find workout history API
3. Download workout FIT files
```

**Approach C: ELEMNT App API**
```
1. Proxy ELEMNT mobile app
2. Intercept API calls
3. Replicate in our code
```

**Approach D: Direct Device Sync**
```
ELEMNT devices expose files via WiFi/USB
- Connect to ELEMNT's WiFi hotspot
- Access FIT files directly
- No cloud needed

Downside: Requires physical device access
```

## Research Notes

### Wahoo Cloud Discovery
```
Check these URLs:
- https://cloud.wahoofitness.com
- https://api.wahoofitness.com
- https://systm.wahoofitness.com

Look for:
- Login flow
- Activity list
- Download endpoints
```

### SYSTM-Specific
```
SYSTM is a subscription training platform
- Has structured workouts with videos
- Tracks workout history
- May have separate API from ELEMNT data
```

### Wahoo + TrainingPeaks
```
Many SYSTM users connect to TrainingPeaks
- If we already have TP integration, might get Wahoo data that way
- Check if TP shows source = "Wahoo" or "SYSTM"
```

## Risks

| Risk | Mitigation |
|------|------------|
| No web interface | May need to intercept mobile app |
| Data split across services | Focus on one (ELEMNT or SYSTM) first |
| Users already sync elsewhere | Wahoo→Strava→us works, just extra hop |
| Subscription required for SYSTM | Need active SYSTM subscription to test |

## Spike Output

### If Successful
- Add `Wahoo.fs` to spike project
- Document which Wahoo service(s) we support
- Update browser extension for wahoo domains

### If Blocked
- Wahoo users typically auto-sync to Strava/TrainingPeaks
- We can pull from those sources instead
- Document as "indirect source"

## Questions for Users
- Do you use ELEMNT, KICKR, or SYSTM?
- Do you sync to Strava/TP automatically?
- Would you need direct Wahoo export, or is Strava fine?

## References
- https://cloud.wahoofitness.com (verify exists)
- https://systm.wahoofitness.com
- https://wahoofitness.com
- Reddit r/wahoofitness for user experiences
