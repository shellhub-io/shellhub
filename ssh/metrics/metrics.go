package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	connectionsActive = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "shellhub",
			Subsystem: "ssh_server",
			Name:      "connections_active",
			Help:      "Number of active SSH connections",
		},
	)

	connectionsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "shellhub",
			Subsystem: "ssh_server",
			Name:      "connections_total",
			Help:      "Total number of SSH connections accepted",
		},
	)

	authSuccesses = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "shellhub",
			Subsystem: "ssh_server",
			Name:      "auth_successes_total",
			Help:      "Total number of SSH authentication successes",
		},
	)

	authFailures = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "shellhub",
			Subsystem: "ssh_server",
			Name:      "auth_failures_total",
			Help:      "Total number of SSH authentication failures",
		},
	)

	sessionsActive = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "shellhub",
			Subsystem: "ssh_server",
			Name:      "sessions_active",
			Help:      "Number of active SSH sessions (channels)",
		},
	)

	sessionsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "shellhub",
			Subsystem: "ssh_server",
			Name:      "sessions_total",
			Help:      "Total number of SSH sessions (channels) opened",
		},
	)

	dialDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "shellhub",
			Subsystem: "ssh_server",
			Name:      "dial_seconds",
			Help:      "Histogram of dial latency (seconds)",
			Buckets:   prometheus.ExponentialBuckets(0.01, 2, 15),
		},
		[]string{"result"},
	)

	dialFailures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "shellhub",
			Subsystem: "ssh_server",
			Name:      "dial_failures_total",
			Help:      "Total number of dial failures",
		},
		[]string{"reason"},
	)
)

func NewMetrics() *prometheus.Registry {
	p := prometheus.NewRegistry()

	p.MustRegister(connectionsActive)
	p.MustRegister(connectionsTotal)
	p.MustRegister(authFailures)
	p.MustRegister(authSuccesses)
	p.MustRegister(sessionsActive)
	p.MustRegister(sessionsTotal)
	p.MustRegister(dialDuration)
	p.MustRegister(dialFailures)

	return p
}

func IncConnectionsActive() { connectionsActive.Inc() }
func DecConnectionsActive() { connectionsActive.Dec() }
func IncConnectionsTotal()  { connectionsTotal.Inc() }

func RecordAuthFailure() { authFailures.Inc() }
func RecordAuthSuccess() { authSuccesses.Inc() }

func IncSessionsActive() { sessionsActive.Inc() }
func DecSessionsActive() { sessionsActive.Dec() }
func IncSessionsTotal()  { sessionsTotal.Inc() }

func ObserveDialDuration(result string, seconds float64) {
	dialDuration.WithLabelValues(result).Observe(seconds)
}
func IncDialFailure(reason string) { dialFailures.WithLabelValues(reason).Inc() }
