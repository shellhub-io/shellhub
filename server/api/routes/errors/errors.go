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
	// ErrCodeForbidden is the error code for when the caller is known and still may not do what it asked.
	ErrCodeForbidden
	// ErrCodeNotFound is the error code for when the route cannot resolve what the request named.
	ErrCodeNotFound
	// ErrCodeTooManyRequests is the error code for when the caller must wait before asking again.
	ErrCodeTooManyRequests
	// ErrCodePaymentRequired is the error code for when the instance's licence does not cover what
	// the caller asked for.
	ErrCodePaymentRequired
)

// ErrDataInvalidEntity carries which fields failed validation and why, so the UI can mark
// them individually rather than showing one message for the whole form.
type ErrDataInvalidEntity struct {
	Fields map[string]string `json:"fields"`
}

// The route layer's own failures, raised before a request reaches a service.
var (
	ErrUnprocessableEntity = errors.New("unprocessable entity", ErrLayer, ErrCodeUnprocessableEntity)
	ErrInvalidEntity       = errors.New("invalid entity", ErrLayer, ErrCodeInvalidEntity)
	ErrUnauthorized        = errors.New("unauthorized", ErrLayer, ErrCodeUnauthorized)
	ErrForbidden           = errors.New("forbidden", ErrLayer, ErrCodeForbidden)
	ErrNotFound            = errors.New("not found", ErrLayer, ErrCodeNotFound)
	ErrTooManyRequests     = errors.New("too many requests", ErrLayer, ErrCodeTooManyRequests)
	ErrPaymentRequired     = errors.New("payment required", ErrLayer, ErrCodePaymentRequired)
)

// NewErrUnprocessableEntity returns an error when input model has syntax errors.
func NewErrUnprocessableEntity(err error) error {
	return errors.Wrap(ErrUnprocessableEntity, err)
}

// NewErrInvalidEntity returns an error with the invalids fields and why it is invalid after a validation.
func NewErrInvalidEntity(fields map[string]string) error {
	return errors.Wrap(errors.WithData(ErrInvalidEntity, ErrDataInvalidEntity{Fields: fields}), nil)
}

// NewErrUnauthorized returns an error with the access is not authorized.
func NewErrUnauthorized(err error) error {
	return errors.Wrap(ErrUnauthorized, err)
}

// NewErrForbidden returns an error for a caller the route identified and still refuses. Its message
// is fixed, so a refusal discloses neither which permission nor which namespace the caller lacks.
func NewErrForbidden(err error) error {
	return errors.Wrap(ErrForbidden, err)
}

// NewErrNotFound returns an error for a resource the route cannot resolve before reaching a
// service. A service that owns the lookup raises its own not-found error instead.
func NewErrNotFound(err error) error {
	return errors.Wrap(ErrNotFound, err)
}

// NewErrTooManyRequests returns an error for a caller the route is rate limiting. Any detail about
// how long to wait belongs in a response header, because this renders a fixed message.
func NewErrTooManyRequests(err error) error {
	return errors.Wrap(ErrTooManyRequests, err)
}

// NewErrPaymentRequired returns an error for a request the instance's licence does not cover. The
// licence middlewares in the cloud repository raise it before any service runs, which is why it
// belongs to this layer and not to the service one.
func NewErrPaymentRequired(err error) error {
	return errors.Wrap(ErrPaymentRequired, err)
}
