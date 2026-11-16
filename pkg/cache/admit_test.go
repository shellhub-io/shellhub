package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	cacheMock "github.com/shellhub-io/shellhub/pkg/cache/mocks"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAdmitOne(t *testing.T) {
	const (
		key   = "guess-in-flight={user}"
		owner = "11111111-1111-4111-1111-111111111111"
		ttl   = 10 * time.Second
	)

	uncancelled := mock.MatchedBy(func(ctx context.Context) bool { return ctx.Err() == nil })

	t.Run("admits a caller and releases only its own reservation, even after ctx is cancelled", func(t *testing.T) {
		uuid.Fix(t, owner)
		c := cacheMock.NewMockCache(t)
		ctx, cancel := context.WithCancel(context.Background())

		c.On("SetNX", ctx, key, owner, ttl).Return(true, nil).Once()
		c.On("CompareAndDelete", uncancelled, key, owner).Return(true, nil).Once()

		release, admitted, err := AdmitOne(ctx, c, key, ttl)
		require.NoError(t, err)
		assert.True(t, admitted)

		cancel()
		release()
	})

	t.Run("refuses a caller while another holds the reservation", func(t *testing.T) {
		uuid.Fix(t, owner)
		c := cacheMock.NewMockCache(t)
		ctx := context.Background()

		c.On("SetNX", ctx, key, owner, ttl).Return(false, nil).Once()

		release, admitted, err := AdmitOne(ctx, c, key, ttl)
		require.NoError(t, err)
		assert.False(t, admitted)

		release()
	})

	t.Run("reports a cache failure without admitting", func(t *testing.T) {
		uuid.Fix(t, owner)
		c := cacheMock.NewMockCache(t)
		ctx := context.Background()
		failure := errors.New("redis down")

		c.On("SetNX", ctx, key, owner, ttl).Return(false, failure).Once()

		release, admitted, err := AdmitOne(ctx, c, key, ttl)
		require.ErrorIs(t, err, failure)
		assert.False(t, admitted)

		release()
	})
}
