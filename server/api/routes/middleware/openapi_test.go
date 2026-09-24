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

func specWithSchema(schema string) string {
	return `{
  "openapi": "3.0.3",
  "info": {"title": "things", "version": "1"},
  "paths": {
    "/api/thing": {
      "get": {
        "responses": {
          "200": {
            "description": "a thing",
            "content": {"application/json": {"schema": ` + schema + `}}
          }
        }
      }
    }
  },
  "components": {
    "schemas": {
      "base": {"type": "object", "properties": {
        "id": {"type": "string"},
        "owner": {"type": "object", "properties": {"name": {"type": "string"}}}
      }}
    }
  }
}`
}

func serveThings(t *testing.T, strict bool, path string, handler echo.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()

	return serveSpec(t, thingsSpec, strict, path, handler)
}

func serveSpec(t *testing.T, spec string, strict bool, path string, handler echo.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()

	cfg := &OpenAPIValidatorConfig{SchemaPath: specFile(t, spec), Strict: strict}

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

func requireRejected(t *testing.T, rec *httptest.ResponseRecorder, reason string) {
	t.Helper()

	require.Equal(t, http.StatusInternalServerError, rec.Code)

	message := OpenAPIValidationMessage{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &message))
	require.NotEmpty(t, message.Errors)
	require.Contains(t, message.Errors[0], reason)
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
		return c.JSON(http.StatusOK, map[string]int{})
	})

	requireRejected(t, rec, `property "id" is missing`)
}

func TestStrictValidatorReplacesAResponseCarryingAnUndeclaredProperty(t *testing.T) {
	rec := serveThings(t, true, "/api/things", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]any{"id": "thing-1", "extra": 1})
	})

	requireRejected(t, rec, `property "extra" is unsupported`)
}

func TestStrictValidatorAcceptsTheSiblingPropertiesOfAComposedSchema(t *testing.T) {
	spec := specWithSchema(`{"allOf": [
    {"$ref": "#/components/schemas/base"},
    {"type": "object", "properties": {"key": {"type": "string"}}}
  ]}`)

	rec := serveSpec(t, spec, true, "/api/thing", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"id": "thing-1", "key": "secret"})
	})

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.JSONEq(t, `{"id":"thing-1","key":"secret"}`, rec.Body.String())
}

func TestStrictValidatorReplacesAResponseCarryingAnUndeclaredPropertyInsideAComposedMember(t *testing.T) {
	spec := specWithSchema(`{"allOf": [
    {"$ref": "#/components/schemas/base"},
    {"type": "object", "properties": {"key": {"type": "string"}}}
  ]}`)

	rec := serveSpec(t, spec, true, "/api/thing", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]any{"id": "thing-1", "key": "secret", "owner": map[string]any{"name": "ada", "extra": 1}})
	})

	requireRejected(t, rec, `property "extra" is unsupported`)
}

func TestStrictValidatorReplacesAnUndeclaredPropertyInASchemaAnotherSiteComposes(t *testing.T) {
	spec := specWithSchema(`{"type": "object", "properties": {
    "alone": {"$ref": "#/components/schemas/base"},
    "either": {"oneOf": [{"$ref": "#/components/schemas/base"}, {"type": "string"}]}
  }}`)

	rec := serveSpec(t, spec, true, "/api/thing", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]any{"alone": map[string]any{"id": "thing-1", "extra": 1}})
	})

	requireRejected(t, rec, `property "extra" is unsupported`)
}

func TestStrictValidatorReplacesAResponseCarryingAnUndeclaredPropertyInANullableWrapper(t *testing.T) {
	spec := specWithSchema(`{"type": "object", "properties": {
    "base": {"nullable": true, "allOf": [{"$ref": "#/components/schemas/base"}]}
  }}`)

	rec := serveSpec(t, spec, true, "/api/thing", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]any{"base": map[string]string{"id": "thing-1", "extra": "x"}})
	})

	requireRejected(t, rec, `property "extra" is unsupported`)
}

func TestStrictValidatorAcceptsAnyKeyInATypedMap(t *testing.T) {
	spec := specWithSchema(`{"type": "object", "properties": {
    "id": {"type": "string"},
    "labels": {
      "type": "object",
      "properties": {"env": {"type": "string"}},
      "additionalProperties": {"type": "string"}
    }
  }}`)

	rec := serveSpec(t, spec, true, "/api/thing", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]any{"id": "thing-1", "labels": map[string]string{"env": "prod", "team": "core"}})
	})

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.JSONEq(t, `{"id":"thing-1","labels":{"env":"prod","team":"core"}}`, rec.Body.String())
}

func TestStrictValidatorReplacesAResponseCarryingAnUndeclaredPropertyBesideATypedMap(t *testing.T) {
	spec := specWithSchema(`{"type": "object", "properties": {
    "id": {"type": "string"},
    "labels": {
      "type": "object",
      "properties": {"env": {"type": "string"}},
      "additionalProperties": {"type": "string"}
    }
  }}`)

	rec := serveSpec(t, spec, true, "/api/thing", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]any{"id": "thing-1", "labels": map[string]string{"team": "core"}, "extra": 1})
	})

	requireRejected(t, rec, `property "extra" is unsupported`)
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
