package handlers

import (
	"bufio"
	"context"
	"io"
	"maps"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBinder(t *testing.T) {
	type request struct {
		Body  string `json:"name"`
		Path  string `param:"name" json:"-"`
		Query string `query:"name" json:"-"`
	}

	cases := []struct {
		description string
		setup       func() *echo.Context
		want        request
		wantErr     bool
	}{
		{
			description: "succeeds to bind json body",
			setup: func() *echo.Context {
				e := echo.New()
				req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", strings.NewReader(`{"name":"test"}`))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

				return e.NewContext(req, httptest.NewRecorder())
			},
			want: request{Body: "test"},
		},
		{
			description: "succeeds to bind path parameters",
			setup: func() *echo.Context {
				e := echo.New()
				req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
				c := e.NewContext(req, httptest.NewRecorder())
				c.SetPathValues(echo.PathValues{{Name: "name", Value: "plain"}})

				return c
			},
			want: request{Path: "plain"},
		},
		{
			description: "decodes URL-encoded path parameters",
			setup: func() *echo.Context {
				e := echo.New()
				req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
				c := e.NewContext(req, httptest.NewRecorder())
				c.SetPathValues(echo.PathValues{{Name: "name", Value: "%40%40%40%40%40%40"}})

				return c
			},
			want: request{Path: "@@@@@@"},
		},
		{
			description: "succeeds to bind query parameters",
			setup: func() *echo.Context {
				e := echo.New()
				req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/?name=test", nil)

				return e.NewContext(req, httptest.NewRecorder())
			},
			want: request{Query: "test"},
		},
		{
			description: "succeeds to bind query parameters with special characters",
			setup: func() *echo.Context {
				e := echo.New()
				req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/?name=%40%40%40", nil)

				return e.NewContext(req, httptest.NewRecorder())
			},
			want: request{Query: "@@@"},
		},
		{
			description: "succeeds to bind a body-less chunked request",
			setup: func() *echo.Context {
				e := echo.New()
				req := httptest.NewRequestWithContext(context.Background(), http.MethodPatch, "/", strings.NewReader(""))
				req.ContentLength = -1
				c := e.NewContext(req, httptest.NewRecorder())
				c.SetPathValues(echo.PathValues{{Name: "name", Value: "test"}})

				return c
			},
			want: request{Path: "test"},
		},
		{
			description: "succeeds to bind a chunked json body",
			setup: func() *echo.Context {
				e := echo.New()
				req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", strings.NewReader(`{"name":"test"}`))
				req.ContentLength = -1
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

				return e.NewContext(req, httptest.NewRecorder())
			},
			want: request{Body: "test"},
		},
		{
			description: "fails to bind a chunked body without a content type",
			setup: func() *echo.Context {
				e := echo.New()
				req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", strings.NewReader(`{"name":"test"}`))
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
			assert.Equal(t, tc.want, req)
		})
	}
}

func TestBinderOverTheWire(t *testing.T) {
	const chunked = "Transfer-Encoding: chunked\r\n"

	cases := []struct {
		description string
		head        string
		body        string
		wantStatus  int
	}{
		{
			description: "a body ending before its Content-Length answers 400 and names the read failure",
			head:        "Content-Length: 100\r\n",
			body:        `{"name":`,
			wantStatus:  http.StatusBadRequest,
		},
		{
			description: "a chunked body cut inside a chunk answers 400 and names the read failure",
			head:        chunked,
			body:        "20\r\n{\"name\":",
			wantStatus:  http.StatusBadRequest,
		},
		{
			description: "a chunked body cut before its first chunk answers 400 and names the read failure",
			head:        chunked,
			wantStatus:  http.StatusBadRequest,
		},
		{
			description: "a body ending before its Content-Length answers 400 even when what arrived cannot parse",
			head:        "Content-Length: 100\r\n",
			body:        `not json`,
			wantStatus:  http.StatusBadRequest,
		},
		{
			description: "a complete body holding an unfinished json value answers 422",
			head:        "Content-Length: 8\r\n",
			body:        `{"name":`,
			wantStatus:  http.StatusUnprocessableEntity,
		},
		{
			description: "a complete chunked body holding an unfinished json value answers 422",
			head:        chunked,
			body:        "1\r\n{\r\n0\r\n\r\n",
			wantStatus:  http.StatusUnprocessableEntity,
		},
		{
			description: "a complete body holding invalid json answers 422",
			head:        "Content-Length: 8\r\n",
			body:        `not json`,
			wantStatus:  http.StatusUnprocessableEntity,
		},
		{
			description: "a complete body holding a value of the wrong type answers 422",
			head:        "Content-Length: 10\r\n",
			body:        `{"name":1}`,
			wantStatus:  http.StatusUnprocessableEntity,
		},
		{
			description: "a complete chunked body answers 200",
			head:        chunked,
			body:        "f\r\n{\"name\":\"test\"}\r\n0\r\n\r\n",
			wantStatus:  http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			bindErrs := make(chan error, 1)

			server := httptest.NewServer(newBindingRouter(bindErrs))
			t.Cleanup(server.Close)

			var dialer net.Dialer
			conn, err := dialer.DialContext(t.Context(), "tcp", server.Listener.Addr().String())
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, conn.Close()) })

			tcpConn, ok := conn.(*net.TCPConn)
			require.True(t, ok)

			request := "POST / HTTP/1.1\r\nHost: test\r\nContent-Type: application/json\r\n" + tc.head + "\r\n" + tc.body
			_, err = conn.Write([]byte(request))
			require.NoError(t, err)
			require.NoError(t, tcpConn.CloseWrite())

			response, err := http.ReadResponse(bufio.NewReader(conn), nil)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, response.Body.Close()) })

			assert.Equal(t, tc.wantStatus, response.StatusCode)

			bindErr := <-bindErrs
			if tc.wantStatus == http.StatusBadRequest {
				const cause = "read request body: unexpected EOF"
				require.ErrorContains(t, bindErr, cause)

				responseBody, err := io.ReadAll(response.Body)
				require.NoError(t, err)
				assert.NotContains(t, string(responseBody), cause)
			}
		})
	}
}

func TestBinderBodyReadFailure(t *testing.T) {
	cases := []struct {
		description string
		failure     error
		wantStatus  int
		wantCause   string
	}{
		{
			description: "a read deadline answers 400 and names the read failure",
			failure:     os.ErrDeadlineExceeded,
			wantStatus:  http.StatusBadRequest,
			wantCause:   "read request body: " + os.ErrDeadlineExceeded.Error(),
		},
		{
			description: "the body limit keeps answering 413",
			failure:     echo.ErrStatusRequestEntityTooLarge,
			wantStatus:  http.StatusRequestEntityTooLarge,
			wantCause:   echo.ErrStatusRequestEntityTooLarge.Error(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			bindErrs := make(chan error, 1)

			body := io.MultiReader(strings.NewReader(`{"name":`), iotest.ErrReader(tc.failure))
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", body)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			rec := httptest.NewRecorder()
			newBindingRouter(bindErrs).ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatus, rec.Code)
			require.ErrorContains(t, <-bindErrs, tc.wantCause)
		})
	}
}

func newBindingRouter(bindErrs chan<- error) *echo.Echo {
	router := echo.New()
	router.Binder = NewBinder()
	router.HTTPErrorHandler = NewErrors(nil)
	router.POST("/", func(c *echo.Context) error {
		var req struct {
			Name string `json:"name"`
		}

		err := c.Bind(&req)
		bindErrs <- err

		return err
	})

	return router
}

func TestBinderKeepsEachFieldToItsSource(t *testing.T) {
	type request struct {
		Tenant  string `param:"tenant" json:"-"`
		UserID  string `header:"X-ID" json:"-"`
		IsAdmin bool   `header:"X-Admin" json:"-"`
		Name    string `json:"name"`
	}

	type untagged struct {
		Name string
	}

	cases := []struct {
		description string
		contentType string
		header      http.Header
		body        string
		bind        func(*echo.Context) (any, error)
		wantStatus  int
		wantBound   any
	}{
		{
			description: "a body naming a path field answers 422",
			contentType: echo.MIMEApplicationJSON,
			body:        `{"Tenant":"other","name":"x"}`,
			wantStatus:  http.StatusUnprocessableEntity,
		},
		{
			description: "a body naming a header field the request lacks answers 422",
			contentType: echo.MIMEApplicationJSON,
			body:        `{"userid":"other","isAdmin":true}`,
			wantStatus:  http.StatusUnprocessableEntity,
		},
		{
			description: "an XML body answers 415",
			contentType: echo.MIMEApplicationXML,
			body:        `<request><Tenant>other</Tenant></request>`,
			wantStatus:  http.StatusUnsupportedMediaType,
		},
		{
			description: "each field binds from its own source",
			contentType: echo.MIMEApplicationJSON,
			header:      http.Header{"X-Id": {"caller"}},
			body:        `{"name":"renamed","unknown":1}`,
			wantStatus:  http.StatusOK,
			wantBound:   request{Tenant: "own", UserID: "caller", Name: "renamed"},
		},
		{
			description: "a request type with an untagged field answers 500",
			contentType: echo.MIMEApplicationJSON,
			body:        `{"name":"x"}`,
			bind: func(c *echo.Context) (any, error) {
				var req untagged
				err := c.Bind(&req)

				return req, err
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			description: "a map target still binds the body",
			contentType: echo.MIMEApplicationJSON,
			body:        `{"Tenant":"any"}`,
			bind: func(c *echo.Context) (any, error) {
				req := map[string]any{}
				err := c.Bind(&req)

				return req["Tenant"], err
			},
			wantStatus: http.StatusOK,
			wantBound:  "any",
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			bind := tc.bind
			if bind == nil {
				bind = func(c *echo.Context) (any, error) {
					var req request
					err := c.Bind(&req)

					return req, err
				}
			}

			var bound any
			router := echo.New()
			router.Binder = NewBinder()
			router.HTTPErrorHandler = NewErrors(nil)
			router.PUT("/namespaces/:tenant", func(c *echo.Context) error {
				var err error
				bound, err = bind(c)
				if err != nil {
					return err
				}

				return c.NoContent(http.StatusOK)
			})

			req := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/namespaces/own", strings.NewReader(tc.body))
			req.Header.Set(echo.HeaderContentType, tc.contentType)
			maps.Copy(req.Header, tc.header)

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatus, rec.Code)
			if tc.wantBound != nil {
				assert.Equal(t, tc.wantBound, bound)
			}
		})
	}
}
