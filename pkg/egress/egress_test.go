package egress

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// The guardian rejects internal/reserved targets at the socket layer, so they are
// never reached and report unreachable. Loopback, link-local, and unspecified are
// blocked regardless of the development allowlist (which only adds RFC1918 ranges).
func TestProbeGuardsEgress(t *testing.T) {
	for _, host := range []string{
		"127.0.0.1",
		"::1",
		"169.254.169.254",
		"0.0.0.0",
		"this.host.does.not.exist.invalid",
	} {
		t.Run(host, func(t *testing.T) {
			assert.False(t, Reachable(context.Background(), host, 22))
		})
	}
}
