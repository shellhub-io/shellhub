package middleware

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"sync"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/server/api/pkg/openapi"
	"github.com/sirupsen/logrus"
)

type capture struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
	held       bool
}

func (rw *capture) Write(b []byte) (int, error) {
	if rw.held {
		return rw.body.Write(b)
	}

	rw.body.Write(b)

	return rw.ResponseWriter.Write(b)
}

func (rw *capture) WriteHeader(statusCode int) {
	rw.statusCode = statusCode

	if rw.held {
		return
	}

	rw.ResponseWriter.WriteHeader(statusCode)
}

// Unwrap lets http.ResponseController reach the underlying writer, which is how echo resolves
// Hijack and Flush.
//
// It is not enough for the WebSocket routes: both libraries in use assert http.Hijacker on the
// writer directly and never consult Unwrap, and capture cannot satisfy it because embedding the
// http.ResponseWriter interface promotes only its three methods. Those routes are excluded from
// validation instead; see openAPIValidationSkipper in the server package.
func (rw *capture) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

// OpenAPIValidatorConfig configures the response validator middleware.
type OpenAPIValidatorConfig struct {
	// EnabledPaths restricts validation to these paths; nil validates every path.
	EnabledPaths []string
	// SchemaPath overrides where the schema is loaded from.
	SchemaPath *url.URL
	// Skipper exempts a request from validation when it returns true.
	Skipper func(*echo.Context) bool
	// Strict holds each response until it is validated and replaces one the schema does not
	// describe with a 500 listing the mismatch, instead of answering as the handler did.
	Strict bool
}

// OpenAPIValidationMessage is the body a strict validator answers with in place of a response
// that does not match the spec, listing each violation so a client can see all of them at once.
type OpenAPIValidationMessage struct {
	Message string   `json:"message"`
	Errors  []string `json:"errors"`
}

type responseValidator struct {
	cfg       OpenAPIValidatorConfig
	once      sync.Once
	validator *openapi.OpenAPIValidator
	err       error
}

// OpenAPIValidator validates every response against the schema once it loads and logs what
// does not match. A schema that fails to load is reported once and every response passes.
func OpenAPIValidator(cfg *OpenAPIValidatorConfig) echo.MiddlewareFunc {
	return newResponseValidator(cfg).middleware()
}

// LoadOpenAPIValidator loads the schema before returning the middleware, so a schema that does
// not load is an error the caller can refuse to start on.
func LoadOpenAPIValidator(ctx context.Context, cfg *OpenAPIValidatorConfig) (echo.MiddlewareFunc, error) {
	r := newResponseValidator(cfg)
	if r.load(ctx) == nil {
		return nil, r.err
	}

	return r.middleware(), nil
}

func newResponseValidator(cfg *OpenAPIValidatorConfig) *responseValidator {
	if cfg == nil {
		cfg = &OpenAPIValidatorConfig{}
	}

	return &responseValidator{cfg: *cfg}
}

func (r *responseValidator) load(ctx context.Context) *openapi.OpenAPIValidator {
	r.once.Do(func() {
		logger := logrus.WithField("component", "openapi_validator")

		r.validator, r.err = openapi.NewOpenAPIValidator(ctx, &openapi.OpenAPIValidatorConfig{
			SchemaPath:   r.cfg.SchemaPath,
			EnabledPaths: r.cfg.EnabledPaths,
			Logger:       logger,
		})
		if r.err != nil {
			logger.WithError(r.err).Error("Failed to initialize OpenAPI validator")
		}
	})

	if r.err != nil {
		return nil
	}

	return r.validator
}

func (r *responseValidator) middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if r.cfg.Skipper != nil && r.cfg.Skipper(c) {
				return next(c)
			}

			validator := r.load(context.WithoutCancel(c.Request().Context()))
			if validator == nil {
				return next(c)
			}

			req := c.Request()
			res := c.Response()

			rw := &capture{
				ResponseWriter: res,
				body:           &bytes.Buffer{},
				statusCode:     http.StatusOK,
				held:           r.cfg.Strict,
			}

			c.SetResponse(rw)

			if err := next(c); err != nil {
				c.SetResponse(res)

				return err
			}

			result := validator.ValidateResponse(req, &http.Response{
				StatusCode:    rw.statusCode,
				Header:        res.Header(),
				Body:          io.NopCloser(bytes.NewReader(rw.body.Bytes())),
				ContentLength: int64(rw.body.Len()),
			}, rw.body.Bytes())

			logger := logrus.WithFields(logrus.Fields{
				"path":        result.Path,
				"method":      result.Method,
				"status_code": result.StatusCode,
				"outcome":     result.Outcome,
			})

			switch result.Outcome {
			case openapi.OutcomeFailed:
				logger.WithField("error", result.Error).Warn("OpenAPI response validation failed")
			case openapi.OutcomeUndeclared:
				logger.WithField("error", result.Error).Warn("no OpenAPI route describes this response")
			case openapi.OutcomePassed:
				logger.Debug("OpenAPI response validation passed")
			case openapi.OutcomeSkipped:
				logger.Debug("OpenAPI response validation skipped")
			}

			if !r.cfg.Strict {
				return nil
			}

			c.SetResponse(res)

			if result.Outcome == openapi.OutcomeFailed || result.Outcome == openapi.OutcomeUndeclared {
				res.Header().Del(echo.HeaderContentLength)

				return c.JSON(http.StatusInternalServerError, OpenAPIValidationMessage{
					Message: "the response does not match the OpenAPI schema",
					Errors:  []string{result.Error},
				})
			}

			res.WriteHeader(rw.statusCode)
			_, err := res.Write(rw.body.Bytes())

			return err
		}
	}
}
