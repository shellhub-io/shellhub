package worker

import (
	"context"
)

// Client represents a client that submits tasks to be handled by the server.
type Client interface {
	// Submit sends a payload to be processed by the task handler registered with the specified pattern.
	// The task will be executed immediately if it matches the pattern.
	//
	// It returns an error if the pattern is invalid or if there is an issue submitting the task.
	Submit(ctx context.Context, pattern TaskPattern, payload []byte) error
	// SubmitToBatch sends a payload to be added to a batch for processing. The task handler registered with
	// the specified pattern will process the batch either when a series of payloads have been enqueued
	// or when the specified time delay is reached.
	//
	// It returns an error if the pattern is invalid or if there is an issue submitting the task to the batch.
	SubmitToBatch(ctx context.Context, pattern TaskPattern, payload []byte) error
	// Close closes the client's connection.
	Close() error
}
