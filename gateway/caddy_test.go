package main

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func configurations() map[string]*GatewayConfig {
	base := func() *GatewayConfig {
		return &GatewayConfig{ //nolint:exhaustruct
			Domain:          "shellhub.example",
			APIBackend:      "server:8080",
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

// TestCaddyfileAdapts is the guard the previous configuration never had: every
// shape the environment can produce is handed to the real adapter, which parses
// the file and resolves every directive against the modules compiled in. A typo,
// a directive that does not exist, a snippet used before it is defined -- all of
// it fails here rather than at a customer's boot.
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

// TestCaddyfileServesTheEditionsRoutes pins the two blocks that are not always
// there, because both are edition-gated and a mistake in either is invisible
// until someone runs that edition.
func TestCaddyfileServesTheEditionsRoutes(t *testing.T) {
	configs := configurations()

	community, err := Caddyfile(configs["community"])
	require.NoError(t, err)

	assert.NotContains(t, string(community), "/admin/api",
		"the admin API is enterprise-only")
	assert.NotContains(t, string(community), "*.shellhub.example",
		"web endpoints are not served unless the feature is on")

	endpoints, err := Caddyfile(configs["web endpoints"])
	require.NoError(t, err)

	assert.Contains(t, string(endpoints), "*.shellhub.example")
	assert.Contains(t, string(endpoints), "/admin/api")
}

// TestCaddyfileServesTheDevelopmentMailInbox pins the Mailpit vhost to the one
// stack that has a Mailpit to reach: the mail code lives in the cloud repository,
// so a community stack has nothing behind the hostname even in development.
func TestCaddyfileServesTheDevelopmentMailInbox(t *testing.T) {
	configs := configurations()

	development, err := Caddyfile(configs["development"])
	require.NoError(t, err)

	assert.Contains(t, string(development), "http://mail.shellhub.example")

	devCommunity, err := Caddyfile(configs["dev community"])
	require.NoError(t, err)

	assert.NotContains(t, string(devCommunity), "mail.shellhub.example",
		"the mail inbox is enterprise-only")

	enterprise, err := Caddyfile(configs["enterprise"])
	require.NoError(t, err)

	assert.NotContains(t, string(enterprise), "mail.shellhub.example",
		"the mail inbox is development-only")
}

func TestEveryUpstreamReceivesTheClientAddress(t *testing.T) {
	for description, cfg := range configurations() {
		t.Run(description, func(t *testing.T) {
			rendered, err := Caddyfile(cfg)
			require.NoError(t, err)
			adapted, _, err := caddyconfig.GetAdapter("caddyfile").Adapt(rendered, nil)
			require.NoError(t, err)
			var config any
			require.NoError(t, json.Unmarshal(adapted, &config))
			proxies := reverseProxies(config)
			require.NotEmpty(t, proxies)

			for _, proxy := range proxies {
				encoded, err := json.Marshal(proxy)
				require.NoError(t, err)
				var handler struct {
					Headers struct {
						Request struct {
							Set map[string][]string `json:"set"`
						} `json:"request"`
					} `json:"headers"`
				}
				require.NoError(t, json.Unmarshal(encoded, &handler))
				assert.Equal(t, []string{"{http.request.remote.host}"}, handler.Headers.Request.Set["X-Real-Ip"], "%s", encoded)
			}
		})
	}
}

func reverseProxies(node any) []map[string]any {
	var proxies []map[string]any
	switch value := node.(type) {
	case map[string]any:
		if value["handler"] == "reverse_proxy" {
			proxies = append(proxies, value)
		}
		for _, child := range value {
			proxies = append(proxies, reverseProxies(child)...)
		}
	case []any:
		for _, child := range value {
			proxies = append(proxies, reverseProxies(child)...)
		}
	}

	return proxies
}

// TestCaddyfileTLSFollowsTheFlagAlone pins that development is not a special
// case: the previous configuration refused TLS there, which meant the setting
// silently did nothing and the certificate path could not be rehearsed.
func TestCaddyfileTLSFollowsTheFlagAlone(t *testing.T) {
	cfg := configurations()["development"]
	cfg.EnableAutoSSL = true

	rendered, err := Caddyfile(cfg)
	require.NoError(t, err)

	assert.Contains(t, string(rendered), "https://shellhub.example")
	assert.NotContains(t, string(rendered), "auto_https off")
}

// TestCaddyfileServesTheSuppliedCertificate pins both halves of the conditional,
// because a mistake in either is silent: the directive missing means an ACME
// request nobody asked for, and the directive appearing when no certificate was
// supplied means a site that cannot serve at all.
func TestCaddyfileServesTheSuppliedCertificate(t *testing.T) {
	cfg := configurations()["supplied certificate"]

	rendered, err := Caddyfile(cfg)
	require.NoError(t, err)

	assert.Contains(t, string(rendered), "tls "+cfg.TLSCertFile+" "+cfg.TLSKeyFile)

	without, err := Caddyfile(configurations()["auto ssl"])
	require.NoError(t, err)

	assert.NotContains(t, string(without), "tls ",
		"automatic issuance is the default and must stay untouched when no certificate is supplied")
}

// TestHealthcheckAnswersOnEveryConfiguration asserts every shape rather than one,
// because the bug it replaces was a healthcheck that passed under one setting and
// failed under another -- and passed by accident, on Caddy's reply to a name it
// did not serve, so nobody had reason to look at it.
func TestHealthcheckAnswersOnEveryConfiguration(t *testing.T) {
	for description, cfg := range configurations() {
		t.Run(description, func(t *testing.T) {
			rendered, err := Caddyfile(cfg)
			require.NoError(t, err)

			assert.Contains(t, string(rendered), "http://healthcheck.internal",
				"the healthcheck site must not depend on how TLS is configured")
		})
	}
}

// TestCaddyfileRefusesAHostItDoesNotServe pins the fallback site. Without it a
// request arriving with any other Host reaches Caddy's own default, which
// answers 200 with an empty body -- indistinguishable from a healthy ShellHub
// until whatever asked tries to read the reply.
func TestCaddyfileRefusesAHostItDoesNotServe(t *testing.T) {
	for description, cfg := range configurations() {
		t.Run(description, func(t *testing.T) {
			rendered, err := Caddyfile(cfg)
			require.NoError(t, err)

			assert.Contains(t, string(rendered), "http:// {",
				"a Host this proxy does not serve must not fall through to Caddy's default")
		})
	}
}

// TestTheFallbackIsTheLastRouteOnPortEighty is the other half of that fallback:
// it matches every Host, so anything ordered after it is unreachable. Automatic
// HTTPS appends its redirect for a name it holds no certificate for -- a supplied
// certificate is exactly that name -- and the site would answer 404 on port 80
// instead of redirecting.
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

// TestCaddyfileRedirectsItsOwnNameToHTTPS pins the redirect the fallback would
// otherwise shadow. Caddy writes one per name it manages a certificate for and a
// single blanket one for the rest, and the blanket one is what a deployment
// serving a supplied certificate has.
func TestCaddyfileRedirectsItsOwnNameToHTTPS(t *testing.T) {
	rendered, err := Caddyfile(configurations()["supplied certificate"])
	require.NoError(t, err)

	assert.Contains(t, string(rendered), "http://shellhub.example {")
	assert.Contains(t, string(rendered), "redir https://{host}{uri} 308")

	withoutTLS, err := Caddyfile(configurations()["community"])
	require.NoError(t, err)

	assert.NotContains(t, string(withoutTLS), "redir ",
		"there is nothing to redirect to when the site is served over plain HTTP")
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

// TestCaddyfileProvesTheWildcardWithTheACMEDNSAccount guards the account the
// acme-dns provider writes the challenge through. The module resolves no
// placeholder but the password's, so any other field left as one reaches the
// acme-dns server as the placeholder's own text.
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

// TestCaddyfileStoresCertificatesOnTheVolume guards a mistake that is invisible
// until a container is recreated and the CA is asked for everything again:
// Caddy's default storage is a directory under $HOME, which no volume covers.
func TestCaddyfileStoresCertificatesOnTheVolume(t *testing.T) {
	rendered, err := Caddyfile(configurations()["auto ssl"])
	require.NoError(t, err)

	assert.Contains(t, string(rendered), "storage file_system /data")
}

// TestTrustedProxiesSurvivesAnEmptyVariable covers the upgrade path, not the
// fresh install: compose passes SHELLHUB_PROXY_TRUSTED_IPS through, so an .env
// from before this key existed sends it set-but-empty, and an env default only
// fires when a variable is absent. Empty renders `trusted_proxies static` with
// no arguments, which adapts without complaint and trusts nobody -- so the
// balancer's own address ends up keying the login lockout.
func TestTrustedProxiesSurvivesAnEmptyVariable(t *testing.T) {
	cfg := configurations()["proxy protocol"]
	cfg.ProxyTrustedIPs = ""
	cfg.applyDefaults()

	rendered, err := Caddyfile(cfg)
	require.NoError(t, err)

	assert.Contains(t, string(rendered), "trusted_proxies static "+defaultProxyTrustedIPs)
	assert.NotContains(t, string(rendered), "trusted_proxies static\n")
}

// TestCaddyfileTrustsOnlyTheConfiguredProxies guards the value the login lockout
// and the GeoIP rules are keyed on: under the PROXY protocol, a peer that is
// allowed to send the preamble is a peer that chooses the client's address.
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

func TestCaddyfileLeavesAnIdleWebTerminalOpen(t *testing.T) {
	for description, cfg := range configurations() {
		t.Run(description, func(t *testing.T) {
			rendered, err := Caddyfile(cfg)
			require.NoError(t, err)

			adapted, _, err := caddyconfig.GetAdapter("caddyfile").Adapt(rendered, nil)
			require.NoError(t, err)

			proxied := false

			for name, server := range httpServers[map[string]any](t, adapted) {
				routes := collect(server["routes"], func(node map[string]any) bool {
					matchers, _ := node["match"].([]any)

					return slices.ContainsFunc(matchers, func(matcher any) bool {
						paths, _ := matcher.(map[string]any)["path"].([]any)

						return slices.Contains(paths, any("/ws*"))
					})
				})
				if len(routes) == 0 {
					continue
				}

				for _, timeout := range []string{"read_timeout", "write_timeout"} {
					assert.Nil(t, server[timeout], "%s serves /ws with a %s", name, timeout)
				}

				proxies := collect(routes, func(node map[string]any) bool { return node["handler"] == "reverse_proxy" })
				for _, proxy := range proxies {
					proxied = true

					assert.Nil(t, proxy["stream_timeout"], "the /ws proxy of %s closes upgraded streams", name)

					transport, _ := proxy["transport"].(map[string]any)
					for _, timeout := range []string{"read_timeout", "write_timeout", "response_header_timeout"} {
						assert.Nil(t, transport[timeout], "the /ws proxy of %s has a transport %s", name, timeout)
					}
				}
			}

			assert.True(t, proxied, "no server proxies /ws")
		})
	}
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
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requested <- r.URL.RequestURI()
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(api.Close)

	cfg := configurations()["community"]
	cfg.APIBackend = strings.TrimPrefix(api.URL, "http://")

	rendered, err := Caddyfile(cfg)
	require.NoError(t, err)

	adapted, _, err := caddyconfig.GetAdapter("caddyfile").Adapt(rendered, nil)
	require.NoError(t, err)

	free, err := new(net.ListenConfig).Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	address := free.Addr().String()
	require.NoError(t, free.Close())

	adapted = bytes.ReplaceAll(adapted, []byte(`":80"`), []byte(`"`+address+`"`))
	adapted = bytes.ReplaceAll(adapted, []byte(`"/data"`), []byte(`"`+t.TempDir()+`"`))

	require.NoError(t, caddy.Load(adapted, true))
	t.Cleanup(func() { _ = caddy.Stop() })

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
