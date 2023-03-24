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
	// ErrCodeUnauthorized is the error code for when the user is not authorized to access the resource.
	ErrCodeUnauthorized
)

type ErrDataInvalidEntity struct {
	Fields []string `json:"fields"`
}

var (
	ErrUnprocessableEntity = errors.New("unprocessable entity", ErrLayer, ErrCodeUnprocessableEntity)
	ErrInvalidEntity       = errors.New("invalid entity", ErrLayer, ErrCodeInvalidEntity)
	ErrUnauthorized        = errors.New("unauthorized", ErrLayer, ErrCodeUnauthorized)
)

// NewErrUnprocessableEntity returns an error when input model has syntax errors.
func NewErrUnprocessableEntity(err error) error {
	return errors.Wrap(ErrUnprocessableEntity, err)
}

// NewErrInvalidEntity returns an error with the invalids fields and why it is invalid after a validation.
func NewErrInvalidEntity(fields []string) error {
	return errors.Wrap(errors.WithData(ErrInvalidEntity, ErrDataInvalidEntity{Fields: fields}), nil)
}

func NewErrUnauthorized(err error) error {
	return errors.Wrap(ErrUnauthorized, err)
}
