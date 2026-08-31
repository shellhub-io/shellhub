package envs

import (
	"errors"
	"fmt"
	"strings"
)

// Edition is the ShellHub edition an instance runs as. It decides which features the
// server exposes, so it is read at startup and never per request.
type Edition string

const (
	// Community is the open-source edition, and the edition an instance falls back to when
	// SHELLHUB_EDITION is unset.
	Community Edition = "community"
	// Enterprise is the self-hosted paid edition, gated by a licence file.
	Enterprise Edition = "enterprise"
	// Cloud is the hosted edition, which adds billing and the multi-tenant surface on top of
	// Enterprise.
	Cloud Edition = "cloud"
)

// Backend is an interface for any sort of underlying key/value store.
type Backend interface {
	Get(key string) string
	Process(prefix string, spec any) error
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

// IsCommunity reports whether this instance runs the community edition. It panics on an
// unrecognized SHELLHUB_EDITION, as CurrentEdition does.
func IsCommunity() bool {
	return CurrentEdition() == Community
}

// IsEnterprise reports whether this instance runs the enterprise edition, which is not the
// same question as "may it use a paid feature" — see IsEnterpriseOrCloud.
func IsEnterprise() bool {
	return CurrentEdition() == Enterprise
}

// IsCloud reports whether this instance runs the hosted edition. Billing and namespace
// limits are the features that turn on here and nowhere else.
func IsCloud() bool {
	return CurrentEdition() == Cloud
}

// IsEnterpriseOrCloud reports whether the paid feature set is available. This is the check a
// feature gate wants, so that a feature added for cloud stays available to enterprise.
func IsEnterpriseOrCloud() bool {
	return CurrentEdition() != Community
}

// IsDevelopment reports whether SHELLHUB_ENV is "development". It gates developer conveniences
// only; never use it to relax a security decision, as the variable is attacker-controlled in
// any deployment that passes the environment through.
func IsDevelopment() bool {
	return DefaultBackend.Get("SHELLHUB_ENV") == "development"
}

// ErrParseWithPrefix is joined with the backend's error when ParseWithPrefix fails, so a
// caller can tell a configuration problem from anything else with errors.Is.
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

// ErrParse is joined with the backend's error when Parse fails, so a caller can tell a
// configuration problem from anything else with errors.Is.
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
