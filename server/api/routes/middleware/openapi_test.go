package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"
)

const thingsSpec = `{
  "openapi": "3.0.3",
  "info": {"title": "things", "version": "1"},
  "paths": {
    "/api/things": {
      "get": {
        "responses": {
          "200": {
            "description": "a thing",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "required": ["id"],
                  "properties": {"id": {"type": "string"}}
                }
              }
            }
          }
        }
      }
    }
  }
}`

func specFile(t *testing.T, content string) *url.URL {
	t.Helper()

	path := filepath.Join(t.TempDir(), "openapi.json")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	return &url.URL{Scheme: "file", Path: path}
}

func serveThings(t *testing.T, strict bool, path string, handler echo.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()

	cfg := &OpenAPIValidatorConfig{SchemaPath: specFile(t, thingsSpec), Strict: strict}

	validate := OpenAPIValidator(cfg)
	if strict {
		var err error

		validate, err = LoadOpenAPIValidator(context.Background(), cfg)
		require.NoError(t, err)
	}

	e := echo.New()
	e.Use(validate)
	e.GET(path, handler)

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, path, nil))

	return rec
}

func TestStrictValidatorAnswersAsTheHandlerDidWhenTheResponseMatches(t *testing.T) {
	rec := serveThings(t, true, "/api/things", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"id": "thing-1"})
	})

	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, `{"id":"thing-1"}`, rec.Body.String())
}

func TestStrictValidatorReplacesAResponseTheSchemaRejects(t *testing.T) {
	rec := serveThings(t, true, "/api/things", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]int{"count": 1})
	})

	require.Equal(t, http.StatusInternalServerError, rec.Code)

	message := OpenAPIValidationMessage{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &message))
	require.NotEmpty(t, message.Errors)
	require.Contains(t, message.Errors[0], "id")
}

func TestStrictValidatorReplacesAResponseNoRouteDeclares(t *testing.T) {
	rec := serveThings(t, true, "/api/nothing", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"id": "thing-1"})
	})

	require.Equal(t, http.StatusInternalServerError, rec.Code)

	message := OpenAPIValidationMessage{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &message))
	require.NotEmpty(t, message.Errors)
}

func TestStrictValidatorLetsAHandlerErrorReachTheClient(t *testing.T) {
	rec := serveThings(t, true, "/api/things", func(*echo.Context) error {
		return echo.NewHTTPError(http.StatusTeapot, "short and stout")
	})

	require.Equal(t, http.StatusTeapot, rec.Code)
	require.Contains(t, rec.Body.String(), "short and stout")
}

func TestReportingValidatorAnswersAsTheHandlerDidWhenTheResponseMismatches(t *testing.T) {
	rec := serveThings(t, false, "/api/things", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]int{"count": 1})
	})

	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, `{"count":1}`, rec.Body.String())
}

func TestLoadOpenAPIValidatorFailsWithoutASchema(t *testing.T) {
	missing := &url.URL{Scheme: "file", Path: filepath.Join(t.TempDir(), "absent.json")}

	_, err := LoadOpenAPIValidator(context.Background(), &OpenAPIValidatorConfig{SchemaPath: missing, Strict: true})
	require.Error(t, err)
}
