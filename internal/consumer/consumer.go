// Package consumer defines the interface for FIT file consumers.
// A consumer receives FIT files and pushes them to a destination.
package consumer

import (
	"context"
	"fmt"
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
	consumers []Consumer
}

// NewDispatcher creates a dispatcher with the given consumers.
func NewDispatcher(consumers ...Consumer) *Dispatcher {
	return &Dispatcher{consumers: consumers}
}

// AddConsumer adds a consumer to the dispatcher.
func (d *Dispatcher) AddConsumer(c Consumer) {
	d.consumers = append(d.consumers, c)
}

// Dispatch sends a FIT file to all registered consumers.
// Returns results for each consumer (success or failure).
func (d *Dispatcher) Dispatch(ctx context.Context, fitPath string) []Result {
	results := make([]Result, 0, len(d.consumers))

	for _, c := range d.consumers {
		err := c.Push(ctx, fitPath)
		results = append(results, Result{
			Consumer: c.Name(),
			FitPath:  fitPath,
			Success:  err == nil,
			Error:    err,
		})
	}

	return results
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
