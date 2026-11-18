package errors

import (
	"errors"
)

// Data is a custom type to carry error's metadata.
// It can be any data type, but a struct is recommended due its fine control on JSON marshalling for each field.
type Data any

// Error is a custom error that carry attributes to specify error's message, resource, layer, code and data.
type Error struct {
	Message string
	Layer   string
	Code    int
	Data    Data
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

// WithData returns the [Error] in parent's chain with data as its Data. A nil parent returns nil,
// and a parent with no [Error] in its chain comes back unchanged.
func WithData(parent error, data Data) error {
	if err, ok := errors.AsType[Error](parent); ok {
		err.Data = data

		return err
	}

	return parent
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

// As wraps [errors.As]. Check [errors.As] for more information.
func As(err error, target any) bool {
	return errors.As(err, target)
}

// Is wraps [errors.Is]. Check [errors.Is] for more information.
func Is(err, target error) bool {
	return errors.Is(err, target)
}
