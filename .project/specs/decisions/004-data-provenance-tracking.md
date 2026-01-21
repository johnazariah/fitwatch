# ADR-004: Data Provenance Tracking

## Status
Accepted

## Context
Since we're positioning as a data freedom/portability platform, we need to track where data comes from. This serves multiple purposes:
- Legal clarity on data ownership
- Deduplication across sources
- Audit trail for user
- Informing export eligibility

## Decision

**Every activity tracks its complete provenance chain.**

### Data Model

```csharp
public class DataProvenance
{
    // How we got this file
    public AcquisitionMethod Method { get; set; }
    public DateTime AcquiredAt { get; set; }
    
    // Original source
    public DeviceInfo? RecordingDevice { get; set; }
    public PlatformInfo? SourcePlatform { get; set; }
    
    // File integrity
    public string OriginalFileName { get; set; }
    public string FileHashSha256 { get; set; }
    public long FileSizeBytes { get; set; }
    
    // Ownership assertion
    public OwnershipAssertion Ownership { get; set; }
}

public enum AcquisitionMethod
{
    // Highest confidence - direct from device
    DirectDeviceSync,       // Watched folder from device sync app
    ManualFileUpload,       // User uploaded FIT file directly
    
    // High confidence - user-initiated export
    PlatformBulkExport,     // User exported from platform, imported zip
    CloudStorageSync,       // From user's Dropbox/OneDrive/GDrive
    
    // Medium confidence - API with user consent
    FriendlyPlatformApi,    // Platform explicitly allows (Garmin, Wahoo)
    
    // Low confidence - potential TOS issues
    RestrictedPlatformApi,  // Platform restricts but user consents
    
    // Unknown
    Unknown                 // Legacy import, source unclear
}

public class DeviceInfo
{
    public string? Manufacturer { get; set; }   // "Garmin", "Wahoo"
    public string? Product { get; set; }        // "Edge 840", "ELEMNT ROAM"
    public string? SerialNumber { get; set; }
    public string? SoftwareVersion { get; set; }
}

public class PlatformInfo
{
    public string PlatformId { get; set; }      // "garmin", "mywhoosh"
    public string? OriginalActivityId { get; set; }
    public DateTime? OriginalUploadDate { get; set; }
}

public class OwnershipAssertion
{
    public OwnershipConfidence Confidence { get; set; }
    public string? UserAssertion { get; set; }  // "I recorded this myself"
    public DateTime AssertedAt { get; set; }
}

public enum OwnershipConfidence
{
    Definite,       // Direct from user's device
    High,           // User-initiated export from platform
    Medium,         // API download with user auth
    Low,            // Unclear origin
    Contested       // Potential TOS conflict
}
```

## Consequences

### For Deduplication
- Same `FileHashSha256` = same activity, regardless of source
- First import wins for provenance; later imports add "also seen at"

### For Export Eligibility
```csharp
public bool CanExportTo(Activity activity, Sink sink)
{
    // Direct device files can go anywhere
    if (activity.Provenance.Method == AcquisitionMethod.DirectDeviceSync)
        return true;
    
    // Check sink-specific restrictions
    return !HasRestriction(activity.Provenance, sink);
}
```

### For User Transparency
Dashboard shows:
```
Activity: Morning Ride
Recorded: 2026-01-20 07:30
Device: Garmin Edge 840 (serial: XXX)
Imported: Via folder sync from C:\Garmin\Activities
Ownership: ✅ Definite - direct from your device
Also uploaded to: Garmin Connect, Intervals.icu
```

## Related Decisions
- ADR-003: FIT Files as Canonical Format
- ADR-005: Platform Integration Policies
