# ADR-003: FIT Files as Canonical Data Format

## Status
Accepted

## Context
We need to decide on the authoritative source of workout data in our system. Options include:
- Platform-specific formats (Strava activities, Garmin Connect activities)
- Intermediate formats (TCX, GPX)
- Raw device output (FIT files)

## Decision Drivers
- **Data ownership**: Users must unambiguously own their data
- **Legal clarity**: Avoid platform TOS complications
- **Completeness**: Preserve all sensor data without loss
- **Portability**: Work with any platform
- **Longevity**: Format should outlive any single platform

## Decision

**FIT files are the canonical, authoritative format for all workout data.**

### Principles

1. **FIT files are the source of truth**
   - The FIT file from the recording device is the original record
   - All other representations are derived from the FIT file
   - We store and preserve the original FIT file forever

2. **Platforms are temporary homes, not owners**
   - Uploading to Strava/Garmin/etc. is a *copy*, not a transfer
   - The user's FIT file archive is the permanent record
   - Platforms come and go; FIT files remain

3. **Clean provenance through file origin**
   - FIT file from device = unambiguous user ownership
   - FIT file from platform export = user exercising their data rights
   - FIT file from platform API = potentially restricted (avoid)

## Rationale

### Why FIT specifically?

| Format | Completeness | Ownership Clarity | Industry Support |
|--------|--------------|-------------------|------------------|
| **FIT** | ✅ Full sensor data | ✅ Device origin clear | ✅ Universal standard |
| TCX | ⚠️ Loses some data | ⚠️ Often platform-exported | ⚠️ Declining |
| GPX | ❌ GPS only, no power | ⚠️ Often platform-exported | ✅ Universal |
| Platform JSON | ⚠️ Varies | ❌ Platform-created format | ❌ Proprietary |

### Legal clarity

```
Device records FIT → User owns FIT → User uploads to Platform
                  ↓
            User still owns FIT
            Platform has a copy under their TOS
                  ↓
            User can always re-upload their original FIT elsewhere
```

By operating on FIT files (not platform APIs), we avoid:
- API TOS restrictions on data portability
- Questions about who "owns" derived/enriched data
- Platform-specific data format lock-in

### What about activities without FIT files?

Some sources only provide other formats:
- **Accept TCX/GPX as fallback**, but flag as "incomplete"
- **Convert to our internal format**, but note original format
- **Never claim FIT-level completeness** for non-FIT sources

## Consequences

### Positive
- Clear legal standing: operating on user's own files
- Complete data: FIT contains everything the device recorded
- Platform-agnostic: any platform that accepts FIT can be a sink
- Future-proof: not dependent on any platform's API

### Negative
- User must obtain FIT files themselves (can't just "connect" accounts as easily)
- Some platforms make FIT export difficult (but this exposes their lock-in)
- We may need to educate users on how to get their FIT files

### Neutral
- We become a FIT file management tool, not a "sync service"
- Positions us as a data freedom/portability tool

## Implementation Notes

### FIT file acquisition paths (P0)
1. **Direct from device**: User syncs device to folder we watch
2. **Manual upload**: User uploads FIT file through web UI
3. **Cloud storage sync**: Watch Dropbox/OneDrive/GDrive folder

### FIT file acquisition paths (P1)
4. **Platform bulk export**: User requests export from Garmin/etc., imports zip
5. **Friendly platform APIs**: Platforms that explicitly allow FIT download (Garmin, Wahoo)

### FIT file acquisition paths (Avoid)
6. **Hostile platform APIs**: Platforms that restrict data export (Strava)

## Related Decisions
- ADR-004: Platform Integration Policies
- ADR-005: Data Provenance Tracking
