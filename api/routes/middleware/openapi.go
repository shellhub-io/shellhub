package middleware

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/api/pkg/openapi"
	"github.com/sirupsen/logrus"
)

var (
	// Global validator instance
	globalValidator *openapi.OpenAPIValidator
	validatorOnce   sync.Once
	validatorErr    error
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

// OpenAPIValidatorConfig holds the configuration for schema validation middleware
type OpenAPIValidatorConfig struct {
	// EnabledPaths specifies which paths to validate (nil = all paths)
	EnabledPaths []string
	// FailOnMismatch determines if validation failures should return HTTP errors
	FailOnMismatch bool
	// SchemaPath overrides the default schema path
	SchemaPath *url.URL
}

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
		return func(c echo.Context) error {
			validator := getOrCreateValidator(*cfg)
			if validator == nil {
				return next(c)
			}

			if !validator.IsEnabled() {
				return next(c)
			}

			req := c.Request()
			res := c.Response()

			body := &bytes.Buffer{}

			rw := &capture{
				ResponseWriter: res.Writer,
				body:           body,
				statusCode:     200,
			}

			res.Writer = rw

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

// getOrCreateValidator initializes or returns the global validator instance
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

		globalValidator, validatorErr = openapi.NewOpenAPIValidator(ctx, validatorConfig)
		if validatorErr != nil {
			logger.WithError(validatorErr).Error("Failed to initialize OpenAPI validator")

			return
		}

		if globalValidator.IsEnabled() {
			logger.Info("OpenAPI response validation middleware initialized")
		}
	})

	if validatorErr != nil {
		return nil
	}

	return globalValidator
}
