package handlers

import (
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/api/pkg/echo/handlers/pkg/converter"
	"github.com/shellhub-io/shellhub/api/pkg/guard"
	routes "github.com/shellhub-io/shellhub/api/routes/errors"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/errors"
)

func report(reporter *sentry.Client, err error, request *http.Request) {
	go func() {
		if reporter != nil {
			reporter.CaptureEvent(&sentry.Event{ //nolint:exhaustruct
				Level:   sentry.LevelError,
				Message: err.Error(),
				Request: sentry.NewRequest(request),
			}, &sentry.EventHint{Request: request}, &sentry.Scope{}) //nolint:exhaustruct
		}
	}()
}

// NewErrors returns a custom echo's error handler.
//
// When the error is from errors.Error type, it will check for the layer and response with the appropriated HTTP status
// code. However, if the error is not from errors.Error type, it will respond with HTTP status code 500. When this error
// occurs, it will also try to send the error to Sentry.
func NewErrors(reporter *sentry.Client) func(error, echo.Context) {
	return func(e error, ctx echo.Context) {
		if err, ok := e.(errors.Error); ok {
			switch err.Layer {
			case guard.ErrLayer:
				ctx.NoContent(http.StatusForbidden) //nolint:errcheck
			case routes.ErrLayer:
				ctx.NoContent(converter.FromErrRouteToHTTPStatus(err.Code)) //nolint:errcheck
			case services.ErrLayer:
				if last := errors.GetLastError(err); last != nil {
					if converted, ok := last.(errors.Error); !ok || (converted.Layer != services.ErrLayer && converted.Layer != store.ErrLayer) {
						report(reporter, last, ctx.Request())
						ctx.NoContent(http.StatusInternalServerError) //nolint:errcheck

						return
					}
				}

				ctx.NoContent(converter.FromErrServiceToHTTPStatus(err.Code)) //nolint:errcheck
			case store.ErrLayer:
				report(reporter, err, ctx.Request())
				ctx.NoContent(http.StatusInternalServerError) //nolint:errcheck
			default:
				report(reporter, err, ctx.Request())
				ctx.NoContent(http.StatusInternalServerError) //nolint:errcheck
			}
		} else {
			report(reporter, e, ctx.Request())

			ctx.NoContent(http.StatusInternalServerError) //nolint:errcheck
		}
	}
}
