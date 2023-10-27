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

var ErrParsePrefix = errors.New("failed to parse environment variables for the given prefix")

// ParseWithPrefix parses the environment variables using the given prefix.
//
// This function uses the [env] package as its default backend, so it requires the struct to be annotated with
// the [env] tags. Check the [env] documentation for more information.
//
// The T generic parameter must be a struct with the fields annotated with the [env] tags, that will be returned
// with the values parsed from the environment variables.
//
// [env]: https://github.com/caarlos0/env
func ParseWithPrefix[T any](prefix string) (*T, error) {
	envs := new(T)

	if err := DefaultBackend.Process(prefix, envs); err != nil {
		return nil, errors.Join(ErrParsePrefix, err)
	}

	return envs, nil
}
