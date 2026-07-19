package envs

import (
	"errors"
	"fmt"
	"strings"
)

type Edition string

const (
	Community  Edition = "community"
	Enterprise Edition = "enterprise"
	Cloud      Edition = "cloud"
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

// ResolveEdition reads SHELLHUB_EDITION, normalizes it (trim + lowercase) and
// defaults to Community when empty. It returns an error for any unrecognized
// value so entrypoints can validate the edition once at startup and fail fast
// with a clean message instead of relying on CurrentEdition's lazy panic.
func ResolveEdition() (Edition, error) {
	raw := strings.TrimSpace(strings.ToLower(DefaultBackend.Get("SHELLHUB_EDITION")))
	if raw == "" {
		return Community, nil
	}

	switch Edition(raw) {
	case Community, Enterprise, Cloud:
		return Edition(raw), nil
	default:
		return "", fmt.Errorf("invalid SHELLHUB_EDITION %q: must be community, enterprise, or cloud", raw)
	}
}

// CurrentEdition returns the resolved edition, panicking on an unrecognized
// value so a misconfigured instance fails rather than silently running as
// community. Entrypoints should call ResolveEdition at startup to surface the
// error cleanly; this panic is a last-resort invariant for the predicates below.
func CurrentEdition() Edition {
	edition, err := ResolveEdition()
	if err != nil {
		panic(err)
	}

	return edition
}

func IsCommunity() bool {
	return CurrentEdition() == Community
}

func IsEnterprise() bool {
	return CurrentEdition() == Enterprise
}

func IsCloud() bool {
	return CurrentEdition() == Cloud
}

func IsEnterpriseOrCloud() bool {
	return CurrentEdition() != Community
}

func IsDevelopment() bool {
	return DefaultBackend.Get("SHELLHUB_ENV") == "development"
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

var ErrParse = errors.New("failed to parse environment variables")

// Parse parses the environment variables.
//
// This function uses the [envconfig] package as its default backend, so it requires the struct to be annotated with
// the [envconfig] tags. Check the [envconfig] documentation for more information.
//
// The T generic parameter must be a struct with the fields annotated with the [envconfig] tags, that will be returned
// with the values parsed from the environment variables.
//
// [envconfig]: https://github.com/sethvargo/go-envconfig
func Parse[T any]() (*T, error) {
	envs := new(T)

	if err := DefaultBackend.Process("", envs); err != nil {
		return nil, errors.Join(ErrParse, err)
	}

	return envs, nil
}
