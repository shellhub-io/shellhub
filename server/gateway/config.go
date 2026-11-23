package gateway

import (
	"context"
	"errors"
	"net/netip"

	"github.com/go-playground/validator/v10"
	"github.com/sethvargo/go-envconfig"
	"github.com/shellhub-io/shellhub/pkg/envs"
)

// DNSProvider names the DNS service that answers the DNS-01 challenge, which is
// the only way to prove a wildcard domain.
type DNSProvider string

// The DNS providers SHELLHUB_WEB_ENDPOINTS_DNS_PROVIDER accepts.
const (
	DigitalOceanDNSProvider DNSProvider = "digitalocean"
	CloudflareDNSProvider   DNSProvider = "cloudflare"
	AcmeDNSProvider         DNSProvider = "acmedns"
)

// GatewayConfig is what the gateway reads from the environment: the names it serves, how it gets
// their certificates, and whom it trusts to name the client.
type GatewayConfig struct {
	Env                     string       `env:"SHELLHUB_ENV"`
	Edition                 envs.Edition `env:"SHELLHUB_EDITION"`
	Domain                  string       `env:"SHELLHUB_DOMAIN,required" validate:"hostname|ipv4"`
	WebEndpoints            bool         `env:"SHELLHUB_WEB_ENDPOINTS,default=false"`
	WebEndpointsDomain      string       `env:"SHELLHUB_WEB_ENDPOINTS_DOMAIN"`
	WebEndpointsDNSProvider DNSProvider  `env:"SHELLHUB_WEB_ENDPOINTS_DNS_PROVIDER,default=digitalocean"`
	EnableAutoSSL           bool         `env:"SHELLHUB_AUTO_SSL"`
	EnableProxyProtocol     bool         `env:"SHELLHUB_PROXY"`

	WebEndpointsACMEDNSURL       string `env:"SHELLHUB_WEB_ENDPOINTS_ACME_DNS_URL" validate:"required_if=WebEndpointsDNSProvider acmedns WebEndpoints true EnableAutoSSL true"`
	WebEndpointsACMEDNSUsername  string `env:"SHELLHUB_WEB_ENDPOINTS_ACME_DNS_USERNAME" validate:"required_if=WebEndpointsDNSProvider acmedns WebEndpoints true EnableAutoSSL true"`
	WebEndpointsACMEDNSSubdomain string `env:"SHELLHUB_WEB_ENDPOINTS_ACME_DNS_SUBDOMAIN" validate:"required_if=WebEndpointsDNSProvider acmedns WebEndpoints true EnableAutoSSL true"`

	ProxyTrustedIPs string `env:"SHELLHUB_PROXY_TRUSTED_IPS" validate:"required_if=EnableProxyProtocol true"`
	ACMECAServer    string `env:"SHELLHUB_ACME_CA_SERVER"`

	TLSCertFile      string `env:"SHELLHUB_TLS_CERT_FILE" validate:"required_with=TLSKeyFile"`
	TLSKeyFile       string `env:"SHELLHUB_TLS_KEY_FILE" validate:"required_with=TLSCertFile"`
	EnableAccessLogs bool   `env:"SHELLHUB_GATEWAY_ACCESS_LOGS,default=true"`

	EnableEnterprise bool

	httpPort  int
	httpsPort int
	dataDir   string
	upstreams map[string]string
}

var validate = validator.New()

// ErrAutoSSLNeedsADomainName is returned when SHELLHUB_AUTO_SSL is on and SHELLHUB_DOMAIN is an IP
// address with no certificate supplied, since automatic HTTPS can obtain a certificate only for a
// domain name.
var ErrAutoSSLNeedsADomainName = errors.New("SHELLHUB_AUTO_SSL needs a domain name in SHELLHUB_DOMAIN, not an IP address, unless SHELLHUB_TLS_CERT_FILE supplies the certificate")

// LoadGatewayConfig reads the gateway's configuration from the environment, fills in the
// defaults and validates it. It returns an error wrapping envconfig.ErrMissingRequired when
// SHELLHUB_DOMAIN is unset, the envconfig error for a variable that does not parse, the edition's
// error for an unknown SHELLHUB_EDITION, and validator.ValidationErrors for a value that parses
// but is not allowed, SHELLHUB_PROXY on with no SHELLHUB_PROXY_TRUSTED_IPS or a SHELLHUB_DOMAIN that
// is neither a hostname nor an IPv4 address among them. It returns ErrAutoSSLNeedsADomainName when
// SHELLHUB_AUTO_SSL is on for an IP address with no SHELLHUB_TLS_CERT_FILE.
func LoadGatewayConfig(ctx context.Context) (*GatewayConfig, error) {
	var config GatewayConfig
	if err := envconfig.Process(ctx, &config); err != nil {
		return nil, err
	}

	config.EnableEnterprise = config.Edition.IsEnterpriseOrCloud()

	if err := validate.Struct(config); err != nil {
		return nil, err
	}

	if config.EnableAutoSSL && config.TLSCertFile == "" && isIPAddress(config.Domain) {
		return nil, ErrAutoSSLNeedsADomainName
	}

	return &config, nil
}

func isIPAddress(domain string) bool {
	_, err := netip.ParseAddr(domain)

	return err == nil
}
