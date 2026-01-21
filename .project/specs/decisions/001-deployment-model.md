# ADR-001: Self-Hosted First, Cloud-Optional

## Status
Superseded by ADR-017

## Superseded By
- **ADR-016**: Browser Extension Token Capture Architecture
- **ADR-017**: Orleans Cloud Architecture

## Note
After spiking TrainingPeaks and MyWhoosh integrations, we discovered that:
1. These platforms require browser-based authentication (no public OAuth)
2. Token capture is most reliable via browser extension
3. A cloud-hosted Orleans backend with browser extension provides the best UX

The original "self-hosted first" approach is deprioritized in favor of a cloud-native architecture with the "FitBridge" browser extension for token capture.

---

## Original Context (Preserved for History)
We need to decide whether this application should be:
- A hosted SaaS service
- A self-hosted application users run themselves
- A hybrid approach

## Decision Drivers
- Privacy concerns (FIT files contain GPS location data)
- Cost of hosting for a side project
- Complexity of multi-tenant infrastructure
- User technical sophistication (likely: cyclists who are often tech-savvy)

## Options Considered

### Option A: Self-Hosted Only
Users run the application on their own machine (Docker, local Python).

**Pros:**
- No hosting costs for us
- Maximum privacy - data never leaves user's machine
- Simpler architecture (no multi-tenancy)
- Can use local LLMs for analysis

**Cons:**
- Requires technical users
- No mobile access without additional setup
- Users must maintain/update the software

### Option B: Hosted SaaS
We run infrastructure, users just sign up.

**Pros:**
- Easiest for users
- Mobile/web access from anywhere
- We control updates and reliability

**Cons:**
- Hosting costs (database, storage, compute)
- Privacy concerns with GPS data
- Multi-tenant complexity
- OAuth callback URLs need stable domain

### Option C: Hybrid (Self-Hosted Core + Optional Cloud Sync)
Core application runs locally, optional cloud features for convenience.

**Pros:**
- Best of both worlds
- Privacy-conscious users stay local
- Convenience users get cloud features
- Gradual path to monetization if desired

**Cons:**
- More complex to build both paths
- Documentation overhead

## Decision
**Option C: Hybrid, with self-hosted as the primary path**

Start by building a robust self-hosted application. Design the architecture so cloud sync could be added later, but don't build it initially.

## Rationale
1. Privacy is a genuine concern for fitness data
2. No hosting costs during development/early users
3. Simpler to start, can add cloud later
4. Target users (serious cyclists) are often technical enough for Docker

## Implementation Notes
- Package as Docker Compose for easy deployment
- Also support direct Python install for developers
- Design config/storage to be pluggable (local vs S3)
- OAuth callbacks will need user-provided domain (ngrok, tailscale, etc.)

## Consequences
- Need good documentation for self-hosting
- OAuth testing is more complex without stable URLs
- Mobile access requires user to expose service (tailscale, cloudflare tunnel)

## Review Date
Revisit after MVP launch based on user feedback.
