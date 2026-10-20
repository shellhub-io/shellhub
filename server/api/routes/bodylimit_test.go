package routes

import (
	"bufio"
	"bytes"
	"io"
	"math"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	routesmiddleware "github.com/shellhub-io/shellhub/server/api/routes/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type countingReader struct {
	reader io.Reader
	read   int
}

func (r *countingReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	r.read += n

	return n, err
}

func jsonBodyOfSize(size int) []byte {
	const envelope = `{"data":""}`

	return []byte(`{"data":"` + strings.Repeat("x", size-len(envelope)) + `"}`)
}

const bodyLimitProbe = "/api/body-limit-probe"

func bodyLimitProbeRouter(t *testing.T, route ...echo.MiddlewareFunc) *echo.Echo {
	t.Helper()

	router, authn, _ := authenticatedRouter(t)

	router.POST(bodyLimitProbe, func(c *echo.Context) error {
		var body struct {
			Data string `json:"data"`
		}

		if err := c.Bind(&body); err != nil {
			return err
		}

		return c.NoContent(http.StatusOK)
	}, route...)
	authn.AllowAnonymous(http.MethodPost, bodyLimitProbe)

	return router
}

func TestRouterLimitsRequestBodies(t *testing.T) {
	const mib = 1 << 20

	tests := []struct {
		description string
		route       []echo.MiddlewareFunc
		size        int
		chunked     bool
		expected    int
		maxRead     int
	}{
		{
			description: "accepts a body under the default limit",
			size:        mib / 2,
			expected:    http.StatusOK,
		},
		{
			description: "refuses a declared length over the default limit without reading it",
			size:        mib + 1,
			expected:    http.StatusRequestEntityTooLarge,
		},
		{
			description: "refuses a chunked body over the default limit",
			size:        2 * mib,
			chunked:     true,
			expected:    http.StatusRequestEntityTooLarge,
			maxRead:     mib + 1,
		},
		{
			description: "accepts a body under a limit the route raised",
			route:       []echo.MiddlewareFunc{routesmiddleware.BodyLimit(2 * mib)},
			size:        mib + mib/2,
			expected:    http.StatusOK,
		},
		{
			description: "refuses a chunked body over a limit the route raised",
			route:       []echo.MiddlewareFunc{routesmiddleware.BodyLimit(2 * mib)},
			size:        3 * mib,
			chunked:     true,
			expected:    http.StatusRequestEntityTooLarge,
			maxRead:     2*mib + 1,
		},
		{
			description: "accepts a body on a route whose limit is the largest int64",
			route:       []echo.MiddlewareFunc{routesmiddleware.BodyLimit(math.MaxInt64)},
			size:        mib / 2,
			expected:    http.StatusOK,
		},
		{
			description: "accepts any body on a route that removed the limit",
			route:       []echo.MiddlewareFunc{routesmiddleware.BodyLimit(0)},
			size:        3 * mib,
			chunked:     true,
			expected:    http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			router := bodyLimitProbeRouter(t, tc.route...)

			body := &countingReader{reader: bytes.NewReader(jsonBodyOfSize(tc.size))}

			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, bodyLimitProbe, body)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.ContentLength = int64(tc.size)
			if tc.chunked {
				req.ContentLength = -1
			}

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected, rec.Code, rec.Body.String())

			switch {
			case tc.expected != http.StatusRequestEntityTooLarge:
				assert.Equal(t, tc.size, body.read)
			case !tc.chunked:
				assert.Zero(t, body.read)
			default:
				assert.LessOrEqual(t, body.read, tc.maxRead)
			}
		})
	}
}

func TestRouterRefusesAnOversizedBodyTheClientWaitsToSend(t *testing.T) {
	router := bodyLimitProbeRouter(t)

	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)

	conn, err := new(net.Dialer).DialContext(t.Context(), "tcp", srv.Listener.Addr().String())
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() }) //nolint:errcheck // the test is over, so a failed close changes nothing

	_, err = conn.Write([]byte("POST " + bodyLimitProbe + " HTTP/1.1\r\nHost: shellhub\r\n" +
		"Content-Type: application/json\r\nContent-Length: 2097152\r\nExpect: 100-continue\r\n\r\n"))
	require.NoError(t, err)

	require.NoError(t, conn.SetReadDeadline(time.Now().Add(5*time.Second))) //nolint:forbidigo // a socket deadline, which a clock.Clock cannot drive

	resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
	require.NoError(t, err)
	t.Cleanup(func() { resp.Body.Close() }) //nolint:errcheck // the test is over, so a failed close changes nothing

	assert.Equal(t, http.StatusRequestEntityTooLarge, resp.StatusCode)
}
