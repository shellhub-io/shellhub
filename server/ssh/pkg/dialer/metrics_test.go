package dialer

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	connectionsMetric = "shellhub_ssh_dialer_connections"
	devicesMetric     = "shellhub_ssh_dialer_devices"
	displacedMetric   = "shellhub_ssh_dialer_connections_displaced_total"
)

// expositionOf renders what a scrape must return. The help text is written out
// once here rather than per case, so rewording it is one edit and a scrape that
// silently changes its own contract still fails.
func expositionOf(connections, devices int) string {
	return fmt.Sprintf(`
		# HELP shellhub_ssh_dialer_connections Reverse connections currently held, across every device. Counted per server process; sum() across replicas for a fleet total.
		# TYPE shellhub_ssh_dialer_connections gauge
		shellhub_ssh_dialer_connections %d
		# HELP shellhub_ssh_dialer_devices Devices holding at least one reverse connection. Counted per server process; sum() across replicas for a fleet total.
		# TYPE shellhub_ssh_dialer_devices gauge
		shellhub_ssh_dialer_devices %d
	`, connections, devices)
}

func displacedExpositionOf(displaced int) string {
	return fmt.Sprintf(`
		# HELP shellhub_ssh_dialer_connections_displaced_total Reverse connections displaced by a newer registration for the same device. Counted per server process; sum() across replicas for a fleet total.
		# TYPE shellhub_ssh_dialer_connections_displaced_total counter
		shellhub_ssh_dialer_connections_displaced_total %d
	`, displaced)
}

func TestCollectorReportsTheStoreSize(t *testing.T) {
	cases := []struct {
		title       string
		setup       func(m *Manager)
		connections int
		devices     int
	}{
		{
			title:       "reports zero for a store nothing has registered on",
			setup:       func(_ *Manager) {},
			connections: 0,
			devices:     0,
		},
		{
			title: "counts one connection per device",
			setup: func(m *Manager) {
				m.Connections.Store(NewKey("tenant", "uid1"), "connection1")
				m.Connections.Store(NewKey("tenant", "uid2"), "connection2")
			},
			connections: 2,
			devices:     2,
		},
		{
			// The gap between the two gauges is the number an operator sizing
			// the store cannot get anywhere else: connections a reconnect
			// displaced and that are not yet torn down.
			title: "separates the gauges when a device holds a second connection",
			setup: func(m *Manager) {
				key := NewKey("tenant", "uid")
				m.Connections.Store(key, "connection1")
				m.Connections.Store(key, "connection2")
			},
			connections: 2,
			devices:     1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			m := NewManager()
			tc.setup(m)

			expected := strings.NewReader(expositionOf(tc.connections, tc.devices))

			assert.NoError(t, testutil.CollectAndCompare(NewCollector(m), expected, connectionsMetric, devicesMetric))
		})
	}
}

// TestCollectorCountsDisplacedConnections drives the counter through the path
// that increments it: a device that reconnects before the server noticed its
// previous socket was gone. That rate is the signal the gauges cannot carry,
// because the store drains at registration time and reads healthy moments
// later.
func TestCollectorCountsDisplacedConnections(t *testing.T) {
	key := NewKey("tenant", "uid")

	m := NewManager()
	collector := NewCollector(m)

	// Every reconnect below waits for the previous teardown, so each one
	// displaces exactly one connection.
	reconnect := func() {
		require.NoError(t, m.Bind("tenant", "uid", newAgentConn(t)))
		require.Eventually(t, func() bool { return m.Connections.Size(key) == 1 },
			time.Second, 10*time.Millisecond, "the displaced connection was never torn down")
	}

	reconnect()
	require.NoError(t, testutil.CollectAndCompare(collector, strings.NewReader(displacedExpositionOf(0)), displacedMetric),
		"a first registration displaces nothing")

	reconnect()
	reconnect()

	assert.NoError(t, testutil.CollectAndCompare(collector, strings.NewReader(displacedExpositionOf(2)), displacedMetric))
}
