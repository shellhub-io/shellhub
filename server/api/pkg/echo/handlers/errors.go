package handlers

import (
	"net/http"
	"os"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/shellhub-io/shellhub/server/api/pkg/echo/handlers/pkg/converter"
	routes "github.com/shellhub-io/shellhub/server/api/routes/errors"
	"github.com/shellhub-io/shellhub/server/api/services"
	"github.com/shellhub-io/shellhub/server/api/store"
)

// genericMessage is the message every 5xx answers with. An internal error's own text describes the
// server's internals, so only the reporter sees it.
const genericMessage = "internal server error"

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

// respond writes the error body. A 5xx never echoes the error's own message, and a status that
// forbids a body keeps its bare status.
func respond(ctx *echo.Context, status int, message string, fields map[string]string) {
	if status >= http.StatusInternalServerError {
		message, fields = genericMessage, nil
	}

	if status == http.StatusNoContent {
		ctx.NoContent(status) //nolint:errcheck

		return
	}

	ctx.JSON(status, responses.Error{Message: message, Fields: fields}) //nolint:errcheck
}

// echoMessage returns the client-safe message an error carrying an HTTP status describes itself
// with. It never falls back to Error(): those texts embed the internal cause, e.g.
// "code=400, message=failed to bind field value to int, err=strconv.ParseInt: …".
func echoMessage(err error, code int) string {
	// A *echo.BindingError unwraps to its cause rather than to the *echo.HTTPError it embeds, so
	// errors.As misses it on the check below. Match it first.
	var binding *echo.BindingError
	if errors.As(err, &binding) && binding.HTTPError != nil && binding.Message != "" {
		return binding.Message
	}

	var httpErr *echo.HTTPError
	if errors.As(err, &httpErr) && httpErr.Message != "" {
		return httpErr.Message
	}

	return statusText(code)
}

// statusText names a status for a client, falling back to the generic message for a code net/http
// does not name.
func statusText(code int) string {
	if text := http.StatusText(code); text != "" {
		return text
	}

	return genericMessage
}

// fieldsOf returns the per-field detail an error carries, or nil when it carries none.
func fieldsOf(e errors.Error) map[string]string {
	switch data := e.Data.(type) {
	case routes.ErrDataInvalidEntity:
		return data.Fields
	case services.ErrDataInvalidFields:
		return data.Fields
	default:
		return nil
	}
}

// NewErrors returns a custom echo's error handler.
func NewErrors(reporter *sentry.Client) echo.HTTPErrorHandler {
	return func(ctx *echo.Context, err error) {
		if errors.Is(err, store.ErrInternal) {
			report(reporter, err, ctx.Request())
			respond(ctx, http.StatusInternalServerError, "", nil)

			return
		}

		if code := echo.StatusCode(err); code != 0 {
			respond(ctx, code, echoMessage(err, code), nil)

			return
		}

		// When the error is a custom error, we need to check its layer to return the correct HTTP status code according
		// to the error's layer. Whether the error is not a custom error, we return a 500 error, because itn't from
		// Mongo, not even from HTTP, indicating that is something unknown by the API
		var e errors.Error
		if ok := errors.As(err, &e); !ok {
			respond(ctx, http.StatusInternalServerError, "", nil)

			return
		}

		var status int
		switch e.Layer {
		case routes.ErrLayer:
			status = converter.FromErrRouteToHTTPStatus(e.Code)
		case services.ErrLayer:
			status = converter.FromErrServiceToHTTPStatus(e.Code)
		case store.ErrLayer:
			status = http.StatusInternalServerError
		default:
			status = http.StatusInternalServerError
		}

		respond(ctx, status, e.Message, fieldsOf(e))
	}
}
