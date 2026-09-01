package dialer

import "github.com/prometheus/client_golang/prometheus"

const perProcess = "Counted per server process; sum() across replicas for a fleet total."

var (
	connectionsDesc = prometheus.NewDesc(
		"shellhub_ssh_dialer_connections",
		"Reverse connections currently held, across every device. "+perProcess,
		nil, nil,
	)

	devicesDesc = prometheus.NewDesc(
		"shellhub_ssh_dialer_devices",
		"Devices holding at least one reverse connection. "+perProcess,
		nil, nil,
	)

	displacedDesc = prometheus.NewDesc(
		"shellhub_ssh_dialer_connections_displaced_total",
		"Reverse connections displaced by a newer registration for the same device. "+perProcess,
		nil, nil,
	)
)

// Collector publishes the state of a [Manager]'s connection store.
//
// The two gauges are equal in a healthy store. Every unit of difference is a
// connection a reconnect displaced and whose teardown has not finished. That
// backlog is deliberately not a third series: subtract the two in the query,
// because a metric that is always the difference of two others is one more
// thing to forget to update.
//
// The gauges answer how big the store is right now; the counter answers how
// often devices reconnect over a live one, which is the rate that stays
// informative once the store drains at registration time.
type Collector struct {
	manager *Manager
}

// NewCollector returns a Prometheus collector reporting the tunnel counts manager keeps.
func NewCollector(manager *Manager) *Collector {
	return &Collector{manager: manager}
}

// Describe implements [prometheus.Collector].
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- connectionsDesc
	ch <- devicesDesc
	ch <- displacedDesc
}

// Collect reads counters the manager already maintains rather than walking its
// store. A scrape of a fleet-sized store must not hold the lock the dial path
// reads under.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	stats := c.manager.Stats()

	ch <- prometheus.MustNewConstMetric(connectionsDesc, prometheus.GaugeValue, float64(stats.Connections))
	ch <- prometheus.MustNewConstMetric(devicesDesc, prometheus.GaugeValue, float64(stats.Devices))
	ch <- prometheus.MustNewConstMetric(displacedDesc, prometheus.CounterValue, float64(stats.Displaced))
}
