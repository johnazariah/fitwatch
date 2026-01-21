package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Token represents a captured platform token
type Token struct {
	Token      string `json:"token"`
	CapturedAt string `json:"capturedAt"`
	Platform   string `json:"platform"`
}

// App struct
type App struct {
	ctx    context.Context
	tokens map[string]Token
	mu     sync.RWMutex
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		tokens: make(map[string]Token),
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// GetTokens returns all captured tokens
func (a *App) GetTokens() map[string]Token {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Return a copy
	result := make(map[string]Token)
	for k, v := range a.tokens {
		result[k] = v
	}
	return result
}

// ClearTokens removes all tokens
func (a *App) ClearTokens() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.tokens = make(map[string]Token)
}

// StartTokenServer starts the HTTP server for extension bridge
func (a *App) StartTokenServer() {
	http.HandleFunc("/api/tokens", func(w http.ResponseWriter, r *http.Request) {
		// CORS headers for extension
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var incoming map[string]Token
		if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		a.mu.Lock()
		for platform, token := range incoming {
			a.tokens[platform] = token
			fmt.Printf("[FitBridge] Received %s token\n", platform)
		}
		a.mu.Unlock()

		// Notify the frontend
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "tokens-updated", a.GetTokens())
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("FitBridge Desktop is running"))
	})

	fmt.Println("[FitBridge] Token server listening on :5847")
	server := &http.Server{
		Addr:         ":5847",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("[FitBridge] Server error: %v\n", err)
	}
}
