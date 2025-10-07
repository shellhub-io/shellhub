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

// HasError checks if there was an error in the HTTP response or if the provided error is not nil.
func HasError(resp *resty.Response, err error) error {
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
