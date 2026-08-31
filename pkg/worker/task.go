package worker

import (
	"context"
	"strings"
)

// TaskPattern represents a pattern to which a task can register to listen.
// It must be in the format "queue:kind".
type TaskPattern string

func (tp TaskPattern) String() string {
	return string(tp)
}

// Validate reports whether the pattern is valid or not.
func (tp TaskPattern) Validate() bool {
	return len(strings.Split(string(tp), ":")) == 2
}

// MustValidate is similar to [TaskPattern.Validate] but panics when invalid.
func (tp TaskPattern) MustValidate() {
	if !tp.Validate() {
		panic("invalid task pattern: " + tp)
	}
}

// Queue returns the queue component of the pattern.
func (tp TaskPattern) Queue() string {
	return strings.Split(string(tp), ":")[0]
}

// TaskHandler processes one message. Returning an error hands the message back to the queue for
// retry, so a handler must be idempotent.
type TaskHandler func(ctx context.Context, payload []byte) error

// Task binds a handler to the pattern it listens on. Build one with NewTask rather than by hand,
// so the pattern is validated before the worker starts.
type Task struct {
	// Pattern is a string to which the task can listen to message/events.
	Pattern TaskPattern
	// Handler is the callback that the task will execute when receiving messages/events.
	Handler TaskHandler
}

// TaskOption configures a task at registration.
type TaskOption func(t *Task)
