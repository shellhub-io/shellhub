package envs

import (
	"errors"
)

const (
	ENABLED = "true"
)

// Backend is an interface for any sort of underlying key/value store.
type Backend interface {
	Get(key string) string
	Process(prefix string, spec interface{}) error
}

// DefaultBackend define the backend to be used to get environment variables.
var DefaultBackend Backend

func init() {
	DefaultBackend = &envBackend{}
}

// IsEnterprise returns true if the current ShellHub server instance is enterprise.
func IsEnterprise() bool {
	return DefaultBackend.Get("SHELLHUB_ENTERPRISE") == ENABLED
}

// IsCloud returns true if the current ShellHub server instance is cloud.
func IsCloud() bool {
	return DefaultBackend.Get("SHELLHUB_CLOUD") == ENABLED
}

// HasBilling returns true if the current ShellHub server instance has billing feature enabled.
func HasBilling() bool {
	return DefaultBackend.Get("SHELLHUB_BILLING") == ENABLED
}

// IsCommunity return true if the current ShellHub server instance is community.
// It evaluates if the current ShellHub instance is neither enterprise or cloud .
func IsCommunity() bool {
	return (DefaultBackend.Get("SHELLHUB_CLOUD") != ENABLED && DefaultBackend.Get("SHELLHUB_ENTERPRISE") != ENABLED)
}

var ErrParseWithPrefix = errors.New("failed to parse environment variables for the given prefix")

// ParseWithPrefix parses the environment variables for the a given prefix.
//
// This function uses the [envconfig] package as its default backend, so it requires the struct to be annotated with
// the [envconfig] tags. Check the [envconfig] documentation for more information.
//
// The T generic parameter must be a struct with the fields annotated with the [envconfig] tags, that will be returned
// with the values parsed from the environment variables.
//
// [envconfig]: https://github.com/sethvargo/go-envconfig
func ParseWithPrefix[T any](prefix string) (*T, error) {
	envs := new(T)

	if err := DefaultBackend.Process(prefix, envs); err != nil {
		return nil, errors.Join(ErrParseWithPrefix, err)
	}

	return envs, nil
}
