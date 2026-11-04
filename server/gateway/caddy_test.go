package gateway

import (
	"encoding/json"
	"net/http"
	"slices"
	"testing"

	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func configurations() map[string]*GatewayConfig {
	base := func() *GatewayConfig {
		return &GatewayConfig{ //nolint:exhaustruct
			Domain:          "shellhub.example",
			ProxyTrustedIPs: "0.0.0.0/0 ::/0",
		}
	}

	with := func(mutate func(*GatewayConfig)) *GatewayConfig {
		cfg := base()
		mutate(cfg)

		return cfg
	}

	return map[string]*GatewayConfig{
		"community":       base(),
		"enterprise":      with(func(c *GatewayConfig) { c.EnableEnterprise = true }),
		"web endpoints":   with(func(c *GatewayConfig) { c.EnableEnterprise = true; c.WebEndpoints = true }),
		"proxy protocol":  with(func(c *GatewayConfig) { c.EnableProxyProtocol = true }),
		"auto ssl":        with(func(c *GatewayConfig) { c.EnableAutoSSL = true }),
		"auto ssl + wild": with(func(c *GatewayConfig) { c.EnableAutoSSL = true; c.WebEndpoints = true; c.EnableEnterprise = true }),
		"dns cloudflare": with(func(c *GatewayConfig) {
			c.EnableAutoSSL = true
			c.WebEndpoints = true
			c.WebEndpointsDNSProvider = CloudflareDNSProvider
		}),
		"dns acmedns": with(func(c *GatewayConfig) {
			c.EnableAutoSSL = true
			c.WebEndpoints = true
			c.WebEndpointsDNSProvider = AcmeDNSProvider
			c.WebEndpointsACMEDNSURL = "https://auth.acme-dns.example"
			c.WebEndpointsACMEDNSUsername = "c36f50e8-4632-44f0-83fe-e070fef28a10"
			c.WebEndpointsACMEDNSSubdomain = "8e5700ea-a4bf-41c7-8a77-e990661dcb6a"
		}),
		"development":    with(func(c *GatewayConfig) { c.Env = "development"; c.EnableEnterprise = true }),
		"dev community":  with(func(c *GatewayConfig) { c.Env = "development" }),
		"no access logs": with(func(c *GatewayConfig) { c.EnableAccessLogs = false }),
		"supplied certificate": with(func(c *GatewayConfig) {
			c.EnableAutoSSL = true
			c.TLSCertFile = "/etc/shellhub/certs/fullchain.pem"
			c.TLSKeyFile = "/etc/shellhub/certs/privkey.pem"
		}),
	}
}

func TestCaddyfileAdapts(t *testing.T) {
	for description, cfg := range configurations() {
		t.Run(description, func(t *testing.T) {
			rendered, err := Caddyfile(cfg)
			require.NoError(t, err)

			_, warnings, err := caddyconfig.GetAdapter("caddyfile").Adapt(rendered, nil)
			require.NoError(t, err, "the generated Caddyfile does not adapt:\n%s", rendered)
			assert.Empty(t, warnings, "the generated Caddyfile adapts with warnings:\n%s", rendered)
		})
	}
}

func TestCaddyfileTLSFollowsTheFlagAlone(t *testing.T) {
	cfg := configurations()["development"]
	cfg.EnableAutoSSL = true

	rendered, err := Caddyfile(cfg)
	require.NoError(t, err)

	assert.Contains(t, string(rendered), "https://shellhub.example")
	assert.NotContains(t, string(rendered), "auto_https off")
}

func TestTheFallbackIsTheLastRouteOnPortEighty(t *testing.T) {
	for description, cfg := range configurations() {
		t.Run(description, func(t *testing.T) {
			rendered, err := Caddyfile(cfg)
			require.NoError(t, err)

			adapted, _, err := caddyconfig.GetAdapter("caddyfile").Adapt(rendered, nil)
			require.NoError(t, err)

			for name, server := range httpServers[httpServer](t, adapted) {
				if !slices.Contains(server.Listen, ":80") {
					continue
				}

				require.NotEmpty(t, server.Routes, "%s serves port 80 with no route at all", name)

				for i, route := range server.Routes[:len(server.Routes)-1] {
					assert.NotEmpty(t, hosts(route),
						"route %d of %s matches every Host, so the fallback is not the only one that does", i, name)
				}

				assert.Empty(t, hosts(server.Routes[len(server.Routes)-1]),
					"the fallback must be the last route of %s", name)
			}
		})
	}
}

type httpRoute struct {
	Match []struct {
		Host []string `json:"host"`
	} `json:"match"`
}

type httpServer struct {
	Listen []string    `json:"listen"`
	Routes []httpRoute `json:"routes"`
}

func httpServers[S any](t *testing.T, adapted []byte) map[string]S {
	t.Helper()

	var config struct {
		Apps struct {
			HTTP struct {
				Servers map[string]S `json:"servers"`
			} `json:"http"`
		} `json:"apps"`
	}

	require.NoError(t, json.Unmarshal(adapted, &config))

	return config.Apps.HTTP.Servers
}

func hosts(route httpRoute) []string {
	var names []string

	for _, match := range route.Match {
		names = append(names, match.Host...)
	}

	return names
}

func TestCaddyfileProvesTheWildcardWithTheACMEDNSAccount(t *testing.T) {
	cfg := configurations()["dns acmedns"]

	rendered, err := Caddyfile(cfg)
	require.NoError(t, err)

	adapted, _, err := caddyconfig.GetAdapter("caddyfile").Adapt(rendered, nil)
	require.NoError(t, err)

	var config any
	require.NoError(t, json.Unmarshal(adapted, &config))

	providers := collect(config, func(node map[string]any) bool { return node["name"] == "acmedns" })
	require.Len(t, providers, 1, "the configuration names one acme-dns provider:\n%s", adapted)

	assert.Equal(t, cfg.WebEndpointsACMEDNSURL, providers[0]["server_url"])
	assert.Equal(t, cfg.WebEndpointsACMEDNSUsername, providers[0]["username"])
	assert.Equal(t, cfg.WebEndpointsACMEDNSSubdomain, providers[0]["subdomain"])
	assert.Equal(t, "{env.SHELLHUB_WEB_ENDPOINTS_ACME_DNS_PASSWORD}", providers[0]["password"],
		"the password stays out of the rendered configuration; the module resolves it")
}

func TestTrustedProxiesSurvivesAnEmptyVariable(t *testing.T) {
	cfg := configurations()["proxy protocol"]
	cfg.ProxyTrustedIPs = ""
	cfg.applyDefaults()

	rendered, err := Caddyfile(cfg)
	require.NoError(t, err)

	assert.Contains(t, string(rendered), "trusted_proxies static "+defaultProxyTrustedIPs)
	assert.NotContains(t, string(rendered), "trusted_proxies static\n")
}

func TestCaddyfileTrustsOnlyTheConfiguredProxies(t *testing.T) {
	cfg := configurations()["proxy protocol"]
	cfg.ProxyTrustedIPs = "10.0.0.0/8"

	rendered, err := Caddyfile(cfg)
	require.NoError(t, err)

	assert.Contains(t, string(rendered), "trusted_proxies static 10.0.0.0/8")
	assert.Contains(t, string(rendered), "allow 10.0.0.0/8")

	without, err := Caddyfile(configurations()["community"])
	require.NoError(t, err)

	assert.NotContains(t, string(without), "proxy_protocol",
		"the listener wrapper has no business being there when the feature is off")
}

func collect(node any, matches func(map[string]any) bool) []map[string]any {
	var found []map[string]any

	switch value := node.(type) {
	case []map[string]any:
		for _, item := range value {
			found = append(found, collect(item, matches)...)
		}
	case []any:
		for _, item := range value {
			found = append(found, collect(item, matches)...)
		}
	case map[string]any:
		if matches(value) {
			found = append(found, value)
		}

		for _, child := range value {
			found = append(found, collect(child, matches)...)
		}
	}

	return found
}

func TestAnInfoQueryEscapedIntoThePathReachesTheAPI(t *testing.T) {
	requested := make(chan string, 1)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requested <- r.URL.RequestURI()
		w.WriteHeader(http.StatusOK)
	})

	cfg := configurations()["community"]
	address := startServing(t, cfg, handler)

	for _, uri := range []string{"/info%3Fagent_version=v0.8.2", "/info?agent_version=v0.8.2"} {
		t.Run(uri, func(t *testing.T) {
			req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://"+address+uri, nil)
			require.NoError(t, err)
			req.Host = cfg.Domain

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			require.NoError(t, res.Body.Close())

			select {
			case got := <-requested:
				assert.Equal(t, "/info?agent_version=v0.8.2", got)
			default:
				t.Fatalf("%s did not reach the API, the gateway answered %d", uri, res.StatusCode)
			}
		})
	}
}
