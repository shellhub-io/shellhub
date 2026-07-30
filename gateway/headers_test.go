package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// forwardedHeaders is the set the main server block declares once, for every
// upstream. It is the whole contract with the API: identity is resolved
// in-process, so what the proxy still owes the backend is the transport context
// -- who connected, over what, and under which name.
var forwardedHeaders = []string{
	"Host",
	"Upgrade",
	"Connection",
	"X-Forwarded-Host",
	"X-Forwarded-Port",
	"X-Forwarded-Proto",
	"X-Request-ID",
	"X-Real-IP",
}

var (
	serverBlock  = regexp.MustCompile(`(?m)^server\s*\{`)
	locationOpen = regexp.MustCompile(`(?m)^\s*location\s+(.+?)\s*\{`)
	setHeader    = regexp.MustCompile(`(?m)^\s*proxy_set_header\s+(\S+)`)
)

// renderMainServer generates the real templates under cfg and returns the body
// of the main server block, plus each location inside it.
func renderMainServer(t *testing.T, cfg *GatewayConfig) (string, map[string]string) {
	t.Helper()

	root := t.TempDir()
	nc := &NginxController{rootDir: root, templatesDir: "nginx", gatewayConfig: cfg}
	nc.generateConfigs()

	generated, err := os.ReadFile(filepath.Join(root, "conf.d", "shellhub.conf")) //nolint:gosec // the path is the test's own temp dir.
	require.NoError(t, err)

	body, ok := firstBlock(string(generated), serverBlock)
	require.True(t, ok, "no server block was generated")

	locations := make(map[string]string)
	for _, match := range locationOpen.FindAllStringSubmatchIndex(body, -1) {
		name := strings.TrimSpace(body[match[2]:match[3]])
		inner, ok := blockAt(body, match[1]-1)
		require.True(t, ok, "unbalanced braces in location %q", name)

		locations[name] = inner
	}

	return body, locations
}

func firstBlock(text string, opener *regexp.Regexp) (string, bool) {
	loc := opener.FindStringIndex(text)
	if loc == nil {
		return "", false
	}

	return blockAt(text, loc[1]-1)
}

// blockAt returns the contents between the brace at open and its match.
func blockAt(text string, open int) (string, bool) {
	depth := 0
	for i := open; i < len(text); i++ {
		switch text[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return text[open+1 : i], true
			}
		}
	}

	return "", false
}

// TestMainServerForwardsTransportHeaders pins the inheritance the whole
// configuration rests on. nginx replaces the inherited proxy_set_header list
// wholesale when a location declares one of its own, so a single
// `proxy_set_header X-Foo` added to a location below would silently drop every
// header in forwardedHeaders for that route -- including X-Real-IP, which the
// login lockout and the GeoIP lookup read.
func TestMainServerForwardsTransportHeaders(t *testing.T) {
	tests := []struct {
		description string
		config      GatewayConfig
	}{
		{description: "community", config: GatewayConfig{}},
		{description: "enterprise", config: GatewayConfig{EnableEnterprise: true}},
		{description: "cloud", config: GatewayConfig{EnableEnterprise: true, EnableCloud: true}},
		{description: "web endpoints", config: GatewayConfig{EnableEnterprise: true, WebEndpoints: true}},
		{description: "proxy protocol", config: GatewayConfig{EnableProxyProtocol: true}},
		{description: "development", config: GatewayConfig{Env: "development", EnableEnterprise: true}},
		{description: "auto ssl", config: GatewayConfig{EnableAutoSSL: true}},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			cfg := tc.config
			cfg.Domain = "localhost"
			cfg.APIBackend = "server:8080"
			cfg.Database = "postgres"

			body, locations := renderMainServer(t, &cfg)

			declared := make(map[string]struct{})
			for _, match := range setHeader.FindAllStringSubmatch(body, -1) {
				declared[match[1]] = struct{}{}
			}

			for _, header := range forwardedHeaders {
				assert.Contains(t, declared, header,
					"the main server block must forward %s to every upstream", header)
			}

			for name, inner := range locations {
				own := setHeader.FindAllStringSubmatch(inner, -1)
				assert.Empty(t, own,
					"location %q declares its own proxy_set_header, which drops the inherited list", name)
			}
		})
	}
}

// TestProxyProtocolTrustsBothAddressFamilies guards a failure that is invisible
// in a config review: with only ::/0 trusted, an IPv4 peer leaves $remote_addr
// as the load balancer's address, so both limit_req zones key every client on
// the balancer and the access log records the wrong address.
func TestProxyProtocolTrustsBothAddressFamilies(t *testing.T) {
	cfg := GatewayConfig{
		Domain:              "localhost",
		APIBackend:          "server:8080",
		Database:            "postgres",
		EnableProxyProtocol: true,
	}

	body, _ := renderMainServer(t, &cfg)

	assert.Contains(t, body, "set_real_ip_from 0.0.0.0/0;")
	assert.Contains(t, body, "set_real_ip_from ::/0;")
	assert.Contains(t, body, "real_ip_header proxy_protocol;")
}
