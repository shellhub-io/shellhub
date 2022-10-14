package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/api/pkg/echo/handlers/pkg/converter"
	"github.com/shellhub-io/shellhub/api/pkg/guard"
	routes "github.com/shellhub-io/shellhub/api/routes/errors"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/pkg/errors"
)

// NewErrors returns a custom echo's error handler.
//
// When error comes from guard.ErrLayer, the HTTP status code is http.StatusForbidden. When error comes from either
// routes.ErrLayer or services.ErrLayer, it is converted using the converter's package to one HTTP status code, and When
// error is not from the errors.Error type, a generic http.StatusInternalServerError.
func NewErrors() func(error, echo.Context) {
	return func(e error, ctx echo.Context) {
		if err, ok := e.(errors.Error); ok {
			switch err.Layer {
			case guard.ErrLayer:
				ctx.NoContent(http.StatusForbidden) // nolint:errcheck
			case routes.ErrLayer:
				ctx.NoContent(converter.FromErrRouteToHTTPStatus(err.Code)) // nolint:errcheck
			case services.ErrLayer:
				ctx.NoContent(converter.FromErrServiceToHTTPStatus(err.Code)) // nolint:errcheck
			default:
				ctx.NoContent(http.StatusInternalServerError) // nolint:errcheck
			}
		} else {
			ctx.NoContent(http.StatusInternalServerError) // nolint:errcheck
		}
	}
}
