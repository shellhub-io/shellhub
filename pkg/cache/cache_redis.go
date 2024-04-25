package cache

import (
	"context"
	"math"
	"strconv"
	"time"

	rediscache "github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/envs"
	log "github.com/sirupsen/logrus"
)

type redisCache struct {
	cache *rediscache.Cache
	cfg   *config
}

var _ Cache = &redisCache{}

func NewRedisCache(uri string, pool int) (Cache, error) {
	opt, err := redis.ParseURL(uri)
	if err != nil {
		return nil, err
	}

	if pool > 0 {
		opt.PoolSize = pool
	}

	cfg, err := envs.ParseWithPrefix[config]("API_")
	if err != nil {
		log.WithError(err).Fatal("Failed to load environment variables")
	}

	return &redisCache{
		cfg: cfg,
		cache: rediscache.New(&rediscache.Options{
			Redis: redis.NewClient(opt),
		}),
	}, nil
}

// Get gets the cache value for the given key.
// NOTE: missing key is not an error.
func (c *redisCache) Get(ctx context.Context, key string, value interface{}) error {
	err := c.cache.Get(ctx, key, value)
	if err == rediscache.ErrCacheMiss {
		return nil
	}

	return err
}

// Set puts value into cache with key and expire time.
func (c *redisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return c.cache.Set(&rediscache.Item{Ctx: ctx, Key: key, Value: value, TTL: ttl})
}

// Delete deletes cached value by given key.
func (c *redisCache) Delete(ctx context.Context, key string) error {
	if err := c.cache.Get(ctx, key, nil); err == rediscache.ErrCacheMiss {
		return nil
	}

	return c.cache.Delete(ctx, key)
}

func (c *redisCache) HasAccountLockout(ctx context.Context, source, id string) (int64, int, error) {
	if c.cfg.MaximumAccountLockout <= 0 {
		return 0, 0, nil
	}

	lockoutSTR := "0"
	if err := c.Get(ctx, "account-lockout="+source+":"+id, &lockoutSTR); err != nil {
		return 0, 0, err
	}

	attemptSTR := "0"
	if err := c.Get(ctx, "login-attempt="+source+":"+id, &attemptSTR); err != nil {
		return 0, 0, err
	}

	lockout, _ := strconv.ParseInt(lockoutSTR, 10, 0)
	attempt, _ := strconv.Atoi(attemptSTR)

	return lockout, attempt, nil
}

func (c *redisCache) StoreLoginAttempt(ctx context.Context, source, id string) (int64, int, error) {
	if c.cfg.MaximumAccountLockout <= 0 {
		return 0, 0, nil
	}

	attemptSTR := "0"
	if err := c.Get(ctx, "login-attempt="+source+":"+id, &attemptSTR); err != nil {
		return 0, 0, err
	}

	attempt, _ := strconv.Atoi(attemptSTR)
	attempt++

	now := clock.Now()
	tmp := math.Min(math.Pow(4, float64(attempt-3)), float64(c.cfg.MaximumAccountLockout))

	attemptTTL := time.Duration(math.Max(math.Min(tmp*2.5, float64(c.cfg.MaximumAccountLockout)), 2)) * time.Minute
	if err := c.Set(ctx, "login-attempt="+source+":"+id, strconv.Itoa(attempt), attemptTTL); err != nil {
		return 0, attempt, err
	}

	if attempt <= 2 {
		return 0, attempt, nil
	}

	// We save 'lockoutTTL' as an absolute lockoutStr to help with time handling
	lockoutTTL := time.Duration(tmp) * time.Minute
	lockoutSTR := strconv.FormatInt(now.Add(lockoutTTL).Unix(), 10)
	if err := c.Set(ctx, "account-lockout="+source+":"+id, lockoutSTR, lockoutTTL); err != nil {
		return 0, attempt, err
	}

	lockout, _ := strconv.ParseInt(lockoutSTR, 10, 0)

	return lockout, attempt, nil
}

func (c *redisCache) ResetLoginAttempts(ctx context.Context, source, id string) error {
	if c.cfg.MaximumAccountLockout <= 0 {
		return nil
	}

	if err := c.Delete(ctx, "login-attempt="+source+":"+id); err != nil {
		return err
	}

	return c.Delete(ctx, "account-lockout="+source+":"+id)
}
