package gateway

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const endpointAddress = "0123456789abcdef0123456789abcdef"

type exchange struct {
	host   string
	target string

	served       bool
	upstream     string
	uri          string
	upstreamHost string
	status       int
}

func (e exchange) String() string {
	return e.host + e.target
}

func serverRecorder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Served-URI", r.URL.RequestURI())
	w.Header().Set("X-Seen-Forwarded-Port", r.Header.Get("X-Forwarded-Port"))
	w.Header().Set("X-Seen-Forwarded-Proto", r.Header.Get("X-Forwarded-Proto"))
	w.WriteHeader(http.StatusNoContent)
}

func taggedUpstreams(t *testing.T) map[string]string {
	t.Helper()

	upstreams := make(map[string]string, len(defaultUpstreams))

	for name := range defaultUpstreams {
		upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Upstream", name)
			w.Header().Set("X-Upstream-URI", r.URL.RequestURI())
			w.Header().Set("X-Upstream-Host", r.Host)
			w.Header().Set("X-Upstream-Real-IP", r.Header.Get("X-Real-IP"))
			w.Header().Set("X-Frame-Options", "SAMEORIGIN")
			w.Header().Set("Content-Security-Policy", "frame-ancestors 'self'")
			w.WriteHeader(http.StatusOK)
		}))
		t.Cleanup(upstream.Close)

		upstreams[name] = upstream.Listener.Addr().String()
	}

	return upstreams
}

type reply struct {
	StatusCode int
	Header     http.Header
	Body       string
}

func send(t *testing.T, client *http.Client, base, host, target string) reply {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, base+target, nil)
	require.NoError(t, err)
	req.Host = host

	res, err := client.Do(req)
	require.NoError(t, err)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())

	return reply{StatusCode: res.StatusCode, Header: res.Header, Body: string(body)}
}

func noRedirects() *http.Client {
	return &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
}

func assertExchanges(t *testing.T, cfg *GatewayConfig, exchanges []exchange) {
	t.Helper()

	address := startServing(t, cfg, http.HandlerFunc(serverRecorder))

	for _, e := range exchanges {
		t.Run(e.String(), func(t *testing.T) {
			res := send(t, noRedirects(), "http://"+address, e.host, e.target)

			if e.served {
				assert.Equal(t, http.StatusNoContent, res.StatusCode)
				assert.Equal(t, e.target, res.Header.Get("X-Served-URI"))

				return
			}

			assert.Empty(t, res.Header.Get("X-Served-URI"), "the server answered")

			if e.upstream == "" {
				assert.Equal(t, e.status, res.StatusCode)
				assert.Empty(t, res.Header.Get("X-Upstream"), "an upstream answered")

				return
			}

			uri := e.uri
			if uri == "" {
				uri = e.target
			}

			assert.Equal(t, e.upstream, res.Header.Get("X-Upstream"))
			assert.Equal(t, uri, res.Header.Get("X-Upstream-URI"))
			upstreamHost := e.upstreamHost
			if upstreamHost == "" {
				upstreamHost = e.host
			}

			assert.Equal(t, upstreamHost, res.Header.Get("X-Upstream-Host"))
			assert.Equal(t, "127.0.0.1", res.Header.Get("X-Upstream-Real-IP"))
		})
	}
}

func TestTheSiteRoutesEachPathToTheServerOrTheConsole(t *testing.T) {
	domain := "shellhub.example"

	served := func(target string) exchange { return exchange{host: domain, target: target, served: true} }
	console := func(target string) exchange { return exchange{host: domain, target: target, upstream: "console"} }
	vhost := func(name string) exchange { return exchange{host: name + "." + domain, target: "/", upstream: name} }

	community := []exchange{
		served("/api/namespaces"),
		served("/mcp"),
		served("/ssh/connection"),
		served("/agent/connection"),
		served("/ws/session"),
		served("/healthcheck"),
		served("/info"),
		served("/info?agent_version=v0.22.0"),
		served("/install.sh"),
		served("/kickstart.sh"),
		{host: domain, target: "/internal/metrics", status: http.StatusNotFound},
		console("/"),
		console("/devices"),
		console("/admin/api/namespaces"),
		console("/openapi/mock/info"),
		{host: domain + ":8443", target: "/devices", upstream: "console", upstreamHost: domain},
		{host: "elsewhere.example", target: "/api/namespaces", status: http.StatusNotFound},
		{host: endpointAddress + "." + domain, target: "/", status: http.StatusNotFound},
		{host: "mail." + domain, target: "/", status: http.StatusNotFound},
	}

	t.Run("community", func(t *testing.T) {
		assertExchanges(t, configurations()["community"], community)
	})

	t.Run("enterprise", func(t *testing.T) {
		assertExchanges(t, configurations()["enterprise"], []exchange{
			served("/admin/api/namespaces"),
			{host: "mail." + domain, target: "/", status: http.StatusNotFound},
		})
	})

	t.Run("development", func(t *testing.T) {
		assertExchanges(t, configurations()["development"], []exchange{
			{host: domain, target: "/openapi/mock/info", upstream: "openapi-mock", uri: "/info"},
			{host: domain, target: "/openapi/proxy/api/namespaces", upstream: "openapi-proxy", uri: "/api/namespaces"},
			{host: domain, target: "/openapi/spec.json", upstream: "openapi"},
			console("/devices"),
			vhost("website"),
			vhost("docs"),
			vhost("idp"),
			vhost("mail"),
		})
	})

	t.Run("development community", func(t *testing.T) {
		assertExchanges(t, configurations()["dev community"], []exchange{
			vhost("website"),
			{host: "idp." + domain, target: "/", status: http.StatusNotFound},
			{host: "mail." + domain, target: "/", status: http.StatusNotFound},
		})
	})
}

func TestTheInstallScriptsAreNeverCached(t *testing.T) {
	cfg := configurations()["community"]
	address := startServing(t, cfg, http.HandlerFunc(serverRecorder))

	for _, target := range []string{"/install.sh", "/kickstart.sh"} {
		res := send(t, noRedirects(), "http://"+address, cfg.Domain, target)
		assert.Equal(t, "no-store", res.Header.Get("Cache-Control"), target)
	}
}

func TestTheServerRoutesReceiveThePortTheClientReached(t *testing.T) {
	cfg := configurations()["web endpoints"]
	address := startServing(t, cfg, http.HandlerFunc(serverRecorder))

	res := send(t, noRedirects(), "http://"+address, cfg.Domain, "/api/namespaces")
	assert.Equal(t, strconv.Itoa(cfg.httpPort), res.Header.Get("X-Seen-Forwarded-Port"))

	res = send(t, noRedirects(), "http://"+address, cfg.Domain+":8443", "/api/namespaces")
	assert.Equal(t, "8443", res.Header.Get("X-Seen-Forwarded-Port"))

	res = send(t, noRedirects(), "http://"+address, endpointAddress+"."+cfg.Domain, "/")
	assert.Empty(t, res.Header.Get("X-Seen-Forwarded-Port"), "a web endpoint request names no ShellHub port")
}

func TestAWebEndpointIsServedOnlyForADeviceAddress(t *testing.T) {
	domain := configurations()["web endpoints"].Domain

	assertExchanges(t, configurations()["web endpoints"], []exchange{
		{host: endpointAddress + "." + domain, target: "/", served: true},
		{host: endpointAddress + "." + domain, target: "/app/index.html?tab=1", served: true},
		{host: "0123456789ABCDEF0123456789ABCDEF." + domain, target: "/", served: true},
		{host: "not-an-address." + domain, target: "/", status: http.StatusNotFound},
		{host: endpointAddress[:31] + "." + domain, target: "/", status: http.StatusNotFound},
	})
}

func TestTheHealthcheckSiteAnswersOnEveryConfiguration(t *testing.T) {
	for description, cfg := range configurations() {
		t.Run(description, func(t *testing.T) {
			if cfg.EnableAutoSSL && cfg.TLSCertFile == "" {
				t.Skip("an automatic certificate needs an ACME CA")
			}

			if cfg.TLSCertFile != "" {
				cfg.TLSCertFile, cfg.TLSKeyFile = selfSignedCertificate(t, cfg.Domain)
			}

			address := startServing(t, cfg, http.HandlerFunc(serverRecorder))

			res := send(t, noRedirects(), "http://"+address, "healthcheck.internal", "/healthz")

			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.Equal(t, "ok", res.Body)
		})
	}
}

func TestASuppliedCertificateServesTheSiteOverHTTPS(t *testing.T) {
	cfg := configurations()["supplied certificate"]
	cfg.TLSCertFile, cfg.TLSKeyFile = selfSignedCertificate(t, cfg.Domain)

	address := startServing(t, cfg, http.HandlerFunc(serverRecorder))

	t.Run("plain HTTP redirects to HTTPS", func(t *testing.T) {
		res := send(t, noRedirects(), "http://"+address, cfg.Domain, "/api/namespaces?page=2")

		assert.Equal(t, http.StatusPermanentRedirect, res.StatusCode)
		assert.Equal(t, "https://"+cfg.Domain+"/api/namespaces?page=2", res.Header.Get("Location"))
	})

	t.Run("HTTPS presents the certificate and reaches the server", func(t *testing.T) {
		res := send(t, trustingClient(t, cfg), httpsBase(cfg), cfg.Domain, "/api/namespaces")

		assert.Equal(t, http.StatusNoContent, res.StatusCode)
		assert.Equal(t, "https", res.Header.Get("X-Seen-Forwarded-Proto"))
	})
}

func TestTheConsoleRefusesToBeFramed(t *testing.T) {
	cfg := configurations()["community"]
	address := startServing(t, cfg, http.HandlerFunc(serverRecorder))

	for _, target := range []string{"/", "/ssh-identities/new/ABCD1234"} {
		t.Run(target, func(t *testing.T) {
			res := send(t, noRedirects(), "http://"+address, cfg.Domain, target)

			require.Equal(t, "console", res.Header.Get("X-Upstream"))
			assert.Equal(t, []string{"DENY"}, res.Header.Values("X-Frame-Options"),
				"the gateway's value replaces whatever the console server sends")
			assert.Equal(t, []string{"frame-ancestors 'none'"}, res.Header.Values("Content-Security-Policy"))
		})
	}
}

func TestHTTPSTellsBrowsersToStayOnHTTPS(t *testing.T) {
	cfg := configurations()["supplied certificate"]
	cfg.TLSCertFile, cfg.TLSKeyFile = selfSignedCertificate(t, cfg.Domain)

	startServing(t, cfg, http.HandlerFunc(serverRecorder))

	for _, target := range []string{"/", "/api/namespaces"} {
		t.Run(target, func(t *testing.T) {
			res := send(t, trustingClient(t, cfg), httpsBase(cfg), cfg.Domain, target)

			assert.Equal(t, "max-age=31536000", res.Header.Get("Strict-Transport-Security"),
				"no includeSubDomains: web endpoints and the development hosts may still be plain HTTP")
		})
	}
}

func TestPlainHTTPSendsNoHSTS(t *testing.T) {
	cfg := configurations()["community"]
	address := startServing(t, cfg, http.HandlerFunc(serverRecorder))

	res := send(t, noRedirects(), "http://"+address, cfg.Domain, "/")
	assert.Empty(t, res.Header.Get("Strict-Transport-Security"))
}

func httpsBase(cfg *GatewayConfig) string {
	return "https://" + net.JoinHostPort("127.0.0.1", strconv.Itoa(cfg.httpsPort))
}

func trustingClient(t *testing.T, cfg *GatewayConfig) *http.Client {
	t.Helper()

	certificate, err := os.ReadFile(cfg.TLSCertFile)
	require.NoError(t, err)

	roots := x509.NewCertPool()
	require.True(t, roots.AppendCertsFromPEM(certificate))

	client := noRedirects()
	client.Transport = &http.Transport{TLSClientConfig: &tls.Config{RootCAs: roots, ServerName: cfg.Domain, MinVersion: tls.VersionTLS12}}

	return client
}

func selfSignedCertificate(t *testing.T, domain string) (string, string) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	now := clock.New().Now()
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: domain},
		DNSNames:              []string{domain},
		NotBefore:             now.Add(-time.Hour),
		NotAfter:              now.Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)

	encodedKey, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)

	dir := t.TempDir()
	certFile := filepath.Join(dir, "fullchain.pem")
	keyFile := filepath.Join(dir, "privkey.pem")

	require.NoError(t, os.WriteFile(certFile, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o600))
	require.NoError(t, os.WriteFile(keyFile, pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: encodedKey}), 0o600))

	return certFile, keyFile
}
