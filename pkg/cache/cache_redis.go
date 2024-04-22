package cache

import (
	"context"
	"sync"
	"time"

	rediscache "github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type redisCache struct {
	client *redis.Client
	cache  *rediscache.Cache
	mu     sync.Mutex
}

var _ Cache = &redisCache{}

// NewRedisCache creates and returns a new redis cache.
func NewRedisCache(uri string, pool int) (Cache, error) {
	opt, err := redis.ParseURL(uri)
	if err != nil {
		return nil, err
	}

	if pool > 0 {
		opt.PoolSize = pool
	}

	client := redis.NewClient(opt)

	return &redisCache{
		client: client,
		cache: rediscache.New(&rediscache.Options{
			Redis: client,
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

func (c *redisCache) GetLastSeen(ctx context.Context, tenant, uid string) (time.Time, bool, error) {
	zeroer := time.Time{}
	tt := time.Time{}

	if err := c.Get(ctx, "last_seen:"+tenant+"={"+uid+"}", &tt); err != nil {
		return tt, false, err
	}

	return tt, tt != zeroer, nil
}

func (c *redisCache) SetLastSeen(ctx context.Context, tenant, uid string, tt time.Time) error {
	return c.Set(ctx, "last_seen:"+tenant+"={"+uid+"}", tt, 2*time.Minute)
}

func (c *redisCache) DelLastSeen(ctx context.Context, tenant, uid string) error {
	return c.Delete(ctx, "last_seen:"+tenant+"={"+uid+"}")
}

func (c *redisCache) CountConnectedDevices(ctx context.Context, tenant string, status models.DeviceStatus) (int64, error) {
	// Since `c.client.[IncrBy|DecrBy]` does not use the caching algorithm, we need to use `c.client.Get`
	// instead of `c.Get` or `c.cache.Get`.
	cmd := c.client.Get(ctx, "connected_devices:"+tenant+"={"+string(status)+"}")
	if cmd.Err() == redis.Nil {
		return 0, nil // Maintain compatibility with `c.Get` and avoid returning an error if the key doesn't exist
	}

	return cmd.Int64()
}

func (c *redisCache) IncreaseConnectedDevices(ctx context.Context, tenant string, status models.DeviceStatus, t int) error {
	return c.client.IncrBy(ctx, "connected_devices:"+tenant+"={"+string(status)+"}", int64(t)).Err()
}

func (c *redisCache) DecreaseConnectedDevices(ctx context.Context, tenant string, status models.DeviceStatus, t int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Prevent decrementing to -1 or less.
	// Since `c.client.[IncrBy|DecrBy]` does not use the caching algorithm, we need to use `c.client.Get`
	// instead of `c.Get` or `c.cache.Get`.
	if v, e := c.client.Get(ctx, "connected_devices:"+tenant+"={"+string(status)+"}").Int64(); e != nil || v == 0 {
		return nil // nolint:nilerr
	}

	return c.client.DecrBy(ctx, "connected_devices:"+tenant+"={"+string(status)+"}", int64(t)).Err()
}
