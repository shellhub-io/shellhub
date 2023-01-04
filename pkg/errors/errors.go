package errors

import (
	"fmt"
	"strings"
)

type Data interface{}

// Error is a custom error that carry attributes to specify error's message, resource, layer, code and data.
// Error implements error and unwrap interfaces.
type Error struct {
	// message is the error message.
	Message string `json:"message"`
	// Layer is the error layer.
	Layer string `json:"layer,omitempty"`
	// Code is the error code.
	Code int `json:"code,omitempty"`
	// Data is the error metadata.
	Data Data `json:"data,omitempty"`
	// Next is the next error in the error's chain. next is nil when has no more error in the error's chain.
	Next error `json:"next,omitempty"`
}

// New creates a new Error.
//
// An Error contains a message, message that will be showed when Error() method is called, a layer, where the error
// happened and a code, that should be unique in the layer.
func New(message, layer string, code int) error {
	return Error{
		Message: message,
		Layer:   layer,
		Code:    code,
		Data:    nil,
		Next:    nil,
	}
}

// WithData creates a new Error from other with data. If parent is not from Error type, just return the parameter.
func WithData(parent error, data Data) error {
	if parent == nil {
		return nil
	}

	if err, ok := parent.(Error); ok {
		return Error{
			Message: err.Message,
			Layer:   err.Layer,
			Code:    err.Code,
			Data:    data,
			Next:    err.Next,
		}
	}

	return parent
}

// Error returns a message aggregating all errors' messages in the chain.
func (e Error) Error() string {
	message := e.Message

	if e.Next != nil {
		// Recursively, get and join all messages in the chain.
		message = strings.Join([]string{message, e.Next.Error()}, ": ")
	}

	return message
}

// Unwrap returns the next error in the error's chain. If there is no next error, returns nil.
func (e Error) Unwrap() error {
	return e.Next
}

// Wrap adds an Error to the error's chain. If err is nil, return nil. If next is nil, return err.
func Wrap(err error, next error) error {
	if err == nil {
		return nil
	}

	if next == nil {
		return err
	}

	e, ok := err.(Error)
	if !ok {
		return fmt.Errorf("%s: %w", err.Error(), next)
	}

	err = nil //nolint:wastedassign
	n, ok := next.(Error)
	if !ok {
		err = Error{Message: next.Error()}
	} else {
		err = n
	}

	return Error{
		Message: e.Message,
		Layer:   e.Layer,
		Code:    e.Code,
		Data:    e.Data,
		Next:    err,
	}
}

// GetLastError returns the last error in the error's chain. If there is no next error, returns nil.
func GetLastError(err error) error {
	for {
		if err == nil {
			break
		}

		next, ok := err.(Error)
		if !ok {
			break
		}

		if next.Next == nil {
			break
		}

		err = next.Next
	}

	return err
}
