package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/api/pkg/guard"
	"github.com/shellhub-io/shellhub/api/routes/handlers/converter"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/pkg/errors"
)

// Errors is a centrlaized echo error handling for all API resonses. When a route returns a generic error, this handler
// evaluates the error and returns a specific HTTP status code and error message.
func Errors(err error, c echo.Context) {
	if err == guard.ErrForbidden {
		_ = c.NoContent(http.StatusForbidden)

		return
	}

	e, ok := err.(errors.Error)
	if !ok {
		// If err is not from the errors.Error type, return its error with an internal error status.
		// Generally, this happen when the service layer does not deal with the error.
		_ = c.NoContent(http.StatusInternalServerError)

		return
	}

	switch e.Layer {
	case services.ErrLayer:
		// When the error layer is from the service layer, return a specific HTTP status code and error message.
		_ = c.NoContent(converter.FromErrServiceToHTTPStatus(e.Code))

		return
	default:
		// No mapped error.
		_ = c.NoContent(http.StatusInternalServerError)

		return
	}
}
