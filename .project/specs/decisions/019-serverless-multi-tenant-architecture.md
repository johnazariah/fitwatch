# ADR-019: Serverless Multi-Tenant Architecture

## Status
Accepted (supersedes ADR-017: Orleans Cloud Architecture)

## Context

FitBridge is a multi-tenant SaaS that syncs fitness activities from platforms without public OAuth APIs to analysis platforms like Intervals.icu. The core USP is a **browser extension that captures authentication tokens** from legitimate user sessions, bridging the moat that these platforms have built.

### The Moat-Crossing Problem

| Platform | Public OAuth | Our Solution |
|----------|--------------|--------------|
| MyWhoosh | ❌ None | Extension captures Bearer token |
| iGPSport | ❌ None | Extension captures Bearer token |
| Zwift | ❌ None (web) | Extension captures JWT (6hr expiry!) |
| TrainingPeaks | ❌ Partner-only | Extension captures session token |
| Strava | ✅ Yes | Could use OAuth, but extension works too |

### Why Orleans is Overkill

ADR-017 proposed Orleans for the backend. Re-evaluating:

| Factor | Orleans | Durable Functions |
|--------|---------|-------------------|
| **Always-on cost** | Yes ($100+/mo minimum) | No (scale-to-zero) |
| **Grain persistence** | Manual setup | Built-in entity storage |
| **Timer/scheduler** | Reminders (need silo running) | Timer triggers (serverless) |
| **F# support** | Good | Excellent |
| **At 100 users** | ~$100/mo | ~$5-15/mo |
| **At 10,000 users** | ~$200/mo | ~$50-100/mo |

For a stealth-mode SaaS targeting $1000/year initial revenue, Orleans' operational overhead and baseline cost is unjustifiable.

## Decision Drivers

- **Browser extension is the USP** - architecture must optimize for token capture flow
- **Scale-to-zero** - pay nothing when no syncs running
- **Multi-tenant from day one** - user isolation, secure token storage
- **Token lifecycle management** - expiry, refresh prompts, platform-specific handling
- **$83/month budget** - must work at 100 users paying $10/year

## Decision

**Azure Functions with Durable Functions for orchestration, Table Storage for data.**

### Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      Browser Extension                           │
│  • Detects API calls to source platforms                        │
│  • Extracts Bearer/JWT tokens from Authorization header         │
│  • Sends to backend: { userId, platform, token, expiresAt }     │
└─────────────────────────┬───────────────────────────────────────┘
                          │ HTTPS POST (user authenticated via extension)
                          ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Azure Functions                               │
│                                                                  │
│  HTTP Triggers:                                                  │
│  ├── POST /api/tokens      ← Extension sends captured tokens    │
│  ├── GET  /api/activities  ← User views sync history            │
│  ├── POST /api/sync/{platform} ← Manual sync trigger            │
│  └── GET  /api/status      ← Dashboard data                     │
│                                                                  │
│  Timer Triggers:                                                 │
│  └── SyncOrchestrator (hourly) ← Syncs all users                │
│                                                                  │
│  Durable Entities:                                               │
│  ├── UserEntity      ← Per-user config and sync state           │
│  └── ActivityEntity  ← Activity register (ADR-018)              │
│                                                                  │
│  Durable Orchestrations:                                         │
│  ├── SyncUserOrchestration    ← Sync one user's platforms       │
│  └── SyncPlatformOrchestration ← Sync one platform              │
└─────────────────────────────────────────────────────────────────┘
                          │
          ┌───────────────┼───────────────┬───────────────┐
          ▼               ▼               ▼               ▼
    ┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐
    │ MyWhoosh │   │ iGPSport │   │  Zwift   │   │Intervals │
    │   API    │   │   API    │   │   API    │   │   icu    │
    └──────────┘   └──────────┘   └──────────┘   └──────────┘
```

### Storage Design

**Table Storage** for cost efficiency ($0.045/GB/month):

```
Tokens Table (encrypted at rest):
┌────────────────┬───────────┬────────────────────────────────┬─────────────┐
│ PartitionKey   │ RowKey    │ Token (encrypted)              │ ExpiresAt   │
│ (UserId)       │ (Platform)│                                │             │
├────────────────┼───────────┼────────────────────────────────┼─────────────┤
│ user_abc123    │ mywhoosh  │ enc(eyJhbGciOi...)             │ 2026-02-01  │
│ user_abc123    │ zwift     │ enc(eyJhbGciOi...)             │ 2026-01-21  │
│ user_abc123    │ igpsport  │ enc(eyJhbGciOi...)             │ 2026-01-28  │
└────────────────┴───────────┴────────────────────────────────┴─────────────┘

Activities Table:
┌────────────────┬────────────────────┬────────────┬───────────┬─────────────┐
│ PartitionKey   │ RowKey             │ Source     │ SourceId  │ SinkId      │
│ (UserId)       │ (StartTime_UUID)   │            │           │             │
├────────────────┼────────────────────┼────────────┼───────────┼─────────────┤
│ user_abc123    │ 2026-01-16_uuid1   │ mywhoosh   │ 12345     │ i118287888  │
│ user_abc123    │ 2024-07-22_uuid2   │ zwift      │ 165863... │ i119406365  │
└────────────────┴────────────────────┴────────────┴───────────┴─────────────┘
```

### Token Lifecycle

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Captured  │────▶│   Active    │────▶│   Expired   │
│  (new token)│     │ (can sync)  │     │(needs recapture)│
└─────────────┘     └─────────────┘     └─────────────┘
                           │                    │
                           │ Sync fails 401     │
                           └───────────────────▶│
                                                │
                                    ┌───────────▼───────────┐
                                    │ Notify user to visit  │
                                    │ platform in browser   │
                                    └───────────────────────┘
```

**Platform-specific handling:**
| Platform | Token TTL | Refresh Strategy |
|----------|-----------|------------------|
| MyWhoosh | ~7 days | Auto-recapture on next visit |
| iGPSport | ~7 days | Auto-recapture on next visit |
| Zwift | ~6 hours | Prompt user to visit zwift.com |
| Intervals.icu | API key | Never expires |

### Cost Estimate

At 100 users, 500 syncs/week:

| Component | Usage | Monthly Cost |
|-----------|-------|--------------|
| Functions execution | ~2000 executions | $0 (free tier) |
| Functions compute | ~1000 GB-seconds | $0.20 |
| Table Storage | 1 GB | $0.05 |
| Blob Storage (FIT files) | 5 GB | $1.00 |
| Egress | 10 GB | $0.87 |
| Key Vault (token encryption) | 1000 operations | $0.03 |
| **Total** | | **~$2-5/month** |

Even at 10x scale (1000 users), stays under $50/month.

## Extension-Backend Protocol

### Token Capture Flow

```typescript
// Extension detects API call
chrome.webRequest.onBeforeSendHeaders.addListener(
  (details) => {
    if (isSourcePlatform(details.url)) {
      const authHeader = details.requestHeaders.find(h => h.name === 'Authorization');
      if (authHeader?.value?.startsWith('Bearer ')) {
        sendToBackend({
          platform: detectPlatform(details.url),
          token: authHeader.value.substring(7),
          capturedAt: new Date().toISOString(),
          expiresAt: extractExpiry(token) // from JWT if available
        });
      }
    }
  },
  { urls: SOURCE_PLATFORM_PATTERNS }
);
```

### Backend Token Endpoint

```fsharp
[<Function("CaptureToken")>]
let captureToken 
    ([<HttpTrigger(AuthorizationLevel.Function, "post", Route = "tokens")>] req: HttpRequest)
    ([<Table("Tokens")>] tokenTable: TableClient) =
    task {
        let! body = req.ReadFromJsonAsync<TokenCapture>()
        let userId = req.Headers["X-User-Id"] // from extension auth
        
        // Encrypt token before storage
        let encryptedToken = KeyVault.encrypt body.Token
        
        let entity = {
            PartitionKey = userId
            RowKey = body.Platform
            Token = encryptedToken
            ExpiresAt = body.ExpiresAt
            CapturedAt = DateTimeOffset.UtcNow
        }
        
        do! tokenTable.UpsertEntityAsync(entity)
        return OkResult()
    }
```

## Implementation Phases

### Phase 1: MVP (Week 1-2)
- [ ] Azure Functions project with HTTP triggers
- [ ] Token capture and storage
- [ ] Single-user sync (reuse CLI sync logic)
- [ ] Basic extension that captures and sends tokens

### Phase 2: Multi-Tenant (Week 3-4)
- [ ] User authentication (Azure AD B2C or simple API keys)
- [ ] Durable Entities for per-user state
- [ ] Timer-triggered sync orchestration
- [ ] Activity register integration (ADR-018)

### Phase 3: User Experience (Week 5-6)
- [ ] Extension popup showing connection status
- [ ] Token expiry notifications
- [ ] Simple web dashboard (sync history, status)

### Phase 4: Growth Features
- [ ] Strava as additional sink
- [ ] Garmin Connect as additional sink
- [ ] Webhook notifications
- [ ] Activity analytics

## Consequences

### Positive
- **90% cost reduction** vs Orleans architecture
- **Zero cold-start penalty** for Durable Functions (reuses existing instances)
- **Browser extension is the moat** - hard to replicate without building same thing
- **F# throughout** - shared domain model between CLI, Functions, and analysis

### Negative
- **No real-time sync** - timer-based (hourly), not event-driven
- **Token expiry is a UX challenge** - especially for Zwift's 6-hour tokens
- **Azure lock-in** - Durable Functions is Azure-specific

### Mitigations
- For near-real-time: Extension could trigger immediate sync on token capture
- For token expiry: Browser notifications prompting platform visit
- For lock-in: Core sync logic is in portable F# library, only orchestration is Azure-specific

## Related Decisions

- [ADR-016: Browser Extension Token Capture](016-browser-extension-token-capture.md) - Extension design
- [ADR-017: Orleans Cloud Architecture](017-orleans-cloud-architecture.md) - **SUPERSEDED**
- [ADR-018: Local Activity Register](018-local-activity-register.md) - Provenance tracking (cloud version)
