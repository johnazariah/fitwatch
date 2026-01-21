# ADR-005: Platform Integration Policies

## Status
Accepted

## Context
Different fitness platforms have vastly different policies on data portability and API usage. We need a clear framework for how we integrate with each platform.

## Decision

**Classify platforms into tiers based on data freedom policies. Prioritize open platforms.**

### Platform Tiers

#### Tier 1: Open Platforms (Full Integration)
Platforms that respect user data ownership and provide unrestricted FIT file access.

| Platform | FIT Download | API Upload | Notes |
|----------|--------------|------------|-------|
| **Garmin Connect** | ✅ API & Export | ✅ | Explicitly supports data portability |
| **Wahoo** | ✅ API & Export | ✅ | User-friendly policies |
| **Intervals.icu** | ✅ | ✅ | Developer-friendly, open philosophy |
| **Local Files** | ✅ | N/A | User's own filesystem |

#### Tier 2: Friendly Platforms (Import Preferred)
Platforms with reasonable policies but some limitations.

| Platform | FIT Download | API Upload | Notes |
|----------|--------------|------------|-------|
| **TrainingPeaks** | ⚠️ Export only | ✅ | Good upload API, manual export |
| **Zwift** | ⚠️ Companion app | ❌ | Files available locally after ride |
| **MyWhoosh** | ⚠️ TBD | ❌ | Need to research API/export |
| **TrainerRoad** | ⚠️ Export only | ❌ | Manual export available |

#### Tier 3: Restricted Platforms (Proceed with Caution)
Platforms with hostile data portability policies.

| Platform | FIT Download | API Upload | Notes |
|----------|--------------|------------|-------|
| **Strava** | ❌ API blocked | ✅ Upload only | API TOS prohibits data export |

### Integration Strategies by Tier

#### Tier 1: Full bidirectional sync
```
User authorizes → We poll for new activities → Download FIT → Store
User records locally → We upload FIT → Platform has copy
```

#### Tier 2: Import via user action
```
User exports from platform → User imports to our system → Store
User records locally → We upload (if API available)
```

#### Tier 3: Output only, input via manual export
```
Strava: User can upload TO Strava (their choice)
        User can import FROM Strava only via manual bulk export
        We NEVER use Strava API to download activities
```

### Strava Special Handling

Given Strava's market position, we handle it specially:

**What we WILL do:**
- Accept manual Strava bulk exports (user exercises GDPR rights)
- Upload TO Strava if user requests (with provenance check)
- Be transparent that Strava restricts data portability

**What we WON'T do:**
- Use Strava API to download activities
- Pretend Strava integration is "just like" open platforms
- Hide the restrictions from users

**User messaging:**
```
ℹ️ Strava restricts data portability in their API terms.

To import your Strava history:
1. Request your data export at strava.com/athlete/delete_your_account
2. Download the ZIP file (this is YOUR data!)
3. Import the ZIP here

Strava cannot prevent you from exporting your own data under GDPR/CCPA.
```

## Platform Policy Registry

Maintain a living document of platform policies:

```
.project/
└── research/
    └── platform-policies/
        ├── garmin.md
        ├── wahoo.md
        ├── strava.md
        ├── intervals-icu.md
        └── ...
```

Each file documents:
- Current TOS version and date reviewed
- Data export capabilities
- API restrictions
- User data rights
- Our integration approach

## Consequences

### Positive
- Clear framework for evaluating new platforms
- Legal clarity on what we will/won't do
- User education on data freedom
- Positions us as a "data rights" advocate

### Negative
- Some users want "just connect Strava" simplicity
- We may be seen as anti-Strava (we're pro-user)
- More complex onboarding for restricted platforms

### Marketing Opportunity
This positions us as the **ethical alternative**:
- "Your workouts. Your data. Your choice."
- "Break free from platform lock-in"
- "We believe you own your fitness data"

## Related Decisions
- ADR-003: FIT Files as Canonical Format
- ADR-004: Data Provenance Tracking
