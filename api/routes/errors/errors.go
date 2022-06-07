package routes

import (
	"github.com/shellhub-io/shellhub/pkg/errors"
)

// ErrLayer is an error level. Each error defined at this level, is container to it.
// ErrLayer is the errors' level for service's error.
const ErrLayer = "route"

const (
	// ErrCodeUnprocessableEntity is the error code for when the input model has syntax errors.
	ErrCodeUnprocessableEntity = iota + 1
	// ErrCodeInvalidEntity is the error code for when he input model is invalid.
	ErrCodeInvalidEntity
)

type ErrDataInvalidEntity struct {
	Fields map[string]string
}

var (
	ErrUnprocessableEntity = errors.New("Unprocessable entity", ErrLayer, ErrCodeUnprocessableEntity)
	ErrInvalidEntity       = errors.New("Invalid entity", ErrLayer, ErrCodeInvalidEntity)
)

// NewErrUnprocessableEntity returns an error when input model has syntax errors.
func NewErrUnprocessableEntity(err error) error {
	return errors.Wrap(ErrUnprocessableEntity, err)
}

// NewErrInvalidEntity returns an error with the invalids fields and why it is invalid after a validation.
func NewErrInvalidEntity(fields map[string]string) error {
	return errors.Wrap(errors.WithData(ErrInvalidEntity, ErrDataInvalidEntity{Fields: fields}), nil)
}
