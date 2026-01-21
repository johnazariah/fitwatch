package intervals

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConsumer_Name(t *testing.T) {
	c := New("athlete123", "apikey456")
	if c.Name() != "Intervals.icu" {
		t.Errorf("expected Intervals.icu, got %s", c.Name())
	}
}

func TestConsumer_Validate_Success(t *testing.T) {
	c := New("athlete123", "apikey456")
	if err := c.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestConsumer_Validate_MissingAthleteID(t *testing.T) {
	c := New("", "apikey456")
	err := c.Validate()
	if err == nil {
		t.Error("expected error for missing athlete ID")
	}
	if !strings.Contains(err.Error(), "athlete ID") {
		t.Errorf("error should mention athlete ID: %v", err)
	}
}

func TestConsumer_Validate_MissingAPIKey(t *testing.T) {
	c := New("athlete123", "")
	err := c.Validate()
	if err == nil {
		t.Error("expected error for missing API key")
	}
	if !strings.Contains(err.Error(), "API key") {
		t.Errorf("error should mention API key: %v", err)
	}
}

func TestConsumer_Push_Success(t *testing.T) {
	// Create a mock server
	var receivedPath string
	var receivedAuth string
	var receivedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		_, pass, _ := r.BasicAuth()
		receivedAuth = pass

		body, _ := io.ReadAll(r.Body)
		receivedBody = body

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": "activity123"}`))
	}))
	defer server.Close()

	// Create a temp FIT file
	tmpDir := t.TempDir()
	fitPath := filepath.Join(tmpDir, "test.fit")
	if err := os.WriteFile(fitPath, []byte("fake FIT data"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create consumer with mock server
	c := New("athlete123", "apikey456")
	c.BaseURL = server.URL

	// Push the file
	ctx := context.Background()
	if err := c.Push(ctx, fitPath); err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	// Verify request
	expectedPath := "/api/v1/athlete/athlete123/activities"
	if receivedPath != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, receivedPath)
	}

	if receivedAuth != "apikey456" {
		t.Errorf("expected API key apikey456, got %s", receivedAuth)
	}

	if len(receivedBody) == 0 {
		t.Error("expected non-empty body")
	}

	// Body should contain filename and FIT data
	bodyStr := string(receivedBody)
	if !strings.Contains(bodyStr, "test.fit") {
		t.Error("body should contain filename")
	}
	if !strings.Contains(bodyStr, "fake FIT data") {
		t.Error("body should contain file content")
	}
}

func TestConsumer_Push_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	fitPath := filepath.Join(tmpDir, "test.fit")
	if err := os.WriteFile(fitPath, []byte("fake FIT data"), 0644); err != nil {
		t.Fatal(err)
	}

	c := New("athlete123", "apikey456")
	c.BaseURL = server.URL

	ctx := context.Background()
	err := c.Push(ctx, fitPath)
	if err == nil {
		t.Error("expected error for server error")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error should contain status code: %v", err)
	}
}

func TestConsumer_Push_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "unauthorized"}`))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	fitPath := filepath.Join(tmpDir, "test.fit")
	if err := os.WriteFile(fitPath, []byte("fake FIT data"), 0644); err != nil {
		t.Fatal(err)
	}

	c := New("athlete123", "wrongkey")
	c.BaseURL = server.URL

	ctx := context.Background()
	err := c.Push(ctx, fitPath)
	if err == nil {
		t.Error("expected error for unauthorized")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("error should contain 401: %v", err)
	}
}

func TestConsumer_Push_FileNotFound(t *testing.T) {
	c := New("athlete123", "apikey456")

	ctx := context.Background()
	err := c.Push(ctx, "/nonexistent/path/to/file.fit")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestConsumer_Push_InvalidConfig(t *testing.T) {
	c := New("", "apikey456") // Missing athlete ID

	tmpDir := t.TempDir()
	fitPath := filepath.Join(tmpDir, "test.fit")
	if err := os.WriteFile(fitPath, []byte("fake FIT data"), 0644); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	err := c.Push(ctx, fitPath)
	if err == nil {
		t.Error("expected error for invalid config")
	}
	if !strings.Contains(err.Error(), "athlete ID") {
		t.Errorf("error should mention athlete ID: %v", err)
	}
}

func TestConsumer_Push_ContextCancellation(t *testing.T) {
	// Create a server that delays
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't respond immediately
		select {}
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	fitPath := filepath.Join(tmpDir, "test.fit")
	if err := os.WriteFile(fitPath, []byte("fake FIT data"), 0644); err != nil {
		t.Fatal(err)
	}

	c := New("athlete123", "apikey456")
	c.BaseURL = server.URL

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := c.Push(ctx, fitPath)
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

func TestConsumer_Push_RealFitFile(t *testing.T) {
	// Skip if no real FIT file available
	fitPath := filepath.Join("..", "..", "..", "testdata", "sample.fit")
	if _, err := os.Stat(fitPath); os.IsNotExist(err) {
		t.Skip("sample.fit not found in testdata")
	}

	var receivedSize int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedSize = len(body)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := New("athlete123", "apikey456")
	c.BaseURL = server.URL

	ctx := context.Background()
	if err := c.Push(ctx, fitPath); err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	// Multipart body should be larger than the original file due to headers
	fi, _ := os.Stat(fitPath)
	if receivedSize < int(fi.Size()) {
		t.Errorf("received body (%d) should be >= file size (%d)", receivedSize, fi.Size())
	}
	t.Logf("Uploaded %d bytes (file was %d bytes)", receivedSize, fi.Size())
}

func TestConsumer_Push_RateLimited(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error": "rate limited"}`))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	fitPath := filepath.Join(tmpDir, "test.fit")
	if err := os.WriteFile(fitPath, []byte("fake FIT data"), 0644); err != nil {
		t.Fatal(err)
	}

	c := New("athlete123", "apikey456")
	c.BaseURL = server.URL

	ctx := context.Background()
	err := c.Push(ctx, fitPath)
	if err == nil {
		t.Error("expected error for rate limit")
	}
	if !strings.Contains(err.Error(), "429") {
		t.Errorf("error should contain 429: %v", err)
	}
}

func TestConsumer_Push_DuplicateActivity(t *testing.T) {
	// Intervals.icu returns 409 for duplicate activities
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"error": "activity already exists"}`))
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	fitPath := filepath.Join(tmpDir, "test.fit")
	if err := os.WriteFile(fitPath, []byte("fake FIT data"), 0644); err != nil {
		t.Fatal(err)
	}

	c := New("athlete123", "apikey456")
	c.BaseURL = server.URL

	ctx := context.Background()
	err := c.Push(ctx, fitPath)
	if err == nil {
		t.Error("expected error for duplicate")
	}
	if !strings.Contains(err.Error(), "409") {
		t.Errorf("error should contain 409: %v", err)
	}
}

func BenchmarkConsumer_Push(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	tmpDir := b.TempDir()
	fitPath := filepath.Join(tmpDir, "test.fit")
	// Create a larger file for more realistic benchmark
	data := make([]byte, 100*1024) // 100KB
	if err := os.WriteFile(fitPath, data, 0644); err != nil {
		b.Fatal(err)
	}

	c := New("athlete123", "apikey456")
	c.BaseURL = server.URL

	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := c.Push(ctx, fitPath); err != nil {
			b.Fatal(err)
		}
	}
}
