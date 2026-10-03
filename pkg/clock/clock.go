package clock

import (
	"time"
)

// Clock is an interface that can provide time related functionality which allows us to test time dependent code.
type Clock interface {
	Now() time.Time
}

// DefaultBackend is used to configure the defaultBackend.
var DefaultBackend Clock

func init() {
	DefaultBackend = &realClock{}
}

// Now returns the current time from the package's backend. Use it rather than time.Now everywhere
// but a deadline or an elapsed-time measurement, so a test can control what "now" means.
func Now() time.Time {
	return DefaultBackend.Now()
}

// Freeze makes Now return ts until t and its subtests finish, then puts back the backend it
// replaced. Calls nest: an inner Freeze gives back the outer one when its test ends. The backend is
// process-global, so a test that calls Freeze must not run in parallel with one that reads the clock.
func Freeze(t interface{ Cleanup(func()) }, ts time.Time) {
	Set(t, frozenClock(ts))
}

// Set serves Now from c until t and its subtests finish, then puts back the backend it replaced.
// Use it when a test needs Now to change; Freeze covers a single instant. Like Freeze, it swaps a
// process-global backend and is not safe under t.Parallel.
func Set(t interface{ Cleanup(func()) }, c Clock) {
	previous := DefaultBackend
	t.Cleanup(func() { DefaultBackend = previous })
	DefaultBackend = c
}

type frozenClock time.Time

func (f frozenClock) Now() time.Time {
	return time.Time(f)
}

type realClock struct{}

// This function is responsible for getting the current time.
func (c *realClock) Now() time.Time {
	return time.Now() //nolint:forbidigo // this is the wall clock every other caller mocks
}
