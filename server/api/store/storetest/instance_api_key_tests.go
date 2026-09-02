package storetest

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInstanceAPIKeyCreate exercises InstanceAPIKeyCreate against the store under test.
func (s *Suite) TestInstanceAPIKeyCreate(t *testing.T) {
	t.Run("succeeds when data is valid", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		digest := s.CreateInstanceAPIKey(t)
		assert.NotEmpty(t, digest)
	})

	t.Run("fails when the name is already taken", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		ctx := context.Background()
		st := s.provider.Store()

		s.CreateInstanceAPIKey(t, WithInstanceAPIKeyName("billing-export"))

		duplicated := &models.InstanceAPIKey{
			ID:        "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			Name:      "billing-export",
			CreatedBy: s.CreateUser(t),
			ExpiresAt: clock.Now().AddDate(0, 0, 30),
		}

		_, err := st.InstanceAPIKeyCreate(ctx, duplicated)
		require.ErrorIs(t, err, store.ErrDuplicate)
	})
}

// TestInstanceAPIKeyResolve exercises InstanceAPIKeyResolve against the store under test.
func (s *Suite) TestInstanceAPIKeyResolve(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when the key is not found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		apiKey, err := st.InstanceAPIKeyResolve(ctx, store.InstanceAPIKeyIDResolver, "nonexistent")
		require.ErrorIs(t, err, store.ErrNoDocuments)
		assert.Nil(t, apiKey)
	})

	t.Run("fails when the resolver is unknown", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		apiKey, err := st.InstanceAPIKeyResolve(ctx, store.InstanceAPIKeyResolver(0), "whatever")
		require.ErrorIs(t, err, store.ErrResolverNotFound)
		assert.Nil(t, apiKey)
	})

	t.Run("succeeds when resolving by digest", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		userID := s.CreateUser(t)
		expiresAt := clock.Now().AddDate(0, 0, 90).UTC().Truncate(time.Second)
		digest := s.CreateInstanceAPIKey(t,
			WithInstanceAPIKeyName("license-sync"),
			WithInstanceAPIKeyCreatedBy(userID),
			WithInstanceAPIKeyExpiresAt(expiresAt),
		)

		apiKey, err := st.InstanceAPIKeyResolve(ctx, store.InstanceAPIKeyIDResolver, digest)
		require.NoError(t, err)
		require.NotNil(t, apiKey)
		assert.Equal(t, digest, apiKey.ID)
		assert.Equal(t, "license-sync", apiKey.Name)
		assert.Equal(t, userID, apiKey.CreatedBy)
		assert.WithinDuration(t, expiresAt, apiKey.ExpiresAt, time.Second)
	})

	t.Run("succeeds when resolving by name", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		digest := s.CreateInstanceAPIKey(t, WithInstanceAPIKeyName("user-export"))

		apiKey, err := st.InstanceAPIKeyResolve(ctx, store.InstanceAPIKeyNameResolver, "user-export")
		require.NoError(t, err)
		require.NotNil(t, apiKey)
		assert.Equal(t, digest, apiKey.ID)
	})
}

// TestInstanceAPIKeyList exercises InstanceAPIKeyList against the store under test.
func (s *Suite) TestInstanceAPIKeyList(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("succeeds when there are no keys", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		apiKeys, count, err := st.InstanceAPIKeyList(ctx)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
		assert.Empty(t, apiKeys)
	})

	t.Run("succeeds when keys exist", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		s.CreateInstanceAPIKey(t, WithInstanceAPIKeyName("first"))
		s.CreateInstanceAPIKey(t, WithInstanceAPIKeyName("second"))

		apiKeys, count, err := st.InstanceAPIKeyList(ctx)
		require.NoError(t, err)
		assert.Equal(t, 2, count)
		require.Len(t, apiKeys, 2)

		names := []string{apiKeys[0].Name, apiKeys[1].Name}
		assert.ElementsMatch(t, []string{"first", "second"}, names)
	})
}

// TestInstanceAPIKeyDelete exercises InstanceAPIKeyDelete against the store under test.
func (s *Suite) TestInstanceAPIKeyDelete(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when the key is not found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		require.ErrorIs(t, st.InstanceAPIKeyDelete(ctx, "nonexistent"), store.ErrNoDocuments)
	})

	t.Run("succeeds when the key exists", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		s.CreateInstanceAPIKey(t, WithInstanceAPIKeyName("retired"))

		require.NoError(t, st.InstanceAPIKeyDelete(ctx, "retired"))

		_, err := st.InstanceAPIKeyResolve(ctx, store.InstanceAPIKeyNameResolver, "retired")
		require.ErrorIs(t, err, store.ErrNoDocuments)
	})
}
