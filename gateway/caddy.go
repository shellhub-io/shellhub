package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"text/template"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/certmagic"
	"github.com/shellhub-io/shellhub/pkg/webendpoints"
)

//go:embed Caddyfile.tmpl
var caddyfileTemplate string

type caddyfileData struct {
	Config *GatewayConfig

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

func dnsProvider(provider DNSProvider) string {
	switch provider {
	case "":
		return dnsProvider(DigitalOceanDNSProvider)
	case AcmeDNSProvider:
		return `acmedns {
	username {env.SHELLHUB_WEB_ENDPOINTS_ACME_DNS_USERNAME}
	password {env.SHELLHUB_WEB_ENDPOINTS_ACME_DNS_PASSWORD}
	subdomain {env.SHELLHUB_WEB_ENDPOINTS_ACME_DNS_SUBDOMAIN}
	server_url {env.SHELLHUB_WEB_ENDPOINTS_ACME_DNS_URL}
}`
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

	return &caddyfileData{
		Config:      cfg,
		Development: development,
		TLS:         cfg.EnableAutoSSL,

		SiteAddress: scheme + cfg.Domain,

		WebEndpointsAddress: scheme + "*." + domain,
		DNSProvider:         dnsProvider(cfg.WebEndpointsDNSProvider),

		ACMECAServer:   cfg.ACMECAServer,
		TrustedProxies: cfg.ProxyTrustedIPs,

		TLSCertFile: cfg.TLSCertFile,
		TLSKeyFile:  cfg.TLSKeyFile,

		WebEndpointsInternalTLS: !certmagic.SubjectQualifiesForPublicCert("*." + domain),
	}
}

// Caddyfile builds the configuration this proxy serves, from the environment
// and nothing else. It is called once, when the process starts.
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
