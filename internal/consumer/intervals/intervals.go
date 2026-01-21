// Package intervals provides a consumer for Intervals.icu.
package intervals

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	defaultBaseURL = "https://intervals.icu"
)

// Consumer pushes FIT files to Intervals.icu.
type Consumer struct {
	AthleteID string
	APIKey    string
	BaseURL   string
	client    *http.Client
}

// New creates an Intervals.icu consumer.
func New(athleteID, apiKey string) *Consumer {
	return &Consumer{
		AthleteID: athleteID,
		APIKey:    apiKey,
		BaseURL:   defaultBaseURL,
		client:    &http.Client{},
	}
}

// Name returns the consumer name.
func (c *Consumer) Name() string {
	return "Intervals.icu"
}

// Validate checks configuration.
func (c *Consumer) Validate() error {
	if c.AthleteID == "" {
		return errors.New("athlete ID is required")
	}
	if c.APIKey == "" {
		return errors.New("API key is required")
	}
	return nil
}

// Push uploads a FIT file to Intervals.icu.
func (c *Consumer) Push(ctx context.Context, fitPath string) error {
	if err := c.Validate(); err != nil {
		return err
	}

	// Open the FIT file
	file, err := os.Open(fitPath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	filename := filepath.Base(fitPath)

	// Extract activity name from filename (e.g., "2025-02-23_Hudayriyat_Ascend.fit" -> "Hudayriyat Ascend")
	activityName := extractActivityName(filename)
	if activityName != "" {
		if err := writer.WriteField("name", activityName); err != nil {
			return fmt.Errorf("write name field: %w", err)
		}
	}

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return fmt.Errorf("create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("copy file: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("close writer: %w", err)
	}

	// Build request
	url := fmt.Sprintf("%s/api/v1/athlete/%s/activities", c.BaseURL, c.AthleteID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, &buf)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.SetBasicAuth("API_KEY", c.APIKey)

	// Send request
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Check response
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// extractActivityName extracts a human-readable name from a FIT filename.
// Examples:
//   - "2025-02-23_Hudayriyat_Ascend.fit" -> "Hudayriyat Ascend"
//   - "2025-02-23_.fit" -> "" (no name)
//   - "activity.fit" -> "" (no recognizable pattern)
func extractActivityName(filename string) string {
	// Remove extension
	name := strings.TrimSuffix(filename, filepath.Ext(filename))

	// Pattern: YYYY-MM-DD_Name or similar date prefix
	datePattern := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}[_-]?`)
	name = datePattern.ReplaceAllString(name, "")

	// If nothing left after removing date, no name
	if name == "" || name == "_" {
		return ""
	}

	// Replace underscores and hyphens with spaces
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ReplaceAll(name, "-", " ")

	// Clean up multiple spaces and trim
	name = strings.Join(strings.Fields(name), " ")

	return name
}
