package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestDecodeParam(t *testing.T) {
	cases := []struct {
		description string
		paramNames  []string
		paramValues []string
		target      string
		expected    string
	}{
		{
			description: "decodes a percent-encoded value",
			paramNames:  []string{"fingerprint"},
			paramValues: []string{"0f%3A8c%3Aae%3Af4"},
			target:      "fingerprint",
			expected:    "0f:8c:ae:f4",
		},
		{
			description: "leaves an already decoded value unchanged",
			paramNames:  []string{"fingerprint"},
			paramValues: []string{"0f:8c:ae:f4"},
			target:      "fingerprint",
			expected:    "0f:8c:ae:f4",
		},
		{
			description: "leaves an invalid percent-encoding unchanged",
			paramNames:  []string{"fingerprint"},
			paramValues: []string{"0f%zz8c"},
			target:      "fingerprint",
			expected:    "0f%zz8c",
		},
		{
			description: "only decodes the named param",
			paramNames:  []string{"tenant", "fingerprint"},
			paramValues: []string{"a%3Ab", "0f%3A8c"},
			target:      "fingerprint",
			expected:    "0f:8c",
		},
		{
			description: "is a no-op when the named param is absent",
			paramNames:  []string{"tenant"},
			paramValues: []string{"a%3Ab"},
			target:      "fingerprint",
			expected:    "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames(tc.paramNames...)
			c.SetParamValues(tc.paramValues...)

			next := func(echo.Context) error { return nil }
			_ = DecodeParam(tc.target)(next)(c)

			assert.Equal(t, tc.expected, c.Param(tc.target))
		})
	}
}
