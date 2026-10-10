package main

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"maps"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
	"golang.org/x/net/websocket"
)

const (
	forgedAddress   = "203.0.113.9"
	declaredAddress = "203.0.113.7"

	mcpInitialize = `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"e2e","version":"1"}}}`
)

var withoutRedirects = &http.Client{
	CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
}

type fetched struct {
	status int
	header http.Header
	body   string
}

func fetch(t *testing.T, client *http.Client, target, host string, header http.Header) fetched {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, target, nil)
	require.NoError(t, err)

	if host != "" {
		req.Host = host
	}

	maps.Copy(req.Header, header)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return fetched{status: resp.StatusCode, header: resp.Header, body: string(body)}
}

func gatewayHost(t *testing.T, compose *environment.DockerCompose) string {
	t.Helper()

	base, err := url.Parse(compose.BaseURL())
	require.NoError(t, err)

	return base.Host
}

func probePath() string {
	return "/api/e2e/" + uuid.Generate()
}

func rawExchange(t *testing.T, address, preamble, path string) string {
	t.Helper()

	conn, err := (&net.Dialer{}).DialContext(t.Context(), "tcp", address)
	require.NoError(t, err)

	defer func() { _ = conn.Close() }()

	require.NoError(t, conn.SetDeadline(clock.Now().Add(30*time.Second)))

	_, err = io.WriteString(conn, preamble+"GET "+path+" HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n")
	require.NoError(t, err)

	status, err := bufio.NewReader(conn).ReadString('\n')
	require.NoError(t, err)

	return strings.TrimSpace(status)
}

func proxyPreamble(source string) string {
	return "PROXY TCP4 " + source + " 127.0.0.1 40000 80\r\n"
}

func execInGateway(t *testing.T, compose *environment.DockerCompose, script string) string {
	t.Helper()

	_, output, err := compose.Service(environment.ServiceGateway).Exec(t.Context(), []string{"sh", "-c", script}, tcexec.Multiplexed())
	require.NoError(t, err)

	out, err := io.ReadAll(output)
	require.NoError(t, err)

	return string(out)
}

var logField = regexp.MustCompile(`(?:^|\s)([a-z_]+)=("(?:[^"\\]|\\.)*"|\S*)`)

func logFields(line string) map[string]string {
	fields := make(map[string]string)
	for _, match := range logField.FindAllStringSubmatch(line, -1) {
		fields[match[1]] = strings.Trim(match[2], `"`)
	}

	return fields
}

func requestLogged(t *testing.T, compose *environment.DockerCompose, mark int, path string) map[string]string {
	t.Helper()

	return logFields(compose.AwaitServerLogLine(t, mark, "uri="+path+" "))
}

// TestGateway drives the community gateway as it ships, through the HTTP port it publishes and,
// where the test needs to know the address a request comes from, from inside its own container.
// The API reports the address and request ID the gateway forwards in its request log, which is
// where these cases read them.
func TestGateway(t *testing.T) {
	ctx := context.Background()

	compose := newSSHEnvironment(t, ctx, "legacy")
	host := gatewayHost(t, compose)
	port := strings.TrimPrefix(host, "localhost:")

	console := fetch(t, withoutRedirects, compose.BaseURL()+"/", "", nil)
	require.Equal(t, http.StatusOK, console.status)
	require.Contains(t, strings.ToLower(console.body), "<!doctype html>")

	t.Run("the MCP endpoint reaches the API", func(t *testing.T) {
		key := compose.CreateAPIKey(t, &requests.CreateAPIKey{Name: "mcp", ExpiresAt: -1, OptRole: authorizer.RoleObserver})

		mcp := func() *http.Request {
			req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, compose.BaseURL()+"/mcp", strings.NewReader(mcpInitialize))
			require.NoError(t, err)

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json, text/event-stream")

			return req
		}

		anonymous, err := http.DefaultClient.Do(mcp())
		require.NoError(t, err)
		_ = anonymous.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, anonymous.StatusCode)

		authenticated := mcp()
		authenticated.Header.Set("X-API-Key", key.Key)

		resp, err := http.DefaultClient.Do(authenticated)
		require.NoError(t, err)

		defer func() { _ = resp.Body.Close() }()

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode, string(body))

		initialized := struct {
			Result struct {
				ServerInfo struct {
					Name string `json:"name"`
				} `json:"serverInfo"`
			} `json:"result"`
		}{}
		require.NoError(t, json.Unmarshal(body, &initialized), string(body))
		assert.Equal(t, "shellhub", initialized.Result.ServerInfo.Name)
	})

	t.Run("the admin API is not routed in the community edition", func(t *testing.T) {
		admin := fetch(t, withoutRedirects, compose.BaseURL()+"/admin/api/users", "", nil)

		assert.Equal(t, http.StatusOK, admin.status)
		assert.Equal(t, strings.TrimSpace(console.body), strings.TrimSpace(admin.body), "the console answered, not the API")
	})

	t.Run("an unknown path serves the console", func(t *testing.T) {
		deepLink := fetch(t, withoutRedirects, compose.BaseURL()+"/devices/e2e/"+uuid.Generate(), "", nil)

		assert.Equal(t, http.StatusOK, deepLink.status)
		assert.Contains(t, deepLink.header.Get("Content-Type"), "text/html")
		assert.Equal(t, strings.TrimSpace(console.body), strings.TrimSpace(deepLink.body))
	})

	t.Run("the health check endpoint reaches the API", func(t *testing.T) {
		mark := compose.ServerLogMark(t)
		probe := uuid.Generate()

		health := fetch(t, withoutRedirects, compose.BaseURL()+"/healthcheck?probe="+probe, "", nil)

		assert.Equal(t, http.StatusOK, health.status)
		assert.Empty(t, health.body)
		compose.AwaitServerLogLine(t, mark, "status=200", probe)
	})

	t.Run("the gateway container reports itself healthy", func(t *testing.T) {
		inspected, err := compose.Service(environment.ServiceGateway).Inspect(t.Context())
		require.NoError(t, err)
		require.NotNil(t, inspected.State.Health)
		assert.Equal(t, "healthy", string(inspected.State.Health.Status))

		assert.Equal(t, "ok", execInGateway(t, compose, "wget -q -O- http://healthcheck.internal/healthz"))
	})

	t.Run("the kickstart script is served uncached and names the gateway's address", func(t *testing.T) {
		kickstart := fetch(t, withoutRedirects, compose.BaseURL()+"/kickstart.sh", "", nil)
		install := fetch(t, withoutRedirects, compose.BaseURL()+"/install.sh", "", nil)

		require.Equal(t, http.StatusOK, kickstart.status, kickstart.body)
		assert.Equal(t, "no-store", kickstart.header.Get("Cache-Control"))
		assert.True(t, strings.HasPrefix(kickstart.body, "#!"), "the kickstart script is a script")
		assert.Contains(t, kickstart.body, "SERVER_ADDRESS='http://"+host+"'")
		assert.Equal(t, install.body, kickstart.body, "kickstart.sh is the install script under its other name")
	})

	t.Run("a request for a name the gateway does not serve is answered 404", func(t *testing.T) {
		unknown := fetch(t, withoutRedirects, compose.BaseURL()+"/info", "unknown.example", nil)

		assert.Equal(t, http.StatusNotFound, unknown.status)
		assert.Empty(t, unknown.body)
	})

	t.Run("plain HTTP is served without a redirect when TLS is off", func(t *testing.T) {
		plain := fetch(t, withoutRedirects, compose.BaseURL()+"/info", "", nil)

		assert.Equal(t, http.StatusOK, plain.status)
		assert.Empty(t, plain.header.Get("Location"))
	})

	t.Run("a PROXY preamble is not read when the PROXY protocol is off", func(t *testing.T) {
		mark := compose.ServerLogMark(t)
		path := probePath()

		assert.Equal(t, "HTTP/1.1 400 Bad Request", rawExchange(t, host, proxyPreamble(declaredAddress), path))

		control := probePath()
		rawExchange(t, host, "", control)
		requestLogged(t, compose, mark, control)
		assert.NotContains(t, compose.ServerLogSince(t, mark), path, "the request after the preamble reached the API")
	})

	t.Run("X-Real-IP carries the address the client connected from", func(t *testing.T) {
		mark := compose.ServerLogMark(t)
		path := probePath()

		execInGateway(t, compose, "wget -q -O- --header 'X-Real-IP: "+forgedAddress+"' http://localhost"+path)

		remote := net.ParseIP(requestLogged(t, compose, mark, path)["remote_ip"])
		require.NotNil(t, remote)
		assert.True(t, remote.IsLoopback(), "the API attributed the request to %s, not to the loopback client", remote)
	})

	t.Run("X-Request-ID is generated for every request", func(t *testing.T) {
		mark := compose.ServerLogMark(t)
		first, second := probePath(), probePath()
		forged := http.Header{"X-Request-Id": {"forged-request-id"}}

		fetch(t, withoutRedirects, compose.BaseURL()+first, "", forged)
		fetch(t, withoutRedirects, compose.BaseURL()+second, "", forged)

		firstID := requestLogged(t, compose, mark, first)["id"]
		secondID := requestLogged(t, compose, mark, second)["id"]

		assert.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, firstID)
		assert.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, secondID)
		assert.NotEqual(t, firstID, secondID)
	})

	t.Run("X-Forwarded-Port names the port the client addressed", func(t *testing.T) {
		apiEndpoint := func(host string) string {
			info := fetch(t, withoutRedirects, compose.BaseURL()+"/info", host, nil)
			require.Equal(t, http.StatusOK, info.status, info.body)

			decoded := struct {
				Endpoints struct {
					API string `json:"api"`
				} `json:"endpoints"`
			}{}
			require.NoError(t, json.Unmarshal([]byte(info.body), &decoded))

			return decoded.Endpoints.API
		}

		assert.Equal(t, "localhost:"+port, apiEndpoint(""), "the port in Host is the one the client addressed")
		assert.Equal(t, "localhost:80", apiEndpoint("localhost"), "a Host without a port is answered with the port the gateway listens on")
	})

	t.Run("an idle web terminal stays open", func(t *testing.T) {
		_, device := startAcceptedAgent(t, ctx, compose)
		conn, _ := dialWebTerminal(t, ctx, compose, device)

		idle := 75 * time.Second
		time.Sleep(idle)

		require.NoError(t, websocket.JSON.Send(conn, map[string]any{"kind": 1, "data": "echo $((40+2))gateway\n"}), "the terminal closed while idle")
		require.NoError(t, conn.SetReadDeadline(clock.Now().Add(30*time.Second)))

		var output strings.Builder
		for !strings.Contains(output.String(), "42gateway") {
			var frame []byte
			require.NoError(t, websocket.Message.Receive(conn, &frame), "the terminal closed after %s idle, having written %q", idle, output.String())
			output.Write(frame)
		}
	})
}
