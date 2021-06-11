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

// The init function will set the defaultBackend to the realClock implementation.
func init() {
	DefaultBackend = &realClock{}
}

// Is responsible for calling method Now of the defaultBackend.
func Now() time.Time {
	return DefaultBackend.Now()
}

type realClock struct{}

// This function is responsible for getting the current time.
func (c *realClock) Now() time.Time {
	return time.Now()
}
