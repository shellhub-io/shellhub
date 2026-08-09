package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBinder(t *testing.T) {
	type request struct {
		Name string `param:"name" json:"name" query:"name"`
	}

	cases := []struct {
		description string
		setup       func() *echo.Context
		wantName    string
		wantErr     bool
	}{
		{
			description: "succeeds to bind json body",
			setup: func() *echo.Context {
				e := echo.New()
				req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"test"}`))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

				return e.NewContext(req, httptest.NewRecorder())
			},
			wantName: "test",
		},
		{
			description: "succeeds to bind path parameters",
			setup: func() *echo.Context {
				e := echo.New()
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				c := e.NewContext(req, httptest.NewRecorder())
				c.SetPathValues(echo.PathValues{{Name: "name", Value: "plain"}})

				return c
			},
			wantName: "plain",
		},
		{
			// Echo does not URL-decode path parameters - c.Param() returns the raw
			// value as extracted from the URL (e.g. "%40" instead of "@"). The binder
			// must decode them so the application never sees percent-encoded strings.
			description: "decodes URL-encoded path parameters",
			setup: func() *echo.Context {
				e := echo.New()
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				c := e.NewContext(req, httptest.NewRecorder())
				c.SetPathValues(echo.PathValues{{Name: "name", Value: "%40%40%40%40%40%40"}})

				return c
			},
			wantName: "@@@@@@",
		},
		{
			description: "succeeds to bind query parameters",
			setup: func() *echo.Context {
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
			setup: func() *echo.Context {
				e := echo.New()
				req := httptest.NewRequest(http.MethodGet, "/?name=%40%40%40", nil)

				return e.NewContext(req, httptest.NewRecorder())
			},
			wantName: "@@@",
		},
		{
			// A body-less request sent with Transfer-Encoding: chunked (resty does so
			// since v2.17) has no Content-Length, and Echo's binder would reject the
			// missing Content-Type with 415 instead of skipping the absent body.
			description: "succeeds to bind a body-less chunked request",
			setup: func() *echo.Context {
				e := echo.New()
				req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(""))
				req.ContentLength = -1
				c := e.NewContext(req, httptest.NewRecorder())
				c.SetPathValues(echo.PathValues{{Name: "name", Value: "test"}})

				return c
			},
			wantName: "test",
		},
		{
			description: "succeeds to bind a chunked json body",
			setup: func() *echo.Context {
				e := echo.New()
				req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"test"}`))
				req.ContentLength = -1
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

				return e.NewContext(req, httptest.NewRecorder())
			},
			wantName: "test",
		},
		{
			description: "fails to bind a chunked body without a content type",
			setup: func() *echo.Context {
				e := echo.New()
				req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"test"}`))
				req.ContentLength = -1

				return e.NewContext(req, httptest.NewRecorder())
			},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			b := NewBinder()
			var req request
			err := b.Bind(tc.setup(), &req)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tc.wantName, req.Name)
		})
	}
}
