package gateway

import (
	"bytes"
	_ "embed"
	"fmt"
	"maps"
	"text/template"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/certmagic"
	"github.com/shellhub-io/shellhub/pkg/webendpoints"
)

//go:embed Caddyfile.tmpl
var caddyfileTemplate string

var defaultUpstreams = map[string]string{
	"openapi-mock":  "openapi:4010",
	"openapi-proxy": "openapi:4020",
	"openapi":       "openapi:8080",
	"console":       "ui:8080",
	"website":       "ui:8082",
	"docs":          "ui:8083",
	"idp":           "saml-idp:8080",
	"mail":          "mailpit:8025",
}

type caddyfileData struct {
	Config *GatewayConfig

	HTTPPort  int
	HTTPSPort int
	DataDir   string
	Upstreams map[string]string

	Development bool
	TLS         bool

	SiteAddress         string
	WebEndpointsAddress string

	DNSProvider string

	ACMECAServer   string
	TrustedProxies string
	TLSCertFile    string
	TLSKeyFile     string

	WebEndpointsInternalTLS bool
}

func dnsProvider(cfg *GatewayConfig) string {
	provider := cfg.WebEndpointsDNSProvider
	if provider == "" {
		provider = DigitalOceanDNSProvider
	}

	switch provider {
	case AcmeDNSProvider:
		return fmt.Sprintf(`acmedns {
	username %s
	password {env.SHELLHUB_WEB_ENDPOINTS_ACME_DNS_PASSWORD}
	subdomain %s
	server_url %s
}`, cfg.WebEndpointsACMEDNSUsername, cfg.WebEndpointsACMEDNSSubdomain, cfg.WebEndpointsACMEDNSURL)
	default:
		return string(provider) + " {env.SHELLHUB_WEB_ENDPOINTS_DNS_PROVIDER_TOKEN}"
	}
}

func newCaddyfileData(cfg *GatewayConfig) *caddyfileData {
	development := cfg.Env == "development"

	scheme := "http://"
	if cfg.EnableAutoSSL {
		scheme = "https://"
	}

	domain := webendpoints.Domain(cfg.WebEndpointsDomain, cfg.Domain)

	httpPort := cfg.httpPort
	if httpPort == 0 {
		httpPort = 80
	}

	httpsPort := cfg.httpsPort
	if httpsPort == 0 {
		httpsPort = 443
	}

	dataDir := cfg.dataDir
	if dataDir == "" {
		dataDir = "/data"
	}

	upstreams := maps.Clone(defaultUpstreams)
	maps.Copy(upstreams, cfg.upstreams)

	return &caddyfileData{
		Config:      cfg,
		Upstreams:   upstreams,
		HTTPPort:    httpPort,
		HTTPSPort:   httpsPort,
		DataDir:     dataDir,
		Development: development,
		TLS:         cfg.EnableAutoSSL,

		SiteAddress: scheme + cfg.Domain,

		WebEndpointsAddress: scheme + "*." + domain,
		DNSProvider:         dnsProvider(cfg),

		ACMECAServer:   cfg.ACMECAServer,
		TrustedProxies: cfg.ProxyTrustedIPs,

		TLSCertFile: cfg.TLSCertFile,
		TLSKeyFile:  cfg.TLSKeyFile,

		WebEndpointsInternalTLS: !certmagic.SubjectQualifiesForPublicCert("*." + domain),
	}
}

// Caddyfile renders the Caddyfile for cfg. It returns an error when the template does not parse
// or does not execute against cfg; it does not check that Caddy can adapt the result.
func Caddyfile(cfg *GatewayConfig) ([]byte, error) {
	data := newCaddyfileData(cfg)

	tmpl, err := template.New("Caddyfile").Parse(caddyfileTemplate)
	if err != nil {
		return nil, fmt.Errorf("the Caddyfile template does not parse: %w", err)
	}

	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, data); err != nil {
		return nil, fmt.Errorf("failed to render the Caddyfile: %w", err)
	}

	return caddyfile.Format(rendered.Bytes()), nil
}
