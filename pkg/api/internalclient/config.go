package internalclient

import "time"

// Config holds configuration options for the client.
type Config struct {
	// RetryCount defines how many times the client should retry a request in case of failure.
	RetryCount int
	// RetryWaitTime defines the wait time between retries.
	RetryWaitTime time.Duration
	// RetryMaxWaitTime defines the maximum wait time between retries.
	RetryMaxWaitTime time.Duration

	// APIBaseURL defines the base URL for the API service.
	APIBaseURL string

	// EnterpriseBaseURL defines the base URL for enterprise endpoints (cloud component).
	EnterpriseBaseURL string
}

// DefaultConfig returns a Config struct with default values.
func DefaultConfig() (*Config, error) {
	return &Config{
		RetryCount:        3,
		RetryWaitTime:     5 * time.Second,
		RetryMaxWaitTime:  20 * time.Second,
		APIBaseURL:        "http://api:8080",
		EnterpriseBaseURL: "http://cloud:8080",
	}, nil
}
