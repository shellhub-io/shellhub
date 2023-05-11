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

func isErrorUnknown(err error) (bool, error) {
	unknown, ok := err.(errors.Error)
	if !ok {
		return true, err
	}

	return (unknown.Layer != services.ErrLayer && unknown.Layer != store.ErrLayer), unknown
}

func isLastErrorUnknown(err error) (bool, error) {
	converted, ok := err.(errors.Error)
	if !ok {
		return true, nil
	}

	last := errors.GetLastError(converted)
	if last == nil {
		return true, nil
	}

	return isErrorUnknown(last)
}

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

func maybeReport(reporter *sentry.Client, err error, request *http.Request) {
	if ok, last := isLastErrorUnknown(err); ok {
		report(reporter, last, request)
	}
}

// NewErrors returns a custom echo's error handler.
//
// When the error is from errors.Error type, it will check for the layer and response with the appropriated HTTP status
// code. However, if the error is not from errors.Error type, it will respond with HTTP status code 500. When this error
// occurs, it will also try to send the error to Sentry.
func NewErrors(reporter *sentry.Client) func(error, echo.Context) {
	return func(err error, ctx echo.Context) {
		var status int

		if converted, ok := err.(errors.Error); ok {
			switch converted.Layer {
			case guard.ErrLayer:
				status = http.StatusForbidden
			case routes.ErrLayer:
				status = converter.FromErrRouteToHTTPStatus(converted.Code)
			case services.ErrLayer:
				status = converter.FromErrServiceToHTTPStatus(converted.Code)
			case store.ErrLayer:
				status = http.StatusInternalServerError
			default:
				status = http.StatusInternalServerError
			}

			maybeReport(reporter, converted, ctx.Request())
		} else if herr, ok := err.(*echo.HTTPError); ok {
			status = herr.Code
		} else {
			status = http.StatusInternalServerError

			report(reporter, err, ctx.Request())
		}

		ctx.NoContent(status) //nolint:errcheck
	}
}
