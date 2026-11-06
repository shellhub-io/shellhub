package gateway

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigNamesTheMissingACMEDNSAccountField(t *testing.T) {
	unsets := map[string]func(*GatewayConfig){
		"WebEndpointsACMEDNSURL":       func(c *GatewayConfig) { c.WebEndpointsACMEDNSURL = "" },
		"WebEndpointsACMEDNSUsername":  func(c *GatewayConfig) { c.WebEndpointsACMEDNSUsername = "" },
		"WebEndpointsACMEDNSSubdomain": func(c *GatewayConfig) { c.WebEndpointsACMEDNSSubdomain = "" },
	}

	for field, unset := range unsets {
		t.Run(field, func(t *testing.T) {
			cfg := configurations()["dns acmedns"]
			unset(cfg)

			var invalid validator.ValidationErrors
			require.ErrorAs(t, validate.Struct(cfg), &invalid)
			require.Len(t, invalid, 1)
			assert.Equal(t, field, invalid[0].Field())
		})
	}
}

func TestConfigAsksForNoACMEDNSAccountWithoutTheWildcard(t *testing.T) {
	for name, mutate := range map[string]func(*GatewayConfig){
		"no web endpoints": func(c *GatewayConfig) { c.WebEndpoints = false },
		"no auto ssl":      func(c *GatewayConfig) { c.EnableAutoSSL = false },
	} {
		t.Run(name, func(t *testing.T) {
			cfg := configurations()["dns acmedns"]
			cfg.WebEndpointsACMEDNSURL = ""
			cfg.WebEndpointsACMEDNSUsername = ""
			cfg.WebEndpointsACMEDNSSubdomain = ""
			mutate(cfg)

			assert.NoError(t, validate.Struct(cfg))
		})
	}
}

func TestLoadGatewayConfigTrustsNoProxyUnlessListed(t *testing.T) {
	t.Setenv("SHELLHUB_DOMAIN", "shellhub.example")

	t.Run("the PROXY protocol without a trusted list is refused", func(t *testing.T) {
		t.Setenv("SHELLHUB_PROXY", "true")
		t.Setenv("SHELLHUB_PROXY_TRUSTED_IPS", "")

		_, err := LoadGatewayConfig(t.Context())

		var invalid validator.ValidationErrors
		require.ErrorAs(t, err, &invalid)
		require.Len(t, invalid, 1)
		assert.Equal(t, "ProxyTrustedIPs", invalid[0].Field())
	})

	t.Run("the PROXY protocol trusts the listed prefixes alone", func(t *testing.T) {
		t.Setenv("SHELLHUB_PROXY", "true")
		t.Setenv("SHELLHUB_PROXY_TRUSTED_IPS", "10.0.0.0/8")

		cfg, err := LoadGatewayConfig(t.Context())
		require.NoError(t, err)
		assert.Equal(t, "10.0.0.0/8", cfg.ProxyTrustedIPs)
	})

	t.Run("without the PROXY protocol an empty list stays empty", func(t *testing.T) {
		t.Setenv("SHELLHUB_PROXY", "false")
		t.Setenv("SHELLHUB_PROXY_TRUSTED_IPS", "")

		cfg, err := LoadGatewayConfig(t.Context())
		require.NoError(t, err)
		assert.Empty(t, cfg.ProxyTrustedIPs)
	})
}
