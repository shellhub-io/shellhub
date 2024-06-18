package handlers

import (
	"net/http"
	"os"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/api/pkg/echo/handlers/pkg/converter"
	routes "github.com/shellhub-io/shellhub/api/routes/errors"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mongo"
	"github.com/shellhub-io/shellhub/pkg/errors"
)

func report(reporter *sentry.Client, err error, request *http.Request) {
	go func() {
		if reporter != nil {
			reporter.CaptureEvent(&sentry.Event{ //nolint:exhaustruct
				Level:   sentry.LevelError,
				Message: err.Error(),
				Request: sentry.NewRequest(request),
				Tags: map[string]string{
					"domain": os.Getenv("SHELLHUB_DOMAIN"),
				},
			}, &sentry.EventHint{Request: request}, &sentry.Scope{}) //nolint:exhaustruct
		}
	}()
}

// NewErrors returns a custom echo's error handler.
func NewErrors(reporter *sentry.Client) func(error, echo.Context) {
	return func(err error, ctx echo.Context) {
		// NOTE(r): The early return approach here, despite it being a bit verbose, is the best way to clarify what
		// happens in each case, avoiding the use of else statements, which would make the code more confusing or a big
		// switch statement, which would make the code less readable.

		// Every Mongo error that isn't mapped as a store error must be reported to Sentry and responded with HTTP
		// status code 500.
		if errors.Is(err, mongo.ErrMongo) {
			report(reporter, err, ctx.Request())
			ctx.NoContent(http.StatusInternalServerError) //nolint:errcheck

			return
		}

		// On HTTP errors, anything related to the HTTP protocol, we just return the error code, avoiding a 500 error.
		var herr *echo.HTTPError
		if ok := errors.As(err, &herr); ok {
			ctx.NoContent(herr.Code) //nolint:errcheck

			return
		}

		// When the error is a custom error, we need to check its layer to return the correct HTTP status code according
		// to the error's layer. Whether the error is not a custom error, we return a 500 error, because itn't from
		// Mongo, not even from HTTP, indicating that is something unknown by the API
		var e errors.Error
		if ok := errors.As(err, &e); !ok {
			ctx.NoContent(http.StatusInternalServerError) //nolint:errcheck

			return
		}

		var status int
		switch e.Layer {
		case routes.ErrLayer:
			status = converter.FromErrRouteToHTTPStatus(e.Code)
		case services.ErrLayer:
			status = converter.FromErrServiceToHTTPStatus(e.Code)
		case store.ErrLayer:
			// What happens when an error is returned directly from the store's layer, which means it doesn't have a
			// service error affecting it, which requires fixing.
			status = http.StatusInternalServerError
		}
		ctx.NoContent(status) //nolint:errcheck
	}
}
