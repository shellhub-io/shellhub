package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/api/pkg/guard"
	"github.com/shellhub-io/shellhub/api/routes/handlers/converter"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/pkg/errors"
)

// Errors maps errors code to http status code.
func Errors(err error, c echo.Context) {
	if err == guard.ErrForbidden {
		_ = c.NoContent(http.StatusForbidden)

		return
	}

	e, ok := err.(errors.Error)
	if !ok {
		// If err is not errors.Error type, return its error with an internal error status.
		// Generally, this is happening in an inferior layer and was not checked by the service.
		_ = c.NoContent(http.StatusInternalServerError)

		return
	}

	switch e.Layer {
	case services.ErrLayer:
		_ = c.NoContent(converter.FromErrServiceToHTTPStatus(e.Code))

		return
	default:
		// No mapped error.
		_ = c.NoContent(http.StatusInternalServerError)

		return
	}
}
