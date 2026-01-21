// Package store provides persistence for FIT file tracking.
package store

import (
	"time"
)

// FitFile represents a discovered FIT file with parsed metadata.
type FitFile struct {
	ID           int64     `json:"id"`
	Path         string    `json:"path"`
	Hash         string    `json:"hash"`
	Size         int64     `json:"size"`
	DiscoveredAt time.Time `json:"discoveredAt"`
	Source       string    `json:"source"` // "watch", "scan", "api"

	// Parsed FIT metadata
	ActivityType    string     `json:"activityType,omitempty"`
	ActivityName    string     `json:"activityName,omitempty"`
	StartedAt       *time.Time `json:"startedAt,omitempty"`
	DurationSecs    int        `json:"durationSecs,omitempty"`
	DistanceM       float64    `json:"distanceM,omitempty"`
	Calories        int        `json:"calories,omitempty"`
	AvgPowerW       int        `json:"avgPowerW,omitempty"`
	MaxPowerW       int        `json:"maxPowerW,omitempty"`
	NormPowerW      int        `json:"normPowerW,omitempty"`
	AvgHR           int        `json:"avgHr,omitempty"`
	MaxHR           int        `json:"maxHr,omitempty"`
	AvgCadence      int        `json:"avgCadence,omitempty"`
	AvgSpeedMPS     float64    `json:"avgSpeedMps,omitempty"`
	TotalAscentM    float64    `json:"totalAscentM,omitempty"`
	DeviceName      string     `json:"deviceName,omitempty"`
	SoftwareVersion string     `json:"softwareVersion,omitempty"`
}

// SyncStatus represents the state of a sync attempt.
type SyncStatus string

const (
	SyncStatusPending SyncStatus = "pending"
	SyncStatusSuccess SyncStatus = "success"
	SyncStatusFailed  SyncStatus = "failed"
)

// SyncRecord represents an attempt to sync a file to a consumer.
type SyncRecord struct {
	ID          int64      `json:"id"`
	FileID      int64      `json:"fileId"`
	Consumer    string     `json:"consumer"`
	Status      SyncStatus `json:"status"`
	AttemptedAt *time.Time `json:"attemptedAt,omitempty"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
	RemoteID    string     `json:"remoteId,omitempty"`
	RemoteURL   string     `json:"remoteUrl,omitempty"`
	Error       string     `json:"error,omitempty"`
	Retries     int        `json:"retries"`
}

// ConsumerConfig stores consumer settings.
type ConsumerConfig struct {
	Name       string     `json:"name"`
	Enabled    bool       `json:"enabled"`
	ConfigJSON string     `json:"configJson,omitempty"`
	LastSync   *time.Time `json:"lastSync,omitempty"`
}

// StoreStats provides aggregate statistics.
type StoreStats struct {
	TotalFiles        int            `json:"totalFiles"`
	TotalSyncs        int            `json:"totalSyncs"`
	PendingByConsumer map[string]int `json:"pendingByConsumer"`
	FailedByConsumer  map[string]int `json:"failedByConsumer"`
	SuccessByConsumer map[string]int `json:"successByConsumer"`
}
