package redis

type config struct {
	// Specifies the maximum duration in minutes for which a user can be blocked from login attempts.
	// The default value is 32768, equivalent to 15 days.
	MaximumAccountLockout int `env:"MAXIMUM_ACCOUNT_LOCKOUT,default=32768"`
}
