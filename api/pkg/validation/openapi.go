package validation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/sirupsen/logrus"
)

// OpenAPIValidator validates HTTP responses against OpenAPI specification
type OpenAPIValidator struct {
	router          routers.Router
	enabled         bool
	failOnMismatch  bool
	enabledPaths    map[string]bool
	mu              sync.RWMutex
	logger          *logrus.Entry
}

// ValidationResult contains the result of response validation
type ValidationResult struct {
	Valid      bool
	Errors     []string
	Path       string
	Method     string
	StatusCode int
}

// OpenAPIValidatorConfig holds configuration for the validator
type OpenAPIValidatorConfig struct {
	// SchemaPath is the path to the OpenAPI schema file
	SchemaPath string
	// EnabledPaths are the paths that should be validated (nil = all paths)
	EnabledPaths []string
	// FailOnMismatch determines if validation failures should cause HTTP errors
	FailOnMismatch bool
	// Logger for validation messages
	Logger *logrus.Entry
}

// NewOpenAPIValidator creates a new OpenAPI response validator
func NewOpenAPIValidator(ctx context.Context, config *OpenAPIValidatorConfig) (*OpenAPIValidator, error) {
	if config == nil {
		config = &OpenAPIValidatorConfig{}
	}

	// Default logger if none provided
	if config.Logger == nil {
		config.Logger = logrus.NewEntry(logrus.StandardLogger())
	}

	validator := &OpenAPIValidator{
		enabled:        shouldEnableValidation(),
		failOnMismatch: config.FailOnMismatch && envs.IsDevelopment(),
		enabledPaths:   make(map[string]bool),
		logger:         config.Logger,
	}

	if !validator.enabled {
		config.Logger.Debug("OpenAPI validation disabled")
		return validator, nil
	}

	// Load OpenAPI schema
	if config.SchemaPath == "" {
		config.SchemaPath = getDefaultSchemaPath()
	}

	config.Logger.WithField("schema_path", config.SchemaPath).Info("Loading OpenAPI schema")

	loader := &openapi3.Loader{Context: ctx, IsExternalRefsAllowed: true}
	doc, err := loader.LoadFromFile(config.SchemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load OpenAPI schema: %w", err)
	}

	// Validate the schema itself
	if err := doc.Validate(ctx); err != nil {
		return nil, fmt.Errorf("invalid OpenAPI schema: %w", err)
	}

	// Create router for path matching
	router, err := gorillamux.NewRouter(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAPI router: %w", err)
	}

	validator.router = router

	// Set up enabled paths
	if len(config.EnabledPaths) > 0 {
		for _, path := range config.EnabledPaths {
			validator.enabledPaths[path] = true
		}
	}

	config.Logger.Info("OpenAPI response validator initialized successfully")
	return validator, nil
}

// ValidateResponse validates an HTTP response against the OpenAPI schema
func (v *OpenAPIValidator) ValidateResponse(r *http.Request, response *http.Response, responseBody []byte) *ValidationResult {
	result := &ValidationResult{
		Valid:      true,
		Path:       r.URL.Path,
		Method:     r.Method,
		StatusCode: response.StatusCode,
	}

	if !v.enabled {
		return result
	}

	v.mu.RLock()
	defer v.mu.RUnlock()

	// Check if this path should be validated
	if len(v.enabledPaths) > 0 && !v.enabledPaths[r.URL.Path] {
		return result
	}

	// Skip validation for certain paths
	if v.shouldSkipPath(r.URL.Path) {
		return result
	}

	// Find matching route in OpenAPI spec
	route, pathParams, err := v.router.FindRoute(r)
	if err != nil {
		// Path not found in OpenAPI spec - log but don't fail
		v.logger.WithFields(logrus.Fields{
			"path":   r.URL.Path,
			"method": r.Method,
			"error":  err.Error(),
		}).Debug("Path not found in OpenAPI spec")
		return result
	}

	// Create validation input
	requestValidationInput := &openapi3filter.RequestValidationInput{
		Request:    r,
		PathParams: pathParams,
		Route:      route,
	}

	// Create a response for validation
	responseForValidation := &http.Response{
		Status:        response.Status,
		StatusCode:    response.StatusCode,
		Header:        response.Header.Clone(),
		Body:          io.NopCloser(bytes.NewReader(responseBody)),
		ContentLength: int64(len(responseBody)),
	}

	responseValidationInput := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: requestValidationInput,
		Status:                 response.StatusCode,
		Header:                 response.Header,
		Body:                   io.NopCloser(bytes.NewReader(responseBody)),
	}

	// Validate response
	ctx := context.Background()
	if err := openapi3filter.ValidateResponse(ctx, responseValidationInput); err != nil {
		result.Valid = false
		result.Errors = []string{err.Error()}

		v.logger.WithFields(logrus.Fields{
			"path":        r.URL.Path,
			"method":      r.Method,
			"status_code": response.StatusCode,
			"error":       err.Error(),
		}).Warn("OpenAPI response validation failed")

		// Log response body for debugging in development
		if envs.IsDevelopment() {
			v.logResponseBody(responseBody)
		}
	} else {
		v.logger.WithFields(logrus.Fields{
			"path":        r.URL.Path,
			"method":      r.Method,
			"status_code": response.StatusCode,
		}).Debug("OpenAPI response validation passed")
	}

	responseForValidation.Body.Close()
	return result
}

// EnablePath enables validation for a specific path
func (v *OpenAPIValidator) EnablePath(path string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.enabledPaths[path] = true
}

// DisablePath disables validation for a specific path
func (v *OpenAPIValidator) DisablePath(path string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	delete(v.enabledPaths, path)
}

// ShouldFailOnMismatch returns whether validation failures should cause HTTP errors
func (v *OpenAPIValidator) ShouldFailOnMismatch() bool {
	return v.failOnMismatch
}

// IsEnabled returns whether validation is enabled
func (v *OpenAPIValidator) IsEnabled() bool {
	return v.enabled
}

// shouldEnableValidation determines if OpenAPI validation should be enabled
func shouldEnableValidation() bool {
	// Enable in development always, in production only if explicitly requested
	return envs.IsDevelopment() || os.Getenv("SHELLHUB_OPENAPI_VALIDATION") == "true"
}

// getDefaultSchemaPath returns the default path to the OpenAPI schema
func getDefaultSchemaPath() string {
	// Determine which schema to use based on environment
	if envs.IsCloud() {
		return "../openapi/spec/cloud-openapi.yaml"
	}
	if envs.IsEnterprise() {
		return "../openapi/spec/enterprise-openapi.yaml"
	}
	return "../openapi/spec/community-openapi.yaml"
}

// shouldSkipPath determines if a path should be skipped from validation
func (v *OpenAPIValidator) shouldSkipPath(path string) bool {
	// Skip internal endpoints
	if strings.HasPrefix(path, "/internal") {
		return true
	}

	// Skip health checks and metrics
	skipPaths := []string{
		"/api/healthcheck",
		"/metrics",
		"/openapi",
	}

	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}

	return false
}

// logResponseBody logs the response body for debugging
func (v *OpenAPIValidator) logResponseBody(body []byte) {
	if len(body) == 0 {
		v.logger.Debug("Response body is empty")
		return
	}

	// Try to format as JSON for better readability
	var formatted bytes.Buffer
	if err := json.Indent(&formatted, body, "", "  "); err == nil {
		v.logger.WithField("response_body", formatted.String()).Debug("Response body")
	} else {
		v.logger.WithField("response_body", string(body)).Debug("Response body")
	}
}