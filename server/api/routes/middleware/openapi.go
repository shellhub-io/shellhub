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

var (
	globalValidator *openapi.OpenAPIValidator
	validatorOnce   sync.Once
	errValidator    error
)

type capture struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (rw *capture) Write(b []byte) (int, error) {
	rw.body.Write(b)

	return rw.ResponseWriter.Write(b)
}

func (rw *capture) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Unwrap lets http.ResponseController reach the underlying writer, which is how echo resolves
// Hijack and Flush.
//
// It is not enough for the WebSocket routes: both libraries in use assert http.Hijacker on the
// writer directly and never consult Unwrap, and capture cannot satisfy it because embedding the
// http.ResponseWriter interface promotes only its three methods. Those routes are excluded from
// validation instead - see openAPIValidationSkipper in the server package.
func (rw *capture) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

// OpenAPIValidatorConfig holds the configuration for schema validation middleware
type OpenAPIValidatorConfig struct {
	// EnabledPaths specifies which paths to validate (nil = all paths)
	EnabledPaths []string
	// FailOnMismatch determines if validation failures should return HTTP errors
	FailOnMismatch bool
	// SchemaPath overrides the default schema path
	SchemaPath *url.URL
	// Skipper defines a function to skip middleware. If Skipper returns true, middleware is skipped.
	Skipper func(*echo.Context) bool
}

// OpenAPIValidationMessage is the body returned when a request does not match the spec,
// listing each violation so a client can see all of them at once.
type OpenAPIValidationMessage struct {
	Message string   `json:"message"`
	Errors  []string `json:"errors"`
}

// OpenAPIValidator returns a middleware that validates API responses against OpenAPIValidator schema
func OpenAPIValidator(cfg *OpenAPIValidatorConfig) echo.MiddlewareFunc {
	if cfg == nil {
		cfg = &OpenAPIValidatorConfig{}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if cfg.Skipper != nil && cfg.Skipper(c) {
				return next(c)
			}

			validator := getOrCreateValidator(*cfg)
			if validator == nil {
				return next(c)
			}

			req := c.Request()
			res := c.Response()

			body := &bytes.Buffer{}

			rw := &capture{
				ResponseWriter: res,
				body:           body,
				statusCode:     200,
			}

			c.SetResponse(rw)

			err := next(c)

			response := &http.Response{
				StatusCode:    rw.statusCode,
				Header:        res.Header(),
				Body:          io.NopCloser(bytes.NewReader(body.Bytes())),
				ContentLength: int64(body.Len()),
			}
			result := validator.ValidateResponse(req, response, body.Bytes())

			logger := logrus.WithFields(logrus.Fields{
				"path":        result.Path,
				"method":      result.Method,
				"status_code": result.StatusCode,
				"valid":       result.Valid,
			})

			if result.Valid {
				logger.Debug("OpenAPI response validation passed")
			} else {
				logger.WithField("error", result.Error).Warn("OpenAPI response validation failed")
			}

			return err
		}
	}
}

func getOrCreateValidator(cfg OpenAPIValidatorConfig) *openapi.OpenAPIValidator {
	validatorOnce.Do(func() {
		logger := logrus.WithField("component", "openapi_validator")

		validatorConfig := &openapi.OpenAPIValidatorConfig{
			SchemaPath:     cfg.SchemaPath,
			EnabledPaths:   cfg.EnabledPaths,
			FailOnMismatch: cfg.FailOnMismatch,
			Logger:         logger,
		}

		ctx := context.Background()

		globalValidator, errValidator = openapi.NewOpenAPIValidator(ctx, validatorConfig)
		if errValidator != nil {
			logger.WithError(errValidator).Error("Failed to initialize OpenAPI validator")

			return
		}
	})

	if errValidator != nil {
		return nil
	}

	return globalValidator
}
