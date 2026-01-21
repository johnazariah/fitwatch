// Package consumer defines the interface for FIT file consumers.
// A consumer receives FIT files and pushes them to a destination.
package consumer

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// Consumer processes FIT files and sends them to a destination.
type Consumer interface {
	// Name returns a human-readable name for the consumer.
	Name() string

	// Push sends a FIT file to the destination.
	// Returns nil on success, or an error if the push failed.
	Push(ctx context.Context, fitPath string) error

	// Validate checks if the consumer is properly configured.
	Validate() error
}

// Result represents the outcome of pushing a FIT file.
type Result struct {
	Consumer string
	FitPath  string
	Success  bool
	Error    error
}

// Dispatcher sends FIT files to multiple consumers.
type Dispatcher struct {
	consumers  []Consumer
	maxRetries int
	logger     *slog.Logger
}

// NewDispatcher creates a dispatcher with the given consumers.
func NewDispatcher(consumers ...Consumer) *Dispatcher {
	return &Dispatcher{
		consumers:  consumers,
		maxRetries: 3,
		logger:     slog.Default(),
	}
}

// SetMaxRetries configures the number of retry attempts for failed pushes.
func (d *Dispatcher) SetMaxRetries(n int) {
	d.maxRetries = n
}

// SetLogger configures the logger for the dispatcher.
func (d *Dispatcher) SetLogger(logger *slog.Logger) {
	d.logger = logger
}

// AddConsumer adds a consumer to the dispatcher.
func (d *Dispatcher) AddConsumer(c Consumer) {
	d.consumers = append(d.consumers, c)
}

// Dispatch sends a FIT file to all registered consumers.
// Returns results for each consumer (success or failure).
// Automatically retries failed pushes with exponential backoff.
func (d *Dispatcher) Dispatch(ctx context.Context, fitPath string) []Result {
	results := make([]Result, 0, len(d.consumers))

	for _, c := range d.consumers {
		err := d.pushWithRetry(ctx, c, fitPath)
		results = append(results, Result{
			Consumer: c.Name(),
			FitPath:  fitPath,
			Success:  err == nil,
			Error:    err,
		})
	}

	return results
}

// pushWithRetry attempts to push with exponential backoff.
func (d *Dispatcher) pushWithRetry(ctx context.Context, c Consumer, fitPath string) error {
	var lastErr error
	backoff := 1 * time.Second

	for attempt := 0; attempt <= d.maxRetries; attempt++ {
		if attempt > 0 {
			d.logger.Info("retrying upload", "consumer", c.Name(), "attempt", attempt, "backoff", backoff)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}

			// Exponential backoff: 1s, 2s, 4s, 8s...
			backoff = min(backoff*2, 30*time.Second)
		}

		lastErr = c.Push(ctx, fitPath)
		if lastErr == nil {
			if attempt > 0 {
				d.logger.Info("retry succeeded", "consumer", c.Name(), "attempts", attempt+1)
			}
			return nil
		}

		d.logger.Warn("push failed", "consumer", c.Name(), "attempt", attempt+1, "error", lastErr)
	}

	return fmt.Errorf("failed after %d attempts: %w", d.maxRetries+1, lastErr)
}

// ValidateAll checks all consumers are properly configured.
func (d *Dispatcher) ValidateAll() error {
	for _, c := range d.consumers {
		if err := c.Validate(); err != nil {
			return fmt.Errorf("%s: %w", c.Name(), err)
		}
	}
	return nil
}
