package main

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
)

// TestGatewayProxyProtocol runs the gateway behind the PROXY protocol, trusting only the loopback
// addresses. A client inside the gateway's container connects from loopback, so it is trusted; the
// test process reaches the published port through Docker's forwarding, so it is not.
func TestGatewayProxyProtocol(t *testing.T) {
	ctx := context.Background()

	compose := environment.New(t, run).
		WithEnv("SHELLHUB_PROXY", "true").
		WithEnv("SHELLHUB_PROXY_TRUSTED_IPS", "127.0.0.1/32 ::1/128").
		Up(ctx)
	t.Cleanup(compose.Down)

	host := gatewayHost(t, compose)

	t.Run("a trusted peer declares the client's address", func(t *testing.T) {
		mark := compose.ServerLogMark(t)
		path := probePath()

		response := execInGateway(t, compose, `{ printf '`+proxyPreamble(declaredAddress)+`GET `+path+` HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n'; sleep 5; } | nc 127.0.0.1 80`)
		assert.Contains(t, response, "HTTP/1.1 401 ", "the API did not answer the request after the preamble")

		assert.Equal(t, declaredAddress, requestLogged(t, compose, mark, path)["remote_ip"])
	})

	t.Run("an untrusted peer cannot declare the client's address", func(t *testing.T) {
		mark := compose.ServerLogMark(t)
		control, declared := probePath(), probePath()

		rawExchange(t, host, "", control)
		peer := requestLogged(t, compose, mark, control)["remote_ip"]

		assert.Equal(t, "HTTP/1.1 401 Unauthorized", rawExchange(t, host, proxyPreamble(declaredAddress), declared), "the gateway drops the preamble and serves the request")

		assert.Equal(t, peer, requestLogged(t, compose, mark, declared)["remote_ip"], "the API attributed the request to the declared %s", declaredAddress)
	})
}
