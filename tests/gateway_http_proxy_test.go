package main

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGatewayBehindAnHTTPProxy runs the gateway behind an HTTP reverse proxy: SHELLHUB_PROXY off,
// loopback listed in SHELLHUB_PROXY_TRUSTED_IPS. A client inside the gateway's container plays
// the proxy and names the client in X-Forwarded-For; the test process reaches the published port
// through Docker's forwarding, so it is not trusted.
func TestGatewayBehindAnHTTPProxy(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	compose := environment.New(t, run).
		WithEnv("SHELLHUB_PROXY_TRUSTED_IPS", "127.0.0.1/32 ::1/128").
		Up(ctx)
	t.Cleanup(compose.Down)

	throughTheProxy := func(t *testing.T, forwardedFor, path string, wgetArgs ...string) string {
		t.Helper()

		args := strings.Join(wgetArgs, " ")

		return execInServer(t, compose, "wget -S -q -O /dev/null --header 'X-Real-IP: "+forgedAddress+"' --header 'X-Forwarded-For: "+forwardedFor+"' "+args+" http://localhost"+path+" 2>&1 | grep -m1 'HTTP/'")
	}

	t.Run("the proxy names the client by the right-most hop it did not add", func(t *testing.T) {
		mark := compose.ServerLogMark(t)
		path := probePath()

		assert.Contains(t, throughTheProxy(t, forgedAddress+", "+declaredAddress+", 127.0.0.1", path), " 401 ")

		assert.Equal(t, declaredAddress, requestLogged(t, compose, mark, path)["remote_ip"])
	})

	t.Run("an untrusted peer cannot name the client", func(t *testing.T) {
		mark := compose.ServerLogMark(t)
		control, claimed := probePath(), probePath()

		fetch(t, withoutRedirects, compose.BaseURL()+control, "", nil)
		peer := requestLogged(t, compose, mark, control)["remote_ip"]

		fetch(t, withoutRedirects, compose.BaseURL()+claimed, "", http.Header{
			"X-Real-Ip":       {forgedAddress},
			"X-Forwarded-For": {declaredAddress},
		})

		assert.Equal(t, peer, requestLogged(t, compose, mark, claimed)["remote_ip"])
	})

	t.Run("one client's failed sign-ins lock out only that client", func(t *testing.T) {
		const (
			attacker = "198.51.100.66"
			victim   = "198.51.100.7"
		)

		compose.NewUser(t, "proxied", "proxied@ossystems.com.br", ShellHubPassword)

		login := func(client, password string) string {
			body := `{"username":"proxied","password":"` + password + `"}`

			return throughTheProxy(t, client, "/api/login", "--header 'Content-Type: application/json'", "--post-data '"+body+"'")
		}

		for _, status := range []string{" 401 ", " 401 ", " 429 "} {
			require.Contains(t, login(attacker, "wrong-password"), status)
		}

		require.Contains(t, login(attacker, ShellHubPassword), " 429 ", "the attacker's own address is not locked out")
		assert.Contains(t, login(victim, ShellHubPassword), " 200 ", "the attacker locked out another client behind the same proxy")
	})
}
