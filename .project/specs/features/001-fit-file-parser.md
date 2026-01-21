# Feature: FIT File Parser

## Priority
P0 - Required for all other features

## Summary
Parse .FIT files and extract structured workout data that can be stored, analyzed, and displayed.

## Background
FIT (Flexible and Interoperable Data Transfer) is the standard file format for fitness devices. It's a binary format developed by Garmin that contains time-series data from workouts.

## User Stories
- As a user, I want my uploaded FIT files to be parsed so I can see workout details
- As a developer, I want parsed data in a consistent format regardless of the source device
- As an analyst, I want access to all available metrics for LLM-based analysis

## Acceptance Criteria

### Core Parsing
- [ ] Parse FIT files from major devices (Garmin, Wahoo, Zwift, MyWhoosh)
- [ ] Extract session-level summary data (duration, distance, avg power, etc.)
- [ ] Extract lap/interval data
- [ ] Extract time-series records (power, HR, cadence, speed, GPS)
- [ ] Handle files with missing/optional fields gracefully

### Data Quality
- [ ] Validate FIT file integrity before parsing
- [ ] Handle corrupt or truncated files with clear error messages
- [ ] Normalize units (always meters, watts, bpm, etc.)
- [ ] Handle timezone/timestamp normalization

### Performance
- [ ] Parse a typical 1-hour ride file in < 2 seconds
- [ ] Memory usage < 100MB for large files

## Technical Design

### Library Choice
Use `fitdecode` library:
```python
pip install fitdecode
```

Rationale:
- Actively maintained
- Handles edge cases well
- Supports all FIT message types
- Good performance

### Data Structures

```python
from dataclasses import dataclass
from datetime import datetime
from typing import Optional

@dataclass
class ActivitySummary:
    """Session-level summary from FIT file"""
    sport: str                          # cycling, running, etc.
    sub_sport: str                      # virtual_cycling, indoor_cycling
    start_time: datetime
    end_time: datetime
    total_elapsed_time: float           # seconds
    total_timer_time: float             # seconds (excluding pauses)
    total_distance: float               # meters
    total_calories: int
    avg_speed: Optional[float]          # m/s
    max_speed: Optional[float]          # m/s
    avg_heart_rate: Optional[int]       # bpm
    max_heart_rate: Optional[int]       # bpm
    avg_power: Optional[int]            # watts
    max_power: Optional[int]            # watts
    normalized_power: Optional[int]     # watts
    avg_cadence: Optional[int]          # rpm
    max_cadence: Optional[int]          # rpm
    total_ascent: Optional[int]         # meters
    total_descent: Optional[int]        # meters
    training_stress_score: Optional[float]
    intensity_factor: Optional[float]

@dataclass
class Lap:
    """Lap/interval data"""
    start_time: datetime
    total_elapsed_time: float
    total_distance: float
    avg_speed: Optional[float]
    avg_heart_rate: Optional[int]
    avg_power: Optional[int]
    avg_cadence: Optional[int]
    max_power: Optional[int]
    max_heart_rate: Optional[int]

@dataclass 
class Record:
    """Single time-series data point"""
    timestamp: datetime
    latitude: Optional[float]           # degrees
    longitude: Optional[float]          # degrees
    altitude: Optional[float]           # meters
    heart_rate: Optional[int]           # bpm
    cadence: Optional[int]              # rpm
    speed: Optional[float]              # m/s
    power: Optional[int]                # watts
    temperature: Optional[int]          # celsius
    distance: Optional[float]           # meters (cumulative)

@dataclass
class ParsedActivity:
    """Complete parsed FIT file"""
    file_hash: str                      # SHA256 of file for dedup
    source_file: str                    # Original filename
    summary: ActivitySummary
    laps: list[Lap]
    records: list[Record]
    raw_messages: dict                  # Unparsed messages for debugging
```

### Parser Interface

```python
class FITParser:
    def parse_file(self, file_path: Path) -> ParsedActivity:
        """Parse a FIT file from disk"""
        
    def parse_bytes(self, data: bytes) -> ParsedActivity:
        """Parse FIT file from bytes (for uploads)"""
        
    def validate(self, file_path: Path) -> tuple[bool, str]:
        """Check if file is valid FIT, return (valid, error_message)"""
```

### Error Handling

```python
class FITParseError(Exception):
    """Base exception for parsing errors"""

class InvalidFITFile(FITParseError):
    """File is not a valid FIT file"""

class CorruptFITFile(FITParseError):
    """File appears to be FIT but is corrupt/truncated"""

class UnsupportedFITVersion(FITParseError):
    """FIT protocol version not supported"""
```

## Test Cases

### Unit Tests
- [ ] Parse sample Garmin cycling FIT file
- [ ] Parse sample Wahoo cycling FIT file
- [ ] Parse sample Zwift cycling FIT file
- [ ] Parse sample MyWhoosh cycling FIT file
- [ ] Parse running FIT file
- [ ] Parse file with GPS data
- [ ] Parse indoor file (no GPS)
- [ ] Parse file with power meter data
- [ ] Parse file without power meter data
- [ ] Handle missing optional fields
- [ ] Reject non-FIT file with clear error
- [ ] Handle truncated FIT file
- [ ] Handle FIT file with CRC errors

### Sample Files Needed
Create `.project/test-data/` with sample FIT files:
- [ ] garmin-outdoor-ride.fit
- [ ] garmin-indoor-ride.fit
- [ ] wahoo-ride.fit
- [ ] zwift-ride.fit
- [ ] mywhoosh-ride.fit
- [ ] corrupt-file.fit
- [ ] not-a-fit-file.txt

## Dependencies
None - this is a foundational feature.

## Estimates
- Research & sample file collection: 2 hours
- Implementation: 4 hours
- Testing: 2 hours
- Documentation: 1 hour

## Notes
- Consider storing raw FIT file alongside parsed data for reprocessing
- May need to handle device-specific quirks (each platform encodes slightly differently)
- GPS coordinates should be considered sensitive data

## References
- [FIT SDK Documentation](https://developer.garmin.com/fit/protocol/)
- [fitdecode library](https://github.com/polyvertex/fitdecode)
- [FIT file message types](https://developer.garmin.com/fit/cookbook/)
