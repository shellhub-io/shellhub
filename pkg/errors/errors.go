package errors

import (
	"errors"
)

// Data is a custom type to carry error's metadata.
// It can be any data type, but a struct is recommended due its fine control on JSON marshalling for each field.
type Data any

// Error is a custom error that carry attributes to specify error's message, resource, layer, code and data.
type Error struct {
	// message is the error message.
	Message string `json:"message"`
	// Layer is the error layer.
	Layer string `json:"layer,omitempty"`
	// Code is the error code.
	Code int `json:"code,omitempty"`
	// Data is the error metadata.
	Data Data `json:"data,omitempty"`
}

// New creates a new [Error].
func New(message, layer string, code int) error {
	return Error{
		Message: message,
		Layer:   layer,
		Code:    code,
		Data:    nil,
	}
}

// WithData insiert [Data] into parent is from type [Error].
func WithData(parent error, data Data) error {
	if parent == nil {
		return nil
	}

	var err Error
	if errors.As(parent, &err) {
		err.Data = data

		return err
	}

	return nil
}

func (e Error) Error() string {
	return e.Message
}

// Is ignores Data so that WithData copies still match their sentinel.
func (e Error) Is(target error) bool {
	t, ok := target.(Error)
	if !ok {
		return false
	}

	return e.Message == t.Message && e.Layer == t.Layer && e.Code == t.Code
}

// Wrap wraps an error with another error.
//
// It is a interface for [errors.Join]. Check [errors.Join] for more information.
func Wrap(err error, next error) error {
	return errors.Join(err, next)
}

// Unwrap returns the next error from the error tree.
func Unwrap(err error) error {
	return errors.Unwrap(err)
}

// As wraps [errors.As]. Check [errors.As] for more information.
func As(err error, target any) bool {
	return errors.As(err, target)
}

// Is wraps [errors.Is]. Check [errors.Is] for more information.
func Is(err, target error) bool {
	return errors.Is(err, target)
}
