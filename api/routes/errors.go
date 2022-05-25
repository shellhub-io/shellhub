package routes

import (
	"github.com/shellhub-io/shellhub/pkg/errors"
)

// ErrLayer is an error level. Each error defined at this level, is container to it.
// ErrLayer is the errors' level for service's error.
const ErrLayer = "route"

const (
	// ErrCodeInvalidEntity is the error code for when he input model is invalid.
	ErrCodeInvalidEntity = iota + 1
)

type ErrDataInvalidEntity struct {
	Fields map[string]string
}

var (
	ErrInvalidEntity = errors.New("Invalid entity", ErrLayer, ErrCodeInvalidEntity)
)

// NewErrInvalidEntity returns an error with the invalids fields and why it is invalid after a validation.
func NewErrInvalidEntity(fields map[string]string) error {
	return errors.Wrap(errors.WithData(ErrInvalidEntity, ErrDataInvalidEntity{Fields: fields}), nil)
}
