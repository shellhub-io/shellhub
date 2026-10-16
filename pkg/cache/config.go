package cache

import (
	"errors"
	"fmt"
	"math"
	"time"
)

// Config is the Redis cache's configuration. The caller parses it, under the API_ prefix in the
// server, and hands it to [NewRedisCache].
type Config struct {
	// MaximumAccountLockout caps, in minutes, how long a run of failed logins locks an account.
	// Zero or less disables the lockout.
	MaximumAccountLockout int `env:"MAXIMUM_ACCOUNT_LOCKOUT,default=60"`
	// AccountLockoutBase is how long the first lockout lasts; each later one lasts four times the
	// previous. The zero value is one minute.
	AccountLockoutBase LockoutBase `env:"ACCOUNT_LOCKOUT_BASE,default=1m"`
}

// LockoutBase is the length of the first account lockout. It parses from a Go duration of at
// least one second, the resolution lockout deadlines are stored in.
type LockoutBase time.Duration

const (
	defaultLockoutBase = time.Minute
	longestLockout     = time.Duration(1 << 62)
)

var errLockoutBase = errors.New("SHELLHUB_ACCOUNT_LOCKOUT_BASE must be a duration of at least 1s, such as 1m")

// UnmarshalText parses text as a Go duration and returns an error naming
// SHELLHUB_ACCOUNT_LOCKOUT_BASE when it does not parse or is under one second.
func (b *LockoutBase) UnmarshalText(text []byte) error {
	base, err := time.ParseDuration(string(text))
	if err != nil {
		return fmt.Errorf("%w: %w", errLockoutBase, err)
	}

	if base < time.Second {
		return fmt.Errorf("%w: got %s", errLockoutBase, text)
	}

	*b = LockoutBase(base)

	return nil
}

func (c Config) lockout(attempt int) (lockout, attemptTTL time.Duration) {
	base := time.Duration(c.AccountLockoutBase)
	if base == 0 {
		base = defaultLockoutBase
	}

	maximum := math.Min(float64(c.MaximumAccountLockout)*float64(time.Minute), float64(longestLockout))

	scaled := math.Min(float64(base)*math.Pow(4, float64(attempt-3)), maximum)
	lifetime := math.Max(math.Min(scaled*2.5, maximum), float64(2*base))

	return time.Duration(scaled), time.Duration(lifetime)
}
