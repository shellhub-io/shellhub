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

// ParseEdition normalizes raw (trim and lowercase) into an Edition. The empty string is
// Community; anything else outside the three editions is an error.
func ParseEdition(raw string) (Edition, error) {
	normalized := strings.TrimSpace(strings.ToLower(raw))
	if normalized == "" {
		return Community, nil
	}

	switch Edition(normalized) {
	case Community, Enterprise, Cloud:
		return Edition(normalized), nil
	default:
		return "", fmt.Errorf("invalid SHELLHUB_EDITION %q: must be community, enterprise, or cloud", normalized)
	}
}

// UnmarshalText parses text with ParseEdition, so a config struct that declares an Edition field
// rejects a bad SHELLHUB_EDITION when it is parsed.
func (e *Edition) UnmarshalText(text []byte) error {
	edition, err := ParseEdition(string(text))
	if err != nil {
		return err
	}

	*e = edition

	return nil
}

// IsCommunity reports whether e is the community edition. The zero value is community too, so an
// edition nobody set behaves as an unset SHELLHUB_EDITION does.
func (e Edition) IsCommunity() bool {
	return e == "" || e == Community
}

// IsEnterprise reports whether e is the enterprise edition, which is not the same question as
// "may it use a paid feature"; see IsEnterpriseOrCloud.
func (e Edition) IsEnterprise() bool {
	return e == Enterprise
}

// IsCloud reports whether e is the hosted edition. Billing and namespace limits are the features
// that turn on here and nowhere else.
func (e Edition) IsCloud() bool {
	return e == Cloud
}

// IsEnterpriseOrCloud reports whether the paid feature set is available. This is the check a
// feature gate wants, so that a feature added for cloud stays available to enterprise.
func (e Edition) IsEnterpriseOrCloud() bool {
	return !e.IsCommunity()
}

// Backend is where a configuration is parsed from: the process environment ([OS]) or, in a
// test, a map ([Map]).
type Backend interface {
	Get(key string) string
	Process(prefix string, spec any) error
}

// OpenAPIValidation is how the server treats a response that does not match its schema.
type OpenAPIValidation string

const (
	// OpenAPIValidationOff checks nothing.
	OpenAPIValidationOff OpenAPIValidation = "off"
	// OpenAPIValidationReport logs a mismatch and answers as the handler did.
	OpenAPIValidationReport OpenAPIValidation = "report"
	// OpenAPIValidationStrict answers a mismatch with a 500 that lists it and refuses to start
	// without a schema.
	OpenAPIValidationStrict OpenAPIValidation = "strict"
)

// OpenAPIValidationMode turns the raw SHELLHUB_OPENAPI_VALIDATION into a mode. on, true, 1 and
// report log a mismatch, strict fails it, and off, false and 0 disable the check. Anything else
// reports in development and is off elsewhere.
func OpenAPIValidationMode(raw string, development bool) OpenAPIValidation {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "on", "true", "1", "report":
		return OpenAPIValidationReport
	case "strict":
		return OpenAPIValidationStrict
	case "off", "false", "0":
		return OpenAPIValidationOff
	default:
		if development {
			return OpenAPIValidationReport
		}

		return OpenAPIValidationOff
	}
}

// ErrParseWithPrefix is joined with the backend's error when ParseWithPrefix fails, so a
// caller can tell a configuration problem from anything else with errors.Is.
var ErrParseWithPrefix = errors.New("failed to parse environment variables for the given prefix")

// ParseWithPrefix fills a T from the [envconfig] tags on its fields, reading each key from backend
// under prefix before falling back to the unprefixed key. It returns ErrParseWithPrefix, joined
// with the cause, when a required key is missing or a value does not decode.
//
// [envconfig]: https://github.com/sethvargo/go-envconfig
func ParseWithPrefix[T any](backend Backend, prefix string) (*T, error) {
	envs := new(T)

	if err := backend.Process(prefix, envs); err != nil {
		return nil, errors.Join(ErrParseWithPrefix, err)
	}

	return envs, nil
}

// ErrParse is joined with the backend's error when Parse fails, so a caller can tell a
// configuration problem from anything else with errors.Is.
var ErrParse = errors.New("failed to parse environment variables")

// Parse fills a T from the [envconfig] tags on its fields, reading each key from backend. It
// returns ErrParse, joined with the cause, when a required key is missing or a value does not
// decode.
//
// [envconfig]: https://github.com/sethvargo/go-envconfig
func Parse[T any](backend Backend) (*T, error) {
	envs := new(T)

	if err := backend.Process("", envs); err != nil {
		return nil, errors.Join(ErrParse, err)
	}

	return envs, nil
}
