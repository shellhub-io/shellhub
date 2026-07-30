package openapi

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	router         routers.Router
	failOnMismatch bool
	enabledPaths   map[string]bool
	mu             sync.RWMutex
	logger         *logrus.Entry
}

// ValidationResult contains the result of response validation
type ValidationResult struct {
	Valid      bool
	Error      string
	Path       string
	Method     string
	StatusCode int
}

// OpenAPIValidatorConfig holds configuration for the validator
type OpenAPIValidatorConfig struct {
	// SchemaPath is the URL to the OpenAPI schema.
	SchemaPath *url.URL
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

	if config.Logger == nil {
		config.Logger = logrus.NewEntry(logrus.StandardLogger())
	}

	validator := &OpenAPIValidator{
		failOnMismatch: config.FailOnMismatch && envs.IsDevelopment(),
		enabledPaths:   make(map[string]bool),
		logger:         config.Logger,
	}

	if config.SchemaPath == nil {
		config.SchemaPath = GetDefaultSchemaPath()
	}

	if config.SchemaPath == nil {
		return nil, fmt.Errorf("OpenAPI schema path is not defined")
	}

	loader := &openapi3.Loader{
		Context:               ctx,
		IsExternalRefsAllowed: true,
	}

	doc, err := loader.LoadFromURI(config.SchemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load OpenAPI schema: %w", err)
	}

	if err := doc.Validate(ctx); err != nil {
		return nil, fmt.Errorf("invalid OpenAPI schema: %w", err)
	}

	router, err := gorillamux.NewRouter(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAPI router: %w", err)
	}

	validator.router = router

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

	v.mu.RLock()
	defer v.mu.RUnlock()

	if len(v.enabledPaths) > 0 && !v.enabledPaths[r.URL.Path] {
		return result
	}

	route, pathParams, err := v.router.FindRoute(r)
	if err != nil {
		v.logger.WithFields(logrus.Fields{
			"path":   r.URL.Path,
			"method": r.Method,
			"error":  err.Error(),
		}).Debug("Path not found in OpenAPI spec")

		return result
	}

	v.logger.WithFields(logrus.Fields{
		"path":   r.URL.Path,
		"method": r.Method,
	}).Debug("Path found in OpenAPI spec, proceeding with validation")

	requestValidationInput := &openapi3filter.RequestValidationInput{
		Request:    r,
		PathParams: pathParams,
		Route:      route,
	}

	responseValidationInput := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: requestValidationInput,
		Status:                 response.StatusCode,
		Header:                 response.Header,
		Body:                   io.NopCloser(bytes.NewReader(responseBody)),
	}

	ctx := context.Background()

	if err := openapi3filter.ValidateResponse(ctx, responseValidationInput); err != nil {
		result.Valid = false
		result.Error = err.Error()

		v.logger.WithFields(logrus.Fields{
			"path":        r.URL.Path,
			"method":      r.Method,
			"status_code": response.StatusCode,
			"error":       err.Error(),
		}).Trace("OpenAPI response validation failed")
	} else {
		v.logger.WithFields(logrus.Fields{
			"path":        r.URL.Path,
			"method":      r.Method,
			"status_code": response.StatusCode,
		}).Trace("OpenAPI response validation passed")
	}

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

// GetDefaultSchemaPath returns the default path to the OpenAPI schema
func GetDefaultSchemaPath() *url.URL {
	// NOTE: This path refers to the generated OpenAPI spec file.
	// TODO: Make this configurable via environment variable if needed.
	u, err := url.Parse("http://openapi:8080/openapi/openapi.json")
	if err != nil {
		return nil
	}

	return u
}
