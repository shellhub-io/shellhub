// Package certbot provides interfaces for SSL certificate management operations.
// These interfaces allow for easy mocking in tests and clean abstractions over
// the exec and time packages.
package certbot

import (
	"context"
	"os/exec"
	"time"
)

// Executor provides an interface for executing system commands.
// This interface allows for easy mocking in tests and provides
// a clean abstraction over the exec package.
type Executor interface {
	// Command creates a new *exec.Cmd with the given name and arguments.
	Command(name string, arg ...string) *exec.Cmd
	// Run executes the given command and waits for it to complete.
	Run(cmd *exec.Cmd) error
}

// Ticker provides an interface for time-based operations with context support.
// This interface allows for easy mocking in tests and provides a clean
// abstraction over the time package's ticker functionality.
type Ticker interface {
	// Init creates a new time.Ticker internally with the specified duration.
	// The ticker will respect the provided context for cancellation.
	Init(context.Context, time.Duration)
	// Tick returns a channel that receives the current time on each tick.
	// If the ticker wasn't initialized, the channel will be nil.
	Tick() chan time.Time
	// Stop stops the ticker. If the ticker wasn't initialized, this is a no-op.
	Stop()
}
