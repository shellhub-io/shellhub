package errors

import (
	"errors"
	"fmt"
)

// Error is the structure that represents an error in the SSH connection. It is a composition of two errors, an internal
// error, that indicates what caused the error, and the external error, that is what the end use will see, more
// user-friendly and simple.
type Error struct {
	// Internal error is the error used to debug the error.
	Internal error
	// External error is the error that is returned to the user.
	External error
}

// New creates a new error with the given internal and external errors.
func New(internal, external error) error {
	return Error{
		Internal: internal,
		External: external,
	}
}

// Error returns the error message from internal error only.
func (e Error) Error() string {
	return e.Internal.Error()
}

// GetInternal returns a internal error if error is from Error type. if error is nil, it returns nil, and if error is
// not from Error type, it returns error.
func GetInternal(err error) error {
	if err == nil {
		return nil
	}

	if e, ok := err.(Error); ok {
		return e.Internal
	}

	return err
}

// GetExternal returns a external error if error is from Error type. if error is nil, it returns nil, and if error is
// not from Error type, it returns error.
func GetExternal(err error) error {
	if err == nil {
		return nil
	}

	if e, ok := err.(Error); ok {
		return e.External
	}

	return err
}

// Is checks if err is from Error type. If is from Error type, it returns true if internal or external error is equal to
// target error. If is not from Error type, normally check if errors are equal. If both errors are nil, it returns true.
func Is(err, target error) bool {
	if err == nil {
		return target == nil
	}

	e, ok := err.(Error)
	if !ok {
		return errors.Is(err, target)
	}

	return errors.Is(GetInternal(e), target) || errors.Is(GetExternal(e), target)
}

// Wrap wraps two errors into a new one.
func Wrap(err, next error) error {
	if next == nil {
		return err
	}

	return fmt.Errorf("%s: %w", err, next)
}
