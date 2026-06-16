package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBinder(t *testing.T) {
	type request struct {
		Name string `param:"name" json:"name" query:"name"`
	}

	cases := []struct {
		description string
		setup       func() echo.Context
		wantName    string
		wantErr     bool
	}{
		{
			description: "succeeds to bind json body",
			setup: func() echo.Context {
				e := echo.New()
				req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"test"}`))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

				return e.NewContext(req, httptest.NewRecorder())
			},
			wantName: "test",
		},
		{
			description: "succeeds to bind path parameters",
			setup: func() echo.Context {
				e := echo.New()
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				c := e.NewContext(req, httptest.NewRecorder())
				c.SetParamNames("name")
				c.SetParamValues("plain")

				return c
			},
			wantName: "plain",
		},
		{
			// Echo does not URL-decode path parameters - c.Param() returns the raw
			// value as extracted from the URL (e.g. "%40" instead of "@"). The binder
			// must decode them so the application never sees percent-encoded strings.
			description: "decodes URL-encoded path parameters",
			setup: func() echo.Context {
				e := echo.New()
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				c := e.NewContext(req, httptest.NewRecorder())
				c.SetParamNames("name")
				c.SetParamValues("%40%40%40%40%40%40")

				return c
			},
			wantName: "@@@@@@",
		},
		{
			description: "succeeds to bind query parameters",
			setup: func() echo.Context {
				e := echo.New()
				req := httptest.NewRequest(http.MethodGet, "/?name=test", nil)

				return e.NewContext(req, httptest.NewRecorder())
			},
			wantName: "test",
		},
		{
			// Unlike path parameters, query strings are decoded automatically by
			// Go's url.ParseQuery before Echo touches them, so no binder intervention
			// is needed for this case.
			description: "succeeds to bind query parameters with special characters",
			setup: func() echo.Context {
				e := echo.New()
				req := httptest.NewRequest(http.MethodGet, "/?name=%40%40%40", nil)

				return e.NewContext(req, httptest.NewRecorder())
			},
			wantName: "@@@",
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			b := NewBinder()
			var req request
			err := b.Bind(&req, tc.setup())
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tc.wantName, req.Name)
		})
	}
}
