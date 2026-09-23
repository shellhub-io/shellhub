package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/clock"
	clockmock "github.com/shellhub-io/shellhub/pkg/clock/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/redis"
)

func setNow(t *testing.T, now time.Time) {
	t.Helper()

	clockMock := clockmock.NewMockClock(t)
	clockMock.On("Now").Return(now).Maybe()

	previous := clock.DefaultBackend
	clock.DefaultBackend = clockMock
	t.Cleanup(func() { clock.DefaultBackend = previous })
}

func TestRedisCacheLockoutEndsAtAnnouncedDeadline(t *testing.T) {
	t.Setenv("API_MAXIMUM_ACCOUNT_LOCKOUT", "60")
	ctx := context.Background()

	redisContainer, err := redis.Run(ctx, "docker.io/valkey/valkey:9.1-alpine")
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, redisContainer.Terminate(ctx))
	})

	uri, err := redisContainer.ConnectionString(ctx)
	require.NoError(t, err)

	c, err := cache.NewRedisCache(uri, 0)
	require.NoError(t, err)

	lockedMidSecond := time.Unix(1_700_000_000, 600_000_000)
	setNow(t, lockedMidSecond)

	var deadline int64
	for range 3 {
		deadline, _, err = c.StoreLoginAttempt(ctx, "10.0.0.1", "user")
		require.NoError(t, err)
	}

	require.Equal(t, int64(1_700_000_060), deadline)

	setNow(t, time.Unix(deadline-1, 0))
	lockout, _, err := c.HasAccountLockout(ctx, "10.0.0.1", "user")
	require.NoError(t, err)
	assert.Equal(t, deadline, lockout)

	setNow(t, time.Unix(deadline, 0))
	lockout, _, err = c.HasAccountLockout(ctx, "10.0.0.1", "user")
	require.NoError(t, err)
	assert.Zero(t, lockout)

	var storedDeadline string
	require.NoError(t, c.Get(ctx, "account-lockout=10.0.0.1:user", &storedDeadline))
	assert.Equal(t, "1700000060", storedDeadline)
}
