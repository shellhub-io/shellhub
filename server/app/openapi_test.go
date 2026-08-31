package app

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/server/api/routes"
	"github.com/shellhub-io/shellhub/server/api/routes/middleware"
	sshhttp "github.com/shellhub-io/shellhub/server/ssh/http"
	"github.com/shellhub-io/shellhub/server/ssh/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	xwebsocket "golang.org/x/net/websocket"
)

const controlUpgradePath = "/api/openapi-validation-control"

const openAPITestSchema = `{
  "openapi": "3.0.0",
  "info": {"title": "test", "version": "1.0.0"},
  "paths": {}
}`

// TestOpenAPIValidationAllowsWebSocketUpgrades drives a real handshake against every route
// that hijacks the connection, with the validator middleware installed exactly as production
// installs it.
//
// The control route is what keeps this test honest. The middleware silently becomes a no-op
// when the validator fails to initialize, which would let every assertion below pass without
// testing anything. An unskipped upgrade route must therefore still fail.
func TestOpenAPIValidationAllowsWebSocketUpgrades(t *testing.T) {
	schema := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(openAPITestSchema)) //nolint:errcheck
	}))
	t.Cleanup(schema.Close)

	schemaURL, err := url.Parse(schema.URL + "/openapi.json")
	require.NoError(t, err)

	upgrader := &websocket.Upgrader{} //nolint:exhaustruct

	upgrade := func(c *echo.Context) error {
		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}

		return conn.Close()
	}

	e := echo.New()
	e.Use(middleware.OpenAPIValidator(&middleware.OpenAPIValidatorConfig{ //nolint:exhaustruct
		SchemaPath: schemaURL,
		Skipper:    openAPIValidationSkipper,
	}))

	for _, path := range []string{
		sshhttp.HandleConnectionV1Path,
		sshhttp.HandleConnectionV2Path,
		sshhttp.HandleRevdialPath,
		controlUpgradePath,
	} {
		e.GET(path, upgrade)
	}

	e.GET(web.WebsocketSSHBridgeRoute, echo.WrapHandler(xwebsocket.Handler(func(conn *xwebsocket.Conn) {
		conn.Close() //nolint:errcheck
	})))

	srv := httptest.NewServer(e)
	t.Cleanup(srv.Close)

	headers := http.Header{"Origin": []string{srv.URL}}

	dial := func(t *testing.T, path string) (*http.Response, error) {
		t.Helper()

		conn, res, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http")+path, headers)
		if conn != nil {
			conn.Close() //nolint:errcheck
		}

		return res, err
	}

	for _, path := range []string{
		sshhttp.HandleConnectionV1Path,
		sshhttp.HandleConnectionV2Path,
		sshhttp.HandleRevdialPath,
		web.WebsocketSSHBridgeRoute,
	} {
		t.Run("upgrade succeeds on "+path, func(t *testing.T) {
			res, err := dial(t, path) //nolint:bodyclose // the handshake response body need not be closed.
			require.NoError(t, err)
			assert.Equal(t, http.StatusSwitchingProtocols, res.StatusCode)
		})
	}

	t.Run("an unskipped route still cannot upgrade", func(t *testing.T) {
		_, err := dial(t, controlUpgradePath) //nolint:bodyclose // the handshake response body need not be closed.
		require.Error(t, err, "the validator did not initialize, so this test proves nothing")
	})
}

func TestOpenAPIValidationSkipper(t *testing.T) {
	cases := []struct {
		path string
		skip bool
	}{
		{sshhttp.HandleConnectionV1Path, true},
		{sshhttp.HandleConnectionV2Path, true},
		{sshhttp.HandleRevdialPath, true},
		{web.WebsocketSSHBridgeRoute, true},
		{"/metrics", true},
		{"/internal/auth", true},
		{routes.InternalMetricsURL, true},
		{web.WebSessionRoute, false},
		{"/api/devices", false},
		{"/api/namespaces", false},
	}

	e := echo.New()

	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, tc.path, nil)
			ctx := e.NewContext(req, httptest.NewRecorder())

			assert.Equal(t, tc.skip, openAPIValidationSkipper(ctx))
		})
	}
}
