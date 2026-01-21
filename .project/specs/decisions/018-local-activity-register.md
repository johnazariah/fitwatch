# ADR-018: Local Activity Register

## Status
Accepted

## Context

During the Zwift integration spike, we discovered a gap: after syncing an activity to Intervals.icu, we have no memory of where it came from. When asked to fix the name of activity `i118287888`, we had to:

1. Query Intervals.icu for the activity's timestamp and duration
2. Search Zwift activities - no match
3. Search MyWhoosh activities - found the match
4. Manually update the name

This works but is inefficient and error-prone. The current sync flow:

```
Source Platform → FIT bytes → Upload to Intervals.icu → Done (no record kept)
```

We lose the **provenance** (where it came from) and the **mapping** (source ID ↔ sink ID) after sync completes.

### Problems This Causes

| Problem | Impact |
|---------|--------|
| No source tracking | Can't determine which platform an activity came from |
| Repeated API calls | Must query Intervals.icu every sync to detect duplicates |
| No audit trail | No record of when activities were synced |
| No batch operations | Can't "fix all MyWhoosh names" without re-querying each source |
| Multi-sink complexity | When we add Strava/Garmin as sinks, mapping becomes critical |

## Decision Drivers

- **Provenance tracking**: Know where each activity originated
- **Efficient duplicate detection**: Avoid repeated API calls
- **Batch operations**: Enable retroactive fixes across activities
- **Multi-sink future**: Support syncing to Strava, Garmin, TrainingPeaks
- **Offline capability**: View sync history without network
- **Simplicity**: Single JSON file, no database dependency

## Options Considered

### Option A: No Local State (Status Quo)

Query Intervals.icu every sync, match by fuzzy heuristics.

**Pros:**
- Simple, no state to manage
- Always fresh data

**Cons:**
- Slow (API call every sync)
- No provenance tracking
- No audit trail
- Fuzzy matching can fail

### Option B: Local JSON Register

Store a `~/.fitsync/activities.json` file mapping source → sink activities.

**Pros:**
- Fast duplicate detection (local lookup)
- Full provenance tracking
- Enables batch operations
- Simple implementation (JSON file)
- Works offline for history/status

**Cons:**
- State can drift from reality (deleted activities, manual uploads)
- Need periodic reconciliation with sinks

### Option C: SQLite Database

Use SQLite for structured queries and indexing.

**Pros:**
- Efficient queries on large datasets
- ACID transactions
- Indexes for fast lookup

**Cons:**
- Overkill for ~1000 activities
- Binary file, harder to inspect/debug
- Additional dependency

## Decision

**Option B: Local JSON Register**

A JSON file provides the right balance of simplicity and capability for a CLI tool. We can always migrate to SQLite later if scale demands it.

## Schema Design

```json
{
  "version": 1,
  "activities": [
    {
      "id": "uuid-v4",
      "source": "zwift",
      "sourceId": "1658639309344882736",
      "title": "Zwift - Climb Portal: La Turbie + Col d'Eze",
      "startTime": "2024-07-22T07:37:11Z",
      "duration": 6551,
      "distance": 16353.2,
      "activityType": "VirtualRide",
      "fitFileHash": "sha256:abc123...",
      "sinks": [
        {
          "sink": "intervals",
          "sinkId": "i119406365",
          "syncedAt": "2026-01-21T02:30:00Z",
          "status": "synced"
        }
      ],
      "metadata": {
        "avgWatts": 108,
        "elevation": 613
      }
    }
  ]
}
```

### Key Design Choices

1. **UUID as primary key**: Stable across source/sink ID changes
2. **Multiple sinks per activity**: Supports future multi-sink sync
3. **FIT file hash**: Enables content-based deduplication
4. **Sink status**: Track sync state (pending, synced, failed, deleted)
5. **Metadata blob**: Extensible for platform-specific data

## Operations Enabled

| Command | Description |
|---------|-------------|
| `fitsync status` | Show pending vs synced counts per source |
| `fitsync history` | Show recent sync activity |
| `fitsync fix-names --source mywhoosh` | Batch update names from source |
| `fitsync reconcile` | Compare local register with actual sinks |
| `fitsync show <sink-id>` | Show activity provenance |

## Implementation Plan

1. **Phase 1**: Record synced activities during sync (write-only)
2. **Phase 2**: Use register for duplicate detection (read during sync)
3. **Phase 3**: Add CLI commands for history/status
4. **Phase 4**: Add reconciliation and batch operations

## Consequences

### Positive
- Fast local duplicate detection
- Full provenance tracking
- Enables `fix-names` style batch operations
- Audit trail for debugging
- Foundation for multi-sink sync

### Negative
- State can drift from reality
- Need to handle corrupt/missing register file
- Migration path needed if schema changes

### Mitigations
- Periodic reconciliation with sinks
- Fallback to API-based duplicate detection if register missing
- Schema version field for future migrations

## Related Decisions

- [ADR-004: Data Provenance Tracking](004-data-provenance-tracking.md)
- [ADR-012: Duplicate Detection Strategy](012-duplicate-detection-strategy.md)
- [ADR-013: Source Activity Domain Model](013-source-activity-domain-model.md)
