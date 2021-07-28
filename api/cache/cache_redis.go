package cache

import (
	"context"
	"time"

	rediscache "github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
)

type redisCache struct {
	cache *rediscache.Cache
}

var _ Cache = &redisCache{}

// NewRedisCache creates and returns a new redis cache.
func NewRedisCache(uri string) (Cache, error) {
	opt, err := redis.ParseURL(uri)
	if err != nil {
		return nil, err
	}

	return &redisCache{
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
