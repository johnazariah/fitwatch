package fitparser

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func getTestdataPath() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..", "testdata")
}

func TestParse_RealFitFile(t *testing.T) {
	samplePath := filepath.Join(getTestdataPath(), "sample.fit")

	if _, err := os.Stat(samplePath); os.IsNotExist(err) {
		t.Skip("sample.fit not found in testdata")
	}

	meta, err := Parse(samplePath)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Basic file info
	if meta.Hash == "" {
		t.Error("expected non-empty hash")
	}
	if len(meta.Hash) != 64 { // SHA256 hex = 64 chars
		t.Errorf("hash should be 64 chars, got %d", len(meta.Hash))
	}
	if meta.Size == 0 {
		t.Error("expected non-zero size")
	}

	t.Logf("Parsed FIT file:")
	t.Logf("  Hash: %s", meta.Hash[:16]+"...")
	t.Logf("  Size: %d bytes", meta.Size)
	t.Logf("  Activity Type: %s", meta.ActivityType)
	t.Logf("  Start Time: %v", meta.StartTime)
	t.Logf("  Duration: %d seconds", meta.DurationSecs)
	t.Logf("  Distance: %.2f meters", meta.DistanceMeters)
	t.Logf("  Avg Power: %d watts", meta.AvgPower)
	t.Logf("  Max Power: %d watts", meta.MaxPower)
	t.Logf("  Avg HR: %d bpm", meta.AvgHeartRate)
	t.Logf("  Max HR: %d bpm", meta.MaxHeartRate)
	t.Logf("  Calories: %d", meta.Calories)
	t.Logf("  Manufacturer: %s", meta.Manufacturer)
}

func TestParse_ActivityType(t *testing.T) {
	samplePath := filepath.Join(getTestdataPath(), "sample.fit")

	if _, err := os.Stat(samplePath); os.IsNotExist(err) {
		t.Skip("sample.fit not found in testdata")
	}

	meta, err := Parse(samplePath)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should have some activity type
	if meta.ActivityType == "" {
		t.Log("Warning: no activity type found")
	}
}

func TestParse_HasTimestamps(t *testing.T) {
	samplePath := filepath.Join(getTestdataPath(), "sample.fit")

	if _, err := os.Stat(samplePath); os.IsNotExist(err) {
		t.Skip("sample.fit not found in testdata")
	}

	meta, err := Parse(samplePath)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if meta.StartTime == nil {
		t.Error("expected start time to be set")
	}

	if meta.DurationSecs == 0 {
		t.Log("Warning: duration is 0")
	}
}

func TestParse_HasPowerData(t *testing.T) {
	samplePath := filepath.Join(getTestdataPath(), "sample.fit")

	if _, err := os.Stat(samplePath); os.IsNotExist(err) {
		t.Skip("sample.fit not found in testdata")
	}

	meta, err := Parse(samplePath)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// MyWhoosh FIT files should have power data
	if meta.AvgPower == 0 {
		t.Log("Warning: no average power data")
	}
	if meta.MaxPower == 0 {
		t.Log("Warning: no max power data")
	}

	// Max should be >= avg
	if meta.MaxPower > 0 && meta.AvgPower > 0 {
		if meta.MaxPower < meta.AvgPower {
			t.Errorf("max power (%d) should be >= avg power (%d)", meta.MaxPower, meta.AvgPower)
		}
	}
}

func TestParse_HasHeartRateData(t *testing.T) {
	samplePath := filepath.Join(getTestdataPath(), "sample.fit")

	if _, err := os.Stat(samplePath); os.IsNotExist(err) {
		t.Skip("sample.fit not found in testdata")
	}

	meta, err := Parse(samplePath)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should have heart rate if user had HR monitor
	if meta.AvgHeartRate > 0 {
		if meta.AvgHeartRate < 30 || meta.AvgHeartRate > 250 {
			t.Errorf("avg heart rate %d seems unrealistic", meta.AvgHeartRate)
		}
	}

	if meta.MaxHeartRate > 0 && meta.AvgHeartRate > 0 {
		if meta.MaxHeartRate < meta.AvgHeartRate {
			t.Errorf("max HR (%d) should be >= avg HR (%d)", meta.MaxHeartRate, meta.AvgHeartRate)
		}
	}
}

func TestParse_DistanceReasonable(t *testing.T) {
	samplePath := filepath.Join(getTestdataPath(), "sample.fit")

	if _, err := os.Stat(samplePath); os.IsNotExist(err) {
		t.Skip("sample.fit not found in testdata")
	}

	meta, err := Parse(samplePath)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if meta.DistanceMeters < 0 {
		t.Error("distance should not be negative")
	}

	// Most rides are less than 300km
	if meta.DistanceMeters > 300000 {
		t.Logf("Warning: distance %.2f km seems very long", meta.DistanceMeters/1000)
	}
}

func TestParse_NonExistentFile(t *testing.T) {
	_, err := Parse("/nonexistent/path/to/file.fit")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestParse_InvalidFile(t *testing.T) {
	// Create a temp file with invalid content
	tmpFile, err := os.CreateTemp("", "invalid-*.fit")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	_, _ = tmpFile.WriteString("this is not a FIT file")
	tmpFile.Close()

	_, err = Parse(tmpFile.Name())
	if err == nil {
		t.Error("expected error for invalid FIT file")
	}
}

func TestParse_EmptyFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "empty-*.fit")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	_ = tmpFile.Close()

	_, err = Parse(tmpFile.Name())
	if err == nil {
		t.Error("expected error for empty file")
	}
}

func TestHashFile(t *testing.T) {
	samplePath := filepath.Join(getTestdataPath(), "sample.fit")

	if _, err := os.Stat(samplePath); os.IsNotExist(err) {
		t.Skip("sample.fit not found in testdata")
	}

	hash1, err := HashFile(samplePath)
	if err != nil {
		t.Fatalf("HashFile failed: %v", err)
	}

	if len(hash1) != 64 {
		t.Errorf("expected 64 char hash, got %d", len(hash1))
	}

	// Hash should be consistent
	hash2, err := HashFile(samplePath)
	if err != nil {
		t.Fatalf("HashFile failed: %v", err)
	}

	if hash1 != hash2 {
		t.Error("hash should be consistent for same file")
	}
}

func TestHashFile_NonExistent(t *testing.T) {
	_, err := HashFile("/nonexistent/file.fit")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestParse_HashMatchesHashFile(t *testing.T) {
	samplePath := filepath.Join(getTestdataPath(), "sample.fit")

	if _, err := os.Stat(samplePath); os.IsNotExist(err) {
		t.Skip("sample.fit not found in testdata")
	}

	meta, err := Parse(samplePath)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	hash, err := HashFile(samplePath)
	if err != nil {
		t.Fatalf("HashFile failed: %v", err)
	}

	if meta.Hash != hash {
		t.Errorf("Parse hash (%s) != HashFile hash (%s)", meta.Hash, hash)
	}
}

func BenchmarkParse(b *testing.B) {
	samplePath := filepath.Join(getTestdataPath(), "sample.fit")

	if _, err := os.Stat(samplePath); os.IsNotExist(err) {
		b.Skip("sample.fit not found in testdata")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Parse(samplePath)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHashFile(b *testing.B) {
	samplePath := filepath.Join(getTestdataPath(), "sample.fit")

	if _, err := os.Stat(samplePath); os.IsNotExist(err) {
		b.Skip("sample.fit not found in testdata")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := HashFile(samplePath)
		if err != nil {
			b.Fatal(err)
		}
	}
}
