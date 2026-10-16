package main

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigNamesTheMissingACMEDNSAccountField guards the error an operator
// reads when the acme-dns account is half set: the field left empty, before
// Caddy reports a directive without its argument.
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

// TestConfigAsksForNoACMEDNSAccountWithoutTheWildcard guards the configurations
// that never render the acme-dns provider: without web endpoints or TLS, no
// wildcard is proven, so an unset account is not an error.
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
