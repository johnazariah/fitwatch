# Platform Integration Spikes

> **Goal**: Validate we can pull FIT files from each source platform before committing to full implementation.

## Spike Checklist (per platform)

- [ ] Authentication method discovered
- [ ] List activities endpoint found
- [ ] Download FIT file endpoint found
- [ ] Token lifetime understood
- [ ] Rate limits documented (if any)
- [ ] Working code in spike folder

---

## Platform Status

| Platform | Status | Auth Method | FIT Available | Extension Needed | Spike Doc |
|----------|--------|-------------|---------------|------------------|-----------|
| MyWhoosh | ✅ Complete | Bearer token | ✅ Yes | ✅ Yes | [complete](../research/mywhoosh-spike.md) |
| TrainingPeaks | ✅ Complete | Bearer token | ✅ Yes | ✅ Yes | [complete](../research/trainingpeaks-intervals-sync-spike.md) |
| iGPSport | ✅ Complete | Bearer JWT | ✅ Yes (Aliyun OSS) | ✅ Yes | [spike-igpsport.md](spike-igpsport.md) |
| Zwift | 📋 Not started | TBD | TBD | Likely | [spike-zwift.md](spike-zwift.md) |
| Garmin Connect | 📋 Not started | Complex SSO | ✅ Yes | Likely | [spike-garmin.md](spike-garmin.md) |
| Wahoo | 📋 Not started | TBD | TBD | TBD | [spike-wahoo.md](spike-wahoo.md) |
| Strava | 📋 Not started | OAuth 2.0 ✨ | ❌ Streams only | ❌ No | [spike-strava.md](spike-strava.md) |
| Polar Flow | 📋 Not started | OAuth 2.0 ✨ | ✅ Yes | ❌ No | [spike-polar.md](spike-polar.md) |

✨ = Official API with proper OAuth (no extension needed)

---

## Priority Order

| Priority | Platform | Rationale |
|----------|----------|-----------|
| 1 | **iGPSport** | Real user pain, you own the device, can test immediately |
| 2 | **Zwift** | Large user base, no easy export, high value |
| 3 | **Garmin** | Most common device ecosystem |
| 4 | **Strava** | Official API, easy win, but no FIT files |
| 5 | **Wahoo** | Popular with serious cyclists |
| 6 | **Polar** | Official API available, lower user base |

---

## Integration Complexity Matrix

```
                    Easy                              Hard
                      │                                 │
    Official API ─────┼─────────────────────────────────┼───── No API
                      │                                 │
         ┌────────────┴────────────┐     ┌──────────────┴────────────┐
         │  Strava    Polar        │     │  TrainingPeaks  MyWhoosh  │
         │  (OAuth)   (AccessLink) │     │  (Bearer)       (Bearer)  │
         └─────────────────────────┘     └───────────────────────────┘
                      │                                 │
                      │                   ┌─────────────┴─────────────┐
                      │                   │  Garmin      Zwift        │
                      │                   │  (SSO)       (Unknown)    │
                      │                   └───────────────────────────┘
```

---

## Browser Extension Requirements by Platform

| Platform | Capture Method | Domains to Monitor |
|----------|---------------|-------------------|
| TrainingPeaks | Authorization header | tpapi.trainingpeaks.com |
| MyWhoosh | Authorization header | service14.mywhoosh.com |
| iGPSport | TBD | cloud.igpsport.com (?) |
| Zwift | TBD | my.zwift.com, api.zwift.com (?) |
| Garmin | Cookies (HTTPOnly) | connect.garmin.com, connectapi.garmin.com |
| Wahoo | TBD | cloud.wahoofitness.com (?) |
| Strava | N/A - OAuth | N/A |
| Polar | N/A - OAuth | N/A |

---

## Time Box

Each spike: **2-4 hours max**

If we can't figure out auth + list + download in 4 hours, document blockers and move on.

---

## Fallback Strategy

For platforms we can't integrate directly:

```
User's Device → Auto-sync to Strava → FitBridge pulls from Strava
```

Most devices/apps support Strava auto-upload. If direct integration fails:
1. Mark platform as "indirect support"
2. Guide users to enable Strava sync on their device
3. Pull from Strava (official API)

Downside: Extra hop, user needs Strava account, no original FIT files.

