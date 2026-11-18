package routes

import (
	"fmt"

	"github.com/shellhub-io/shellhub/pkg/errors"
)

// ErrLayer is an error level. Each error defined at this level, is container to it.
// ErrLayer is the errors' level for service's error.
const ErrLayer = "route"

// The route layer's error codes, the kinds [errors.HTTPStatus] maps.
const (
	ErrCodeUnprocessableEntity = errors.CodeUnprocessable
	ErrCodeInvalidEntity       = errors.CodeInvalid
	ErrCodeUnauthorized        = errors.CodeUnauthorized
	ErrCodeForbidden           = errors.CodeForbidden
	ErrCodeNotFound            = errors.CodeNotFound
	ErrCodeTooManyRequests     = errors.CodeTooManyRequests
	ErrCodePaymentRequired     = errors.CodePayment
)

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
	return errors.Wrap(errors.WithData(ErrInvalidEntity, errors.InvalidFields{Fields: fields}), nil)
}

// NewErrUnreadableBody returns [ErrInvalidEntity] for a request body the server could not finish
// reading, with err kept in the chain for the request log. A status error in err, such as the body
// limit's 413, outranks the 400 in the error handler.
func NewErrUnreadableBody(err error) error {
	return errors.Wrap(ErrInvalidEntity, fmt.Errorf("read request body: %w", err))
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
