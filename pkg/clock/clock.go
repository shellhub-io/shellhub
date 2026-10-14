package clock

import (
	"sync"
	"time"
)

// Clock is where code reads the current time. A component receives one when it is built: the wall
// clock from [New] in production, a [Fixed] clock in a test.
type Clock interface {
	Now() time.Time
}

// New returns the wall clock. Production builds every service with it; a test passes a [Fixed]
// clock instead.
func New() Clock {
	return &realClock{}
}

type realClock struct{}

func (c *realClock) Now() time.Time {
	return time.Now() //nolint:forbidigo // the wall clock New hands to every component built for production
}

// Fixed is a test clock that returns the instant it holds until Set or Advance moves it. It is safe
// for concurrent use, so a service under test can read it while the test moves it.
type Fixed struct {
	mu  sync.Mutex
	now time.Time
}

// NewFixed returns a [Fixed] clock holding at.
func NewFixed(at time.Time) *Fixed {
	return &Fixed{now: at}
}

// Now returns the instant the clock holds.
func (f *Fixed) Now() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.now
}

// Set makes the clock hold at.
func (f *Fixed) Set(at time.Time) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.now = at
}

// Advance moves the instant the clock holds by d; a negative d moves it back.
func (f *Fixed) Advance(d time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.now = f.now.Add(d)
}
