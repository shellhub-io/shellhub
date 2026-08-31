package cache

type config struct {
	MaximumAccountLockout int `env:"MAXIMUM_ACCOUNT_LOCKOUT,default=60"`
}
