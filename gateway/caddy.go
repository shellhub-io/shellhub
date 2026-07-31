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

// caddyfileData is what the template reads. It exists so every decision that
// needs Go -- which address to listen on, which scheme, where a certificate
// lives -- is made and named here, leaving the template to lay out the result.
type caddyfileData struct {
	Config *GatewayConfig

	Development bool
	TLS         bool

	SiteAddress         string
	WebEndpointsAddress string

	// DNSProvider is the whole `dns` directive body, because each provider
	// takes its credentials differently. Secrets are left as {env....}
	// placeholders so they never appear in the rendered configuration, which is
	// logged at debug level and read by whoever is debugging a deployment.
	DNSProvider string

	ACMECAServer   string
	TrustedProxies string

	// WebEndpointsInternalTLS asks for a locally-signed certificate instead of
	// a public one. Caddy falls back to its internal authority on its own for a
	// plain name it cannot get publicly signed, but not for a wildcard: it
	// decides *.localhost does not qualify and then retries forever, once a
	// minute, for as long as the process runs.
	WebEndpointsInternalTLS bool
}

// dnsProvider renders the credentials for the provider that proves the wildcard.
func dnsProvider(provider DNSProvider) string {
	switch provider {
	case "":
		// Same fallback the environment declares, so a configuration built in
		// code behaves like one built from env vars.
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

	// Only the flag decides. The previous configuration also refused TLS in
	// development, which meant the setting silently did nothing there and the
	// whole certificate path could not be rehearsed before production met it.
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

		// One block for every device tunnel: the wildcard matches exactly one
		// label, which is the address the API reads back off the Host header.
		WebEndpointsAddress: scheme + "*." + domain,
		DNSProvider:         dnsProvider(cfg.WebEndpointsDNSProvider),

		ACMECAServer:   cfg.ACMECAServer,
		TrustedProxies: cfg.ProxyTrustedIPs,

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

	// Through Caddy's own formatter: a template that emits conditional blocks
	// cannot indent them all correctly, and the adapter warns about it. This
	// also means what a customer reads while debugging is canonically laid out.
	return caddyfile.Format(rendered.Bytes()), nil
}
