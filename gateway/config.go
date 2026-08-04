package main

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/sethvargo/go-envconfig"
	"github.com/shellhub-io/shellhub/pkg/envs"
)

const defaultProxyTrustedIPs = "0.0.0.0/0 ::/0"

// DNSProvider names the DNS service that answers the DNS-01 challenge, which is
// the only way to prove a wildcard domain.
type DNSProvider string

const (
	// DigitalOceanDNSProvider represents the DigitalOcean DNS provider.
	DigitalOceanDNSProvider DNSProvider = "digitalocean"
	// CloudflareDNSProvider represents the Cloudflare DNS provider.
	CloudflareDNSProvider DNSProvider = "cloudflare"
	// AcmeDNSProvider represents the acme-dns provider.
	AcmeDNSProvider DNSProvider = "acmedns"
)

// GatewayConfig holds the configuration settings for the gateway.
type GatewayConfig struct {
	Env                          string      `env:"SHELLHUB_ENV"`
	Domain                       string      `env:"SHELLHUB_DOMAIN,required" validate:"hostname"`
	WebEndpoints                 bool        `env:"SHELLHUB_WEB_ENDPOINTS,default=false"`
	WebEndpointsDomain           string      `env:"SHELLHUB_WEB_ENDPOINTS_DOMAIN"`
	WebEndpointsDNSProvider      DNSProvider `env:"SHELLHUB_WEB_ENDPOINTS_DNS_PROVIDER,default=digitalocean"`
	WebEndpointsDNSProviderToken string      `env:"SHELLHUB_WEB_ENDPOINTS_DNS_PROVIDER_TOKEN"`
	WebEndpointsAcmeDNSURL       string      `env:"SHELLHUB_WEB_ENDPOINTS_ACME_DNS_URL"`
	WebEndpointsAcmeDNSUsername  string      `env:"SHELLHUB_WEB_ENDPOINTS_ACME_DNS_USERNAME"`
	WebEndpointsAcmeDNSPassword  string      `env:"SHELLHUB_WEB_ENDPOINTS_ACME_DNS_PASSWORD"`
	WebEndpointsAcmeDNSSubdomain string      `env:"SHELLHUB_WEB_ENDPOINTS_ACME_DNS_SUBDOMAIN"`
	EnableAutoSSL                bool        `env:"SHELLHUB_AUTO_SSL"`
	EnableProxyProtocol          bool        `env:"SHELLHUB_PROXY"`
	// defaultProxyTrustedIPs is applied by applyDefaults as well as by the env
	// tag, because the two do not cover the same case: a compose file that
	// passes SHELLHUB_PROXY_TRUSTED_IPS=${SHELLHUB_PROXY_TRUSTED_IPS} against an
	// .env from before this key existed sends the variable set-but-empty, and an
	// env default only applies when the variable is absent. An empty list here
	// trusts nobody, silently, on exactly the deployments that sit behind a
	// balancer.

	// ProxyTrustedIPs names the peers allowed to declare the client's address,
	// through the PROXY protocol preamble and the X-Forwarded-* headers. The
	// default trusts anyone, which is what the previous configuration did; a
	// deployment that knows its load balancer's range should narrow it, because
	// this is the value the login lockout and the GeoIP rules are keyed on.
	ProxyTrustedIPs string `env:"SHELLHUB_PROXY_TRUSTED_IPS"`
	// ACMECAServer is the directory a certificate is asked for. Empty means the
	// default, which is Let's Encrypt's production endpoint; point it at their
	// staging endpoint to rehearse without spending the real rate limit.
	ACMECAServer string `env:"SHELLHUB_ACME_CA_SERVER"`

	// TLSCertFile and TLSKeyFile name a certificate to serve instead of obtaining
	// one, which is the only way to run TLS on a name no public authority will
	// sign.
	//
	// Both or neither, checked even with TLS off: a half-set pair is a typo in
	// every case, and the alternative to refusing it is quietly asking a public
	// CA instead of serving what the operator supplied.
	TLSCertFile      string `env:"SHELLHUB_TLS_CERT_FILE" validate:"required_with=TLSKeyFile"`
	TLSKeyFile       string `env:"SHELLHUB_TLS_KEY_FILE" validate:"required_with=TLSCertFile"`
	Database         string `env:"SHELLHUB_DATABASE,default=mongo"`
	EnableAccessLogs bool   `env:"SHELLHUB_GATEWAY_ACCESS_LOGS,default=true"`

	EnableEnterprise bool
	EnableCloud      bool
	APIBackend       string
}

var validate = validator.New()

// loadGatewayConfig loads and validates the configuration from environment variables.
func loadGatewayConfig() (*GatewayConfig, error) {
	var config GatewayConfig
	if err := envconfig.Process(context.Background(), &config); err != nil {
		return nil, err
	}

	edition, err := envs.ResolveEdition()
	if err != nil {
		return nil, err
	}

	config.EnableEnterprise = edition == envs.Enterprise || edition == envs.Cloud
	config.EnableCloud = edition == envs.Cloud

	config.applyDefaults()

	if err := validate.Struct(config); err != nil {
		return nil, err
	}

	return &config, nil
}

// applyDefaults sets default values for the GatewayConfig if not provided.
func (gc *GatewayConfig) applyDefaults() {
	if gc.ProxyTrustedIPs == "" {
		gc.ProxyTrustedIPs = defaultProxyTrustedIPs
	}

	gc.APIBackend = "server:8080"
}
