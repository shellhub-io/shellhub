package client

import (
	"errors"
	"fmt"
	"net/http"
)

// Response is the part of an HTTP response the error mapping needs: everything here is decided by
// the status code, so the body is not part of the contract.
type Response interface {
	StatusCode() int
}

var (
	// ErrUnknown is returned when a non-mapped error occurred.
	ErrUnknown = errors.New("unknown error")
	// ErrConnectionFailed is returned when the client could not communicate with the sever.
	ErrConnectionFailed = errors.New("connection failed")
	// ErrNotFound is returned when the resource wasn't found or the route does't exist.
	ErrNotFound = errors.New("not found")
	// ErrBadRequest is returned when the request is malformed or invalid.
	ErrBadRequest = errors.New("bad request")
	// ErrUnauthorized is returned when the client is not authenticated to perform the operation.
	ErrUnauthorized = errors.New("unauthorized")
	// ErrForbidden is returned when the client is authenticated but not allowed to perform the operation.
	ErrForbidden = errors.New("forbidden")
	// ErrMethodNotAllowed is returned when the HTTP method used is not allowed for the resource.
	ErrMethodNotAllowed = errors.New("method not allowed")
	// ErrConflict is returned when there is a conflict with the current state of the resource.
	ErrConflict = errors.New("conflict")
	// ErrPreconditionFailed is returned when a precondition set by the client fails.
	ErrPreconditionFailed = errors.New("precondition failed")
	// ErrTooManyRequests is returned when the client has exceeded its rate limit.
	ErrTooManyRequests = errors.New("too many requests")
	// ErrInternalServerError is returned when the server has cannot response to the request due an error.
	ErrInternalServerError = errors.New("internal server error")
	// ErrEmptyResponse is returned when a successful response carries no body for the client to
	// decode, so there is no value to hand the caller.
	ErrEmptyResponse = errors.New("empty response")
)

// ErrorFromResponse returns an error based on the response status code.
// Each Error is mapped to a specific status code, if the status code is not mapped ErrUnknown is returned.
func ErrorFromResponse(response Response) error {
	if (response.StatusCode() >= 200) && (response.StatusCode() <= 299) {
		return nil
	}

	switch response.StatusCode() {
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusInternalServerError:
		return ErrInternalServerError
	case http.StatusBadRequest:
		return ErrBadRequest
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusForbidden:
		return ErrForbidden
	case http.StatusMethodNotAllowed:
		return ErrMethodNotAllowed
	case http.StatusConflict:
		return ErrConflict
	case http.StatusPreconditionFailed:
		return ErrPreconditionFailed
	case http.StatusTooManyRequests:
		return ErrTooManyRequests
	default:
		return errors.Join(ErrUnknown, fmt.Errorf("%d", response.StatusCode()))
	}
}

func requireBody[T any](result *T) (*T, error) {
	if result == nil {
		return nil, ErrEmptyResponse
	}

	return result, nil
}
