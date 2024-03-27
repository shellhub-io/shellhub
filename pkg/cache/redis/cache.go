package redis

import (
	"context"
	"math"
	"time"

	rediscache "github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/envs"
	log "github.com/sirupsen/logrus"
)

type redisCache struct {
	cache *rediscache.Cache
	cfg   *config
}

var _ cache.Cache = &redisCache{}

// New creates and returns a new redis cache.
func New(uri string) (cache.Cache, error) {
	opt, err := redis.ParseURL(uri)
	if err != nil {
		return nil, err
	}

	cfg, err := envs.ParseWithPrefix[config]("API_")
	if err != nil {
		log.WithError(err).Fatal("Failed to load environment variables")
	}

	return &redisCache{
		cache: rediscache.New(&rediscache.Options{Redis: redis.NewClient(opt)}),
		cfg:   cfg,
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

func (c *redisCache) uniqueKey(kind, source, id string) string {
	return kind + ":" + source + "-" + id
}

func (c *redisCache) HasAccountLockout(ctx context.Context, source, id string) (int64, int, error) {
	if c.cfg.MaximumAccountLockout <= 0 {
		return 0, 0, nil
	}

	lockout := int64(0)
	if err := c.Get(ctx, c.uniqueKey("account-lockout", source, id), &lockout); err != nil {
		return 0, 0, err
	}

	attempt := 0
	if err := c.Get(ctx, c.uniqueKey("login-attempt", source, id), &attempt); err != nil {
		return 0, 0, err
	}

	return lockout, attempt, nil
}

func (c *redisCache) StoreLoginAttempt(ctx context.Context, source, id string) (int64, int, error) {
	if c.cfg.MaximumAccountLockout <= 0 {
		return 0, 0, nil
	}

	a := 0
	if err := c.Get(ctx, c.uniqueKey("login-attempt", source, id), &a); err != nil {
		return 0, 0, err
	}

	// Start the incognitos
	a++
	curr := clock.Now()
	M := c.cfg.MaximumAccountLockout
	// We store 'x' as 'tmp' to use it as a [float64] instead of [time.Duration]
	// TODO: maybe we can put it in a function.
	tmp := math.Min(math.Pow(4, float64(a-3)), float64(M))
	x := time.Duration(tmp) * time.Minute
	y := time.Duration(math.Max(math.Min(tmp*2.5, float64(M)), 2)) * time.Minute

	if err := c.Set(ctx, c.uniqueKey("login-attempt", source, id), a, y); err != nil {
		return 0, a, err
	}

	if a <= 2 {
		return 0, a, nil
	}

	// We save 'x' as an absolute timestamp to help with time handling
	xTimestamp := curr.Add(x).Unix()

	if err := c.Set(ctx, c.uniqueKey("account-lockout", source, id), xTimestamp, x); err != nil {
		return 0, a, err
	}

	return xTimestamp, a, nil
}

func (c *redisCache) ResetLoginAttempts(ctx context.Context, source, id string) error {
	if c.cfg.MaximumAccountLockout <= 0 {
		return nil
	}

	if err := c.Delete(ctx, c.uniqueKey("login-attempt", source, id)); err != nil {
		return err
	}

	return c.Delete(ctx, c.uniqueKey("account-lockout", source, id))
}
