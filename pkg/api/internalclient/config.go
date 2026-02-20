package internalclient

import (
	"github.com/shellhub-io/shellhub/pkg/envs"
)

// Config holds configuration options for the client.
type Config struct {
	// RetryCount defines how many times the client should retry a request in case of failure.
	RetryCount int `env:"SHELLHUB_INTERNAL_HTTP_CLIENT_RETRY_COUNT,default=3"`
	// RetryWaitTime defines the wait time between retries.
	RetryWaitTime int `env:"SHELLHUB_INTERNAL_HTTP_CLIENT_RETRY_WAIT_TIME,default=5"`
	// RetryMaxWaitTime defines the maximum wait time between retries.
	RetryMaxWaitTime int `env:"SHELLHUB_INTERNAL_HTTP_CLIENT_RETRY_MAX_WAIT_TIME,default=20"`

	// APIBaseURL defines the base URL for the API service.
	// All routes — community and enterprise — are served by this single backend.
	APIBaseURL string `env:"SHELLHUB_INTERNAL_HTTP_CLIENT_API_BASE_URL,default=http://api:8080"`
}

func NewConfigFromEnv() (*Config, error) {
	env, err := envs.Parse[Config]()
	if err != nil {
		return nil, err
	}

	return env, nil
}

// DefaultConfig returns a Config struct with default values.
func DefaultConfig() (*Config, error) {
	return &Config{
		RetryCount:       3,
		RetryWaitTime:    5,
		RetryMaxWaitTime: 20,
		APIBaseURL:       "http://api:8080",
	}, nil
}
