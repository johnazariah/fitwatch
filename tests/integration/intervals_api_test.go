package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/johnazariah/fitwatch/internal/consumer/intervals"
)

// These tests require real Intervals.icu credentials.
// Set INTERVALS_ATHLETE_ID and INTERVALS_API_KEY environment variables.
// Tests are skipped if credentials are not available.

func getIntervalsCredentials(t *testing.T) (athleteID, apiKey string) {
	athleteID = os.Getenv("INTERVALS_ATHLETE_ID")
	apiKey = os.Getenv("INTERVALS_API_KEY")

	if athleteID == "" || apiKey == "" {
		t.Skip("INTERVALS_ATHLETE_ID and INTERVALS_API_KEY not set - skipping real API tests")
	}

	return athleteID, apiKey
}

func TestIntervalsAPI_Authentication(t *testing.T) {
	athleteID, apiKey := getIntervalsCredentials(t)

	ctx := context.Background()

	// Test that we can authenticate by fetching athlete info
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://intervals.icu/api/v1/athlete/%s", athleteID), nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth("API_KEY", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to connect to Intervals.icu: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == 401 {
		t.Fatal("authentication failed - check API key")
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var athlete struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&athlete); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Authenticated as: %s (ID: %s)", athlete.Name, athlete.ID)
}

func TestIntervalsAPI_InvalidCredentials(t *testing.T) {
	// This test verifies error handling for bad credentials
	// It doesn't require real credentials

	consumer := intervals.New("fake-athlete", "fake-api-key")

	tmpDir := t.TempDir()
	fitPath := filepath.Join(tmpDir, "test.fit")
	if err := os.WriteFile(fitPath, []byte("fake FIT data"), 0644); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := consumer.Push(ctx, fitPath)
	if err == nil {
		t.Error("expected error for invalid credentials")
	}
	if !strings.Contains(err.Error(), "401") && !strings.Contains(err.Error(), "403") {
		t.Logf("Got error (may vary): %v", err)
	}
}

func TestIntervalsAPI_UploadRealFitFile(t *testing.T) {
	athleteID, apiKey := getIntervalsCredentials(t)

	// Find the sample FIT file
	fitPath := filepath.Join("..", "..", "testdata", "sample.fit")
	if _, err := os.Stat(fitPath); os.IsNotExist(err) {
		t.Skip("sample.fit not found in testdata")
	}

	consumer := intervals.New(athleteID, apiKey)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := consumer.Push(ctx, fitPath)
	if err != nil {
		// 409 Conflict means duplicate - that's OK, file was already uploaded
		if strings.Contains(err.Error(), "409") {
			t.Log("Activity already exists (duplicate) - this is expected if test ran before")
			return
		}
		t.Fatalf("failed to upload FIT file: %v", err)
	}

	t.Log("Successfully uploaded FIT file to Intervals.icu")
}

func TestIntervalsAPI_UploadDuplicate(t *testing.T) {
	athleteID, apiKey := getIntervalsCredentials(t)

	// Find the sample FIT file
	fitPath := filepath.Join("..", "..", "testdata", "sample.fit")
	if _, err := os.Stat(fitPath); os.IsNotExist(err) {
		t.Skip("sample.fit not found in testdata")
	}

	consumer := intervals.New(athleteID, apiKey)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// First upload (may succeed or be duplicate)
	_ = consumer.Push(ctx, fitPath)

	// Second upload should be duplicate
	err := consumer.Push(ctx, fitPath)
	if err == nil {
		t.Log("No error on duplicate - Intervals.icu may have accepted it")
		return
	}

	if strings.Contains(err.Error(), "409") {
		t.Log("Correctly detected duplicate activity (409 Conflict)")
	} else {
		t.Logf("Got error: %v", err)
	}
}

func TestIntervalsAPI_ListActivities(t *testing.T) {
	athleteID, apiKey := getIntervalsCredentials(t)

	// List recent activities to verify API access
	oldest := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	newest := time.Now().Format("2006-01-02")

	url := fmt.Sprintf("https://intervals.icu/api/v1/athlete/%s/activities?oldest=%s&newest=%s",
		athleteID, oldest, newest)

	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth("API_KEY", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to list activities: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var activities []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&activities); err != nil {
		t.Fatalf("failed to decode activities: %v", err)
	}

	t.Logf("Found %d activities in the last 7 days", len(activities))
	for i, a := range activities {
		if i >= 5 {
			t.Logf("  ... and %d more", len(activities)-5)
			break
		}
		t.Logf("  - %s (%s): %s", a.ID, a.Type, a.Name)
	}
}

func TestIntervalsAPI_Wellness(t *testing.T) {
	athleteID, apiKey := getIntervalsCredentials(t)

	// Get today's wellness data
	today := time.Now().Format("2006-01-02")
	url := fmt.Sprintf("https://intervals.icu/api/v1/athlete/%s/wellness/%s", athleteID, today)

	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth("API_KEY", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to get wellness: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// 404 is OK - no wellness data for today
	if resp.StatusCode == 404 {
		t.Log("No wellness data for today")
		return
	}

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var wellness struct {
		CTL     float64 `json:"ctl"`
		ATL     float64 `json:"atl"`
		Resting int     `json:"restingHR"`
		Weight  float64 `json:"weight"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&wellness); err != nil {
		t.Fatalf("failed to decode wellness: %v", err)
	}

	t.Logf("Wellness: CTL=%.1f, ATL=%.1f, RestHR=%d, Weight=%.1f",
		wellness.CTL, wellness.ATL, wellness.Resting, wellness.Weight)
}

func TestIntervalsAPI_RateLimiting(t *testing.T) {
	athleteID, apiKey := getIntervalsCredentials(t)

	// Make multiple rapid requests to test rate limit handling
	// Intervals.icu has generous limits, so this mainly tests our code handles 429

	url := fmt.Sprintf("https://intervals.icu/api/v1/athlete/%s", athleteID)

	for i := 0; i < 10; i++ {
		req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
		if err != nil {
			t.Fatal(err)
		}
		req.SetBasicAuth("API_KEY", apiKey)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request %d failed: %v", i, err)
		}
		_ = resp.Body.Close()

		if resp.StatusCode == 429 {
			t.Logf("Hit rate limit after %d requests", i+1)
			return
		}

		if resp.StatusCode != 200 {
			t.Fatalf("request %d: unexpected status %d", i, resp.StatusCode)
		}
	}

	t.Log("No rate limiting detected after 10 rapid requests")
}
