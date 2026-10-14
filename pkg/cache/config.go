package cache

// Config is the Redis cache's configuration. The caller parses it, under the API_ prefix in the
// server, and hands it to [NewRedisCache].
type Config struct {
	// MaximumAccountLockout caps, in minutes, how long a run of failed logins locks an account.
	// Zero or less disables the lockout.
	MaximumAccountLockout int `env:"MAXIMUM_ACCOUNT_LOCKOUT,default=60"`
}
