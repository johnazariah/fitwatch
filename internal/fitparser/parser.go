// Package fitparser provides FIT file parsing functionality.
package fitparser

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/tormoder/fit"
)

// Metadata contains parsed FIT file information.
type Metadata struct {
	// File info
	Hash string
	Size int64

	// Activity data
	ActivityType   string
	ActivityName   string
	StartTime      *time.Time
	EndTime        *time.Time
	DurationSecs   int
	ElapsedSecs    int
	DistanceMeters float64
	Calories       int

	// Performance metrics
	AvgPower     int
	MaxPower     int
	NormPower    int
	AvgHeartRate int
	MaxHeartRate int
	AvgCadence   int
	MaxCadence   int
	AvgSpeedMPS  float64
	MaxSpeedMPS  float64

	// Elevation
	TotalAscent  float64
	TotalDescent float64

	// Device info
	Manufacturer    string
	Product         string
	SerialNumber    uint32
	SoftwareVersion string
}

// Parse reads a FIT file and extracts metadata.
func Parse(path string) (*Metadata, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	// Get file size
	stat, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat file: %w", err)
	}

	// Calculate hash while reading
	hash := sha256.New()
	reader := io.TeeReader(f, hash)

	// Parse FIT file
	fitFile, err := fit.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("decode FIT: %w", err)
	}

	meta := &Metadata{
		Hash: hex.EncodeToString(hash.Sum(nil)),
		Size: stat.Size(),
	}

	// Extract activity data
	if activity, err := fitFile.Activity(); err == nil {
		extractActivity(meta, activity)
	}

	// Extract device info
	extractDeviceInfo(meta, fitFile)

	return meta, nil
}

// ParseReader reads FIT data from a reader.
func ParseReader(r io.Reader, size int64) (*Metadata, error) {
	hash := sha256.New()
	reader := io.TeeReader(r, hash)

	fitFile, err := fit.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("decode FIT: %w", err)
	}

	meta := &Metadata{
		Hash: hex.EncodeToString(hash.Sum(nil)),
		Size: size,
	}

	if activity, err := fitFile.Activity(); err == nil {
		extractActivity(meta, activity)
	}

	extractDeviceInfo(meta, fitFile)

	return meta, nil
}

// HashFile calculates SHA256 hash of a file.
func HashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func extractActivity(meta *Metadata, activity *fit.ActivityFile) {
	// Session data (summary)
	for _, session := range activity.Sessions {
		// Activity type
		meta.ActivityType = session.Sport.String()
		if session.SubSport != fit.SubSportGeneric {
			meta.ActivityType = session.SubSport.String()
		}

		// Timestamps
		if !session.StartTime.IsZero() {
			t := session.StartTime
			meta.StartTime = &t
		}
		if !session.Timestamp.IsZero() {
			t := session.Timestamp
			meta.EndTime = &t
		}

		// Duration (stored as milliseconds in uint32)
		if session.TotalTimerTime != 0 {
			meta.DurationSecs = int(session.TotalTimerTime / 1000)
		}
		if session.TotalElapsedTime != 0 {
			meta.ElapsedSecs = int(session.TotalElapsedTime / 1000)
		}

		// Distance
		if session.TotalDistance != 0 {
			meta.DistanceMeters = float64(session.TotalDistance) / 100 // cm to m
		}

		// Calories
		meta.Calories = int(session.TotalCalories)

		// Power
		if session.AvgPower != 0xFFFF {
			meta.AvgPower = int(session.AvgPower)
		}
		if session.MaxPower != 0xFFFF {
			meta.MaxPower = int(session.MaxPower)
		}
		if session.NormalizedPower != 0xFFFF {
			meta.NormPower = int(session.NormalizedPower)
		}

		// Heart rate
		if session.AvgHeartRate != 0xFF {
			meta.AvgHeartRate = int(session.AvgHeartRate)
		}
		if session.MaxHeartRate != 0xFF {
			meta.MaxHeartRate = int(session.MaxHeartRate)
		}

		// Cadence
		if session.AvgCadence != 0xFF {
			meta.AvgCadence = int(session.AvgCadence)
		}
		if session.MaxCadence != 0xFF {
			meta.MaxCadence = int(session.MaxCadence)
		}

		// Speed
		if session.AvgSpeed != 0xFFFF {
			meta.AvgSpeedMPS = float64(session.AvgSpeed) / 1000 // mm/s to m/s
		}
		if session.MaxSpeed != 0xFFFF {
			meta.MaxSpeedMPS = float64(session.MaxSpeed) / 1000
		}

		// Elevation
		if session.TotalAscent != 0xFFFF {
			meta.TotalAscent = float64(session.TotalAscent)
		}
		if session.TotalDescent != 0xFFFF {
			meta.TotalDescent = float64(session.TotalDescent)
		}

		// Only use first session
		break
	}

	// Activity name from file (if available)
	if activity.Activity != nil {
		// The activity message doesn't have a name field directly
		// Name would come from a separate event or be set by the platform
	}
}

func extractDeviceInfo(meta *Metadata, fitFile *fit.File) {
	// File ID message contains device info
	if fitFile.FileId.Manufacturer != fit.ManufacturerInvalid {
		meta.Manufacturer = fitFile.FileId.Manufacturer.String()
	}
	if fitFile.FileId.Product != 0 {
		meta.Product = fmt.Sprintf("%d", fitFile.FileId.Product)
	}
	if fitFile.FileId.SerialNumber != 0 {
		meta.SerialNumber = fitFile.FileId.SerialNumber
	}
}
