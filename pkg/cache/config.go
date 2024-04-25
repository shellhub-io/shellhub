package cache

type config struct {
	// Specifies the maximum duration in minutes for which a user can be blocked from login attempts.
	// The default value is 60, equivalent to 1 hour.
	MaximumAccountLockout int `env:"MAXIMUM_ACCOUNT_LOCKOUT,default=60"`
}
