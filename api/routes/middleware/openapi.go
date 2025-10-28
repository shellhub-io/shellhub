package middleware

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/api/pkg/openapi"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/sirupsen/logrus"
)

var (
	// Global validator instance
	globalValidator *openapi.OpenAPIValidator
	validatorOnce   sync.Once
	validatorErr    error
)

// responseWriter is a custom response writer that captures the response body
type responseWriter struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body.Write(b)

	return rw.ResponseWriter.Write(b)
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// OpenAPIValidatorConfig holds the configuration for schema validation middleware
type OpenAPIValidatorConfig struct {
	// EnabledPaths specifies which paths to validate (nil = all paths)
	EnabledPaths []string
	// FailOnMismatch determines if validation failures should return HTTP errors
	FailOnMismatch bool
	// OpenAPIPath overrides the default schema path
	OpenAPIPath string
}

// OpenAPIValidator returns a middleware that validates API responses against OpenAPIValidator schema
func OpenAPIValidator(config ...OpenAPIValidatorConfig) echo.MiddlewareFunc {
	var cfg OpenAPIValidatorConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	if cfg.FailOnMismatch {
		cfg.FailOnMismatch = false
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			validator := getOrCreateValidator(cfg)
			if validator == nil {
				// NOTE: If validator failed to initialize, just continue without validation.
				return next(c)
			}

			if !validator.IsEnabled() {
				return next(c)
			}

			req := c.Request()
			res := c.Response()

			body := &bytes.Buffer{}
			rw := &responseWriter{
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
				logger.WithField("errors", result.Errors).Warn("OpenAPI response validation failed")

				res.Header().Set("X-OpenAPI-Validation", "failed")
				res.Header().Set("X-OpenAPI-Errors", result.Errors[0])

				if validator.ShouldFailOnMismatch() && cfg.FailOnMismatch {
					return echo.NewHTTPError(http.StatusInternalServerError, "Response validation failed: "+result.Errors[0])
				}
			}

			return err
		}
	}
}

// CriticalEndpointsOnly returns a middleware that only validates critical endpoints
func CriticalEndpointsOnly() echo.MiddlewareFunc {
	criticalPaths := []string{
		"/api/devices",
		"/api/sessions",
		"/api/namespaces",
		"/api/login",
		"/api/auth",
		"/api/sshkeys/public-keys",
		"/api/stats",
	}

	return OpenAPIValidator(OpenAPIValidatorConfig{
		EnabledPaths:   criticalPaths,
		FailOnMismatch: envs.IsDevelopment(),
	})
}

// DevelopmentValidation returns a middleware for development with full validation
func DevelopmentValidation() echo.MiddlewareFunc {
	if !envs.IsDevelopment() {
		// Return no-op middleware in production
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		}
	}

	return OpenAPIValidator(OpenAPIValidatorConfig{
		FailOnMismatch: true,
	})
}

// ProductionValidation returns a middleware for production with logging only
func ProductionValidation() echo.MiddlewareFunc {
	return OpenAPIValidator(OpenAPIValidatorConfig{
		FailOnMismatch: false,
	})
}

// getOrCreateValidator initializes or returns the global validator instance
func getOrCreateValidator(cfg OpenAPIValidatorConfig) *openapi.OpenAPIValidator {
	validatorOnce.Do(func() {
		logger := logrus.WithField("component", "openapi_validator")

		validatorConfig := &openapi.OpenAPIValidatorConfig{
			SchemaPath:     cfg.OpenAPIPath,
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
