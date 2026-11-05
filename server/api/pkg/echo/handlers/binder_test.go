package handlers

import (
	"bufio"
	"context"
	"io"
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
				req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", strings.NewReader(`{"name":"test"}`))
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

				return e.NewContext(req, httptest.NewRecorder())
			},
			wantName: "test",
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
			wantName: "plain",
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
			wantName: "@@@@@@",
		},
		{
			description: "succeeds to bind query parameters",
			setup: func() *echo.Context {
				e := echo.New()
				req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/?name=test", nil)

				return e.NewContext(req, httptest.NewRecorder())
			},
			wantName: "test",
		},
		{
			description: "succeeds to bind query parameters with special characters",
			setup: func() *echo.Context {
				e := echo.New()
				req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/?name=%40%40%40", nil)

				return e.NewContext(req, httptest.NewRecorder())
			},
			wantName: "@@@",
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
			wantName: "test",
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
			wantName: "test",
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
			assert.Equal(t, tc.wantName, req.Name)
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
