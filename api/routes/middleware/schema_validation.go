package middleware

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/api/pkg/validation"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/sirupsen/logrus"
)

var (
	// Global validator instance
	globalValidator *validation.OpenAPIValidator
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

// SchemaValidationConfig holds the configuration for schema validation middleware
type SchemaValidationConfig struct {
	// EnabledPaths specifies which paths to validate (nil = all paths)
	EnabledPaths []string
	// FailOnMismatch determines if validation failures should return HTTP errors
	FailOnMismatch bool
	// SchemaPath overrides the default schema path
	SchemaPath string
}

// SchemaValidation returns a middleware that validates API responses against OpenAPI schema
func SchemaValidation(config ...SchemaValidationConfig) echo.MiddlewareFunc {
	var cfg SchemaValidationConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	// Set default fail behavior
	if cfg.FailOnMismatch && !envs.IsDevelopment() {
		cfg.FailOnMismatch = false // Never fail in production
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Initialize validator once
			validator := getOrCreateValidator(cfg)
			if validator == nil {
				// If validator failed to initialize, just continue without validation
				return next(c)
			}

			// Skip if validation is disabled
			if !validator.IsEnabled() {
				return next(c)
			}

			// Create custom response writer to capture response
			req := c.Request()
			res := c.Response()
			
			body := &bytes.Buffer{}
			rw := &responseWriter{
				ResponseWriter: res.Writer,
				body:          body,
				statusCode:    200, // Default status code
			}
			res.Writer = rw

			// Call the next handler
			err := next(c)

			// Perform validation
			response := &http.Response{
				StatusCode:    rw.statusCode,
				Header:        res.Header(),
				Body:          io.NopCloser(bytes.NewReader(body.Bytes())),
				ContentLength: int64(body.Len()),
			}

			result := validator.ValidateResponse(req, response, body.Bytes())

			// Log validation results
			logger := logrus.WithFields(logrus.Fields{
				"path":         result.Path,
				"method":       result.Method,
				"status_code":  result.StatusCode,
				"valid":        result.Valid,
			})

			if result.Valid {
				logger.Debug("OpenAPI response validation passed")
			} else {
				logger.WithField("errors", result.Errors).Warn("OpenAPI response validation failed")

				// Add validation headers for debugging
				res.Header().Set("X-OpenAPI-Validation", "failed")
				res.Header().Set("X-OpenAPI-Errors", result.Errors[0]) // First error only due to header length limits

				// Fail the request if configured to do so (only in development)
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

	return SchemaValidation(SchemaValidationConfig{
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

	return SchemaValidation(SchemaValidationConfig{
		FailOnMismatch: true,
	})
}

// ProductionValidation returns a middleware for production with logging only
func ProductionValidation() echo.MiddlewareFunc {
	return SchemaValidation(SchemaValidationConfig{
		FailOnMismatch: false,
	})
}

// getOrCreateValidator initializes or returns the global validator instance
func getOrCreateValidator(cfg SchemaValidationConfig) *validation.OpenAPIValidator {
	validatorOnce.Do(func() {
		logger := logrus.WithField("component", "openapi_validator")
		
		validatorConfig := &validation.OpenAPIValidatorConfig{
			SchemaPath:     cfg.SchemaPath,
			EnabledPaths:   cfg.EnabledPaths,
			FailOnMismatch: cfg.FailOnMismatch,
			Logger:         logger,
		}

		ctx := context.Background()
		globalValidator, validatorErr = validation.NewOpenAPIValidator(ctx, validatorConfig)
		
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

// UpdateValidatorPaths allows dynamic updating of validated paths
func UpdateValidatorPaths(enablePaths []string, disablePaths []string) {
	if globalValidator == nil {
		return
	}

	for _, path := range enablePaths {
		globalValidator.EnablePath(path)
	}

	for _, path := range disablePaths {
		globalValidator.DisablePath(path)
	}
}