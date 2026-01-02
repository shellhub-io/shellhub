package environment

import "time"

const (
	// Test timeouts and intervals
	EventuallyTimeout  = 30 * time.Second
	EventuallyInterval = 1 * time.Second

	// Health check timeouts
	HealthCheckTimeout  = 120 * time.Second
	HealthCheckInterval = 500 * time.Millisecond

	// Default test credentials
	DefaultUsername = "test"
	DefaultPassword = "password"
	DefaultEmail    = "test@ossystems.com.br"

	// Default test namespace
	DefaultNamespaceName = "testspace"
	DefaultNamespace     = "00000000-0000-4000-0000-000000000000"

	// Default agent credentials
	DefaultAgentUsername = "root"
	DefaultAgentPassword = "password"
)
