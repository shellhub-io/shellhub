package cache_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/testimage"
	"github.com/shellhub-io/shellhub/pkg/testport"
	"github.com/shellhub-io/shellhub/pkg/testrun"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/redis"
)

func TestRedisCacheLockoutEndsAtAnnouncedDeadline(t *testing.T) {
	ctx := context.Background()

	uri := startValkey(t)

	lockedMidSecond := time.Unix(1_700_000_000, 600_000_000)
	clk := clock.NewFixed(lockedMidSecond)

	c, err := cache.NewRedisCache(uri, 0, cache.Config{MaximumAccountLockout: 60}, clk)
	require.NoError(t, err)

	var deadline int64
	for range 3 {
		deadline, _, err = c.StoreLoginAttempt(ctx, "10.0.0.1", "user")
		require.NoError(t, err)
	}

	require.Equal(t, int64(1_700_000_060), deadline)

	clk.Set(time.Unix(deadline-1, 0))
	lockout, _, err := c.HasAccountLockout(ctx, "10.0.0.1", "user")
	require.NoError(t, err)
	assert.Equal(t, deadline, lockout)

	clk.Set(time.Unix(deadline, 0))
	lockout, _, err = c.HasAccountLockout(ctx, "10.0.0.1", "user")
	require.NoError(t, err)
	assert.Zero(t, lockout)

	var storedDeadline string
	require.NoError(t, c.Get(ctx, "account-lockout=10.0.0.1:user", &storedDeadline))
	assert.Equal(t, "1700000060", storedDeadline)
}

func TestRedisCacheLockoutGrowsFromTheBase(t *testing.T) {
	ctx := context.Background()

	uri := startValkey(t)

	opt, err := goredis.ParseURL(uri)
	require.NoError(t, err)

	client := goredis.NewClient(opt)
	t.Cleanup(func() { _ = client.Close() })

	type attempt struct {
		lockout    time.Duration
		attemptTTL time.Duration
	}

	cases := []struct {
		description string
		cfg         cache.Config
		attempts    []attempt
	}{
		{
			description: "unset, the base is one minute",
			cfg:         cache.Config{MaximumAccountLockout: 60},
			attempts: []attempt{
				{0, 2 * time.Minute},
				{0, 2 * time.Minute},
				{time.Minute, 150 * time.Second},
				{4 * time.Minute, 10 * time.Minute},
				{16 * time.Minute, 40 * time.Minute},
				{60 * time.Minute, 60 * time.Minute},
			},
		},
		{
			description: "a configured base scales the lockout and the attempt lifetime",
			cfg:         cache.Config{MaximumAccountLockout: 60, AccountLockoutBase: cache.LockoutBase(5 * time.Second)},
			attempts: []attempt{
				{0, 10 * time.Second},
				{0, 10 * time.Second},
				{5 * time.Second, 12500 * time.Millisecond},
				{20 * time.Second, 50 * time.Second},
				{80 * time.Second, 200 * time.Second},
				{320 * time.Second, 800 * time.Second},
				{1280 * time.Second, 3200 * time.Second},
				{60 * time.Minute, 60 * time.Minute},
			},
		},
		{
			description: "the maximum caps the lockout and the attempt lifetime",
			cfg:         cache.Config{MaximumAccountLockout: 3, AccountLockoutBase: cache.LockoutBase(time.Minute)},
			attempts: []attempt{
				{0, 2 * time.Minute},
				{0, 2 * time.Minute},
				{time.Minute, 150 * time.Second},
				{3 * time.Minute, 3 * time.Minute},
			},
		},
		{
			description: "a maximum beyond what a duration holds still locks",
			cfg:         cache.Config{MaximumAccountLockout: 200_000_000},
			attempts: []attempt{
				{0, 2 * time.Minute},
				{0, 2 * time.Minute},
				{time.Minute, 150 * time.Second},
				{4 * time.Minute, 10 * time.Minute},
			},
		},
		{
			description: "the attempt lifetime never drops below twice the base",
			cfg:         cache.Config{MaximumAccountLockout: 1, AccountLockoutBase: cache.LockoutBase(time.Minute)},
			attempts: []attempt{
				{0, 2 * time.Minute},
				{0, 2 * time.Minute},
				{time.Minute, 2 * time.Minute},
			},
		},
	}

	for i, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			lockedAt := time.Unix(1_700_000_000, 0)

			c, err := cache.NewRedisCache(uri, 0, tc.cfg, clock.NewFixed(lockedAt))
			require.NoError(t, err)

			user := strconv.Itoa(i)

			for n, want := range tc.attempts {
				deadline, _, err := c.StoreLoginAttempt(ctx, "10.0.0.1", user)
				require.NoError(t, err)

				if want.lockout == 0 {
					assert.Zero(t, deadline, "attempt %d", n+1)
				} else {
					assert.Equal(t, lockedAt.Add(want.lockout).Unix(), deadline, "attempt %d", n+1)
				}

				ttl, err := client.PTTL(ctx, "login-attempt=10.0.0.1:"+user).Result()
				require.NoError(t, err)
				assert.InDelta(t, want.attemptTTL, ttl, float64(time.Second), "attempt %d", n+1)
			}
		})
	}
}

func TestRedisCacheCompareAndDeleteKeepsAnotherOwnersKey(t *testing.T) {
	ctx := context.Background()

	uri := startValkey(t)

	c, err := cache.NewRedisCache(uri, 0, cache.Config{MaximumAccountLockout: 60}, clock.New())
	require.NoError(t, err)

	taken, err := c.SetNX(ctx, "guard", "owner", time.Minute)
	require.NoError(t, err)
	require.True(t, taken)

	deleted, err := c.CompareAndDelete(ctx, "guard", "someone-else")
	require.NoError(t, err)
	assert.False(t, deleted)

	taken, err = c.SetNX(ctx, "guard", "intruder", time.Minute)
	require.NoError(t, err)
	assert.False(t, taken)

	deleted, err = c.CompareAndDelete(ctx, "guard", "owner")
	require.NoError(t, err)
	assert.True(t, deleted)

	taken, err = c.SetNX(ctx, "guard", "next", time.Minute)
	require.NoError(t, err)
	assert.True(t, taken)
}

func startValkey(t *testing.T) string {
	t.Helper()

	valkey, err := redis.Run(t.Context(), testimage.DockerHub("valkey/valkey:9.1-alpine"), testport.Bind("6379/tcp"), testrun.Labelled())
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, valkey.Terminate(context.WithoutCancel(t.Context())))
	})

	uri, err := valkey.ConnectionString(t.Context())
	require.NoError(t, err)

	return uri
}
