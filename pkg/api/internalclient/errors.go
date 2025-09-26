package internalclient

import (
	"errors"
	"fmt"

	resty "github.com/go-resty/resty/v2"
)

// ErrRequestFailed is returned when an HTTP request fails to be executed.
var ErrRequestFailed = errors.New("request failed")

// Error represents a custom error structure that includes an HTTP status code and a message.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("code=%d, message=%s", e.Code, e.Message)
}

// NewError creates a new error based on the response from the Resty HTTP client.
// If the response indicates an error (status code >= 400), it returns a custom Error
// containing the status code and message. If there was an error during the request,
// it joins that error with a predefined ErrRequestFailed. If there are no errors,
// it returns nil.
func NewError(resp *resty.Response, err error) error {
	if err != nil {
		return errors.Join(ErrRequestFailed, err)
	}

	if resp.IsError() {
		// NOTE: If we got a error, status code great than 399, we create our custom error to load it with the response
		// data.
		return &Error{Code: resp.StatusCode(), Message: resp.Status()}
	}

	return nil
}

// HasError checks if there was an error in the HTTP response or if the provided error is not nil.
// It returns true if there was an error, otherwise false.
func HasError(resp *resty.Response, err error) bool {
	if err != nil {
		return true
	}

	if resp.IsError() {
		return true
	}

	return false
}
