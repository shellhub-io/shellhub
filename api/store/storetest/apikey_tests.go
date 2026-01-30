package storetest

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *Suite) TestAPIKeyCreate(t *testing.T) {
	t.Run("succeeds", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		keyID := s.CreateAPIKey(t, WithAPIKeyName("dev"))
		assert.NotEmpty(t, keyID)
	})
}

func (s *Suite) TestAPIKeyConflicts(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("no conflicts when target is empty", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		s.CreateAPIKey(t, WithAPIKeyName("dev"), WithAPIKeyTenant(tenantID))

		conflicts, has, err := st.APIKeyConflicts(ctx, tenantID, &models.APIKeyConflicts{})
		require.NoError(t, err)
		assert.False(t, has)
		assert.Empty(t, conflicts)
	})

	t.Run("no conflicts with non existing name", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		s.CreateAPIKey(t, WithAPIKeyName("dev"), WithAPIKeyTenant(tenantID))

		conflicts, has, err := st.APIKeyConflicts(ctx, tenantID, &models.APIKeyConflicts{Name: "nonexistent"})
		require.NoError(t, err)
		assert.False(t, has)
		assert.Empty(t, conflicts)
	})

	t.Run("no conflict detected with existing attribute but different tenant id", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenant1 := s.CreateNamespace(t)
		tenant2 := s.CreateNamespace(t)
		s.CreateAPIKey(t, WithAPIKeyName("dev"), WithAPIKeyTenant(tenant1))

		conflicts, has, err := st.APIKeyConflicts(ctx, tenant2, &models.APIKeyConflicts{Name: "dev"})
		require.NoError(t, err)
		assert.False(t, has)
		assert.Empty(t, conflicts)
	})

	t.Run("conflict detected with existing name", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		s.CreateAPIKey(t, WithAPIKeyName("dev"), WithAPIKeyTenant(tenantID))

		conflicts, has, err := st.APIKeyConflicts(ctx, tenantID, &models.APIKeyConflicts{Name: "dev"})
		require.NoError(t, err)
		assert.True(t, has)
		assert.Equal(t, []string{"name"}, conflicts)
	})

	t.Run("conflict detected with existing id", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		keyID := s.CreateAPIKey(t, WithAPIKeyName("dev"), WithAPIKeyTenant(tenantID))

		conflicts, has, err := st.APIKeyConflicts(ctx, tenantID, &models.APIKeyConflicts{ID: keyID})
		require.NoError(t, err)
		assert.True(t, has)
		assert.Equal(t, []string{"id"}, conflicts)
	})
}

func (s *Suite) TestAPIKeyResolve(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when ID does not exist", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)

		apiKey, err := st.APIKeyResolve(ctx, store.APIKeyIDResolver, "nonexistent-id", st.Options().InNamespace(tenantID))
		assert.ErrorIs(t, err, store.ErrNoDocuments)
		assert.Nil(t, apiKey)
	})

	t.Run("succeeds resolving API key by ID", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		keyID := s.CreateAPIKey(t, WithAPIKeyName("dev"), WithAPIKeyTenant(tenantID), WithAPIKeyRole("administrator"))

		apiKey, err := st.APIKeyResolve(ctx, store.APIKeyIDResolver, keyID, st.Options().InNamespace(tenantID))
		require.NoError(t, err)
		require.NotNil(t, apiKey)
		assert.Equal(t, keyID, apiKey.ID)
		assert.Equal(t, "dev", apiKey.Name)
		assert.Equal(t, tenantID, apiKey.TenantID)
		assert.Equal(t, "administrator", string(apiKey.Role))
	})

	t.Run("fails when name and tenant ID does not exist", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)

		apiKey, err := st.APIKeyResolve(ctx, store.APIKeyNameResolver, "nonexistent", st.Options().InNamespace(tenantID))
		assert.ErrorIs(t, err, store.ErrNoDocuments)
		assert.Nil(t, apiKey)
	})

	t.Run("succeeds resolving API key by name", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		keyID := s.CreateAPIKey(t, WithAPIKeyName("dev"), WithAPIKeyTenant(tenantID), WithAPIKeyRole("administrator"))

		apiKey, err := st.APIKeyResolve(ctx, store.APIKeyNameResolver, "dev", st.Options().InNamespace(tenantID))
		require.NoError(t, err)
		require.NotNil(t, apiKey)
		assert.Equal(t, keyID, apiKey.ID)
		assert.Equal(t, "dev", apiKey.Name)
		assert.Equal(t, tenantID, apiKey.TenantID)
		assert.Equal(t, "administrator", string(apiKey.Role))
	})

	t.Run("fails when API key exists but belongs to different tenant", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenant1 := s.CreateNamespace(t)
		tenant2 := s.CreateNamespace(t)
		s.CreateAPIKey(t, WithAPIKeyName("dev"), WithAPIKeyTenant(tenant1))

		apiKey, err := st.APIKeyResolve(ctx, store.APIKeyNameResolver, "dev", st.Options().InNamespace(tenant2))
		assert.ErrorIs(t, err, store.ErrNoDocuments)
		assert.Nil(t, apiKey)
	})
}

func (s *Suite) TestAPIKeyList(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("succeeds when there are no api keys", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)

		apiKeys, count, err := st.APIKeyList(ctx,
			st.Options().InNamespace(tenantID),
			st.Options().Sort(&query.Sorter{By: "expires_in", Order: query.OrderAsc}),
			st.Options().Paginate(&query.Paginator{Page: 1, PerPage: 10}),
		)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
		assert.Empty(t, apiKeys)
	})

	t.Run("succeeds when there are api keys", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		s.CreateAPIKey(t, WithAPIKeyName("key1"), WithAPIKeyTenant(tenantID))
		s.CreateAPIKey(t, WithAPIKeyName("key2"), WithAPIKeyTenant(tenantID))

		apiKeys, count, err := st.APIKeyList(ctx,
			st.Options().InNamespace(tenantID),
			st.Options().Sort(&query.Sorter{By: "expires_in", Order: query.OrderAsc}),
			st.Options().Paginate(&query.Paginator{Page: 1, PerPage: 10}),
		)
		require.NoError(t, err)
		assert.Equal(t, 2, count)
		assert.Len(t, apiKeys, 2)
	})

	t.Run("succeeds when there are api keys and pagination", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		s.CreateAPIKey(t, WithAPIKeyName("key1"), WithAPIKeyTenant(tenantID))
		s.CreateAPIKey(t, WithAPIKeyName("key2"), WithAPIKeyTenant(tenantID))

		apiKeys, count, err := st.APIKeyList(ctx,
			st.Options().InNamespace(tenantID),
			st.Options().Sort(&query.Sorter{By: "expires_in", Order: query.OrderAsc}),
			st.Options().Paginate(&query.Paginator{Page: 1, PerPage: 1}),
		)
		require.NoError(t, err)
		assert.Equal(t, 2, count) // Total count
		assert.Len(t, apiKeys, 1) // Page 1 with perPage=1 returns 1 item
	})
}

func (s *Suite) TestAPIKeyUpdate(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when API key does not exist", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)

		// Create and delete an API key to get a valid but non-existent ID
		keyID := s.CreateAPIKey(t, WithAPIKeyName("temp"), WithAPIKeyTenant(tenantID))
		apiKey, err := st.APIKeyResolve(ctx, store.APIKeyIDResolver, keyID, st.Options().InNamespace(tenantID))
		require.NoError(t, err)
		err = st.APIKeyDelete(ctx, apiKey)
		require.NoError(t, err)

		// Try to update the deleted API key
		apiKey.Name = "updated"
		err = st.APIKeyUpdate(ctx, apiKey)
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})

	t.Run("succeeds when API key exists", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		keyID := s.CreateAPIKey(t, WithAPIKeyName("dev"), WithAPIKeyTenant(tenantID), WithAPIKeyRole("administrator"))

		// Get the full API key
		apiKey, err := st.APIKeyResolve(ctx, store.APIKeyIDResolver, keyID, st.Options().InNamespace(tenantID))
		require.NoError(t, err)

		// Update name
		apiKey.Name = "updated-dev"
		err = st.APIKeyUpdate(ctx, apiKey)
		require.NoError(t, err)

		// Verify update
		updatedKey, err := st.APIKeyResolve(ctx, store.APIKeyIDResolver, keyID, st.Options().InNamespace(tenantID))
		require.NoError(t, err)
		assert.Equal(t, "updated-dev", updatedKey.Name)
	})
}

func (s *Suite) TestAPIKeyDelete(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when API key does not exist", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)

		// Create and delete an API key to get a valid but non-existent ID
		keyID := s.CreateAPIKey(t, WithAPIKeyName("temp"), WithAPIKeyTenant(tenantID))
		apiKey, err := st.APIKeyResolve(ctx, store.APIKeyIDResolver, keyID, st.Options().InNamespace(tenantID))
		require.NoError(t, err)
		err = st.APIKeyDelete(ctx, apiKey)
		require.NoError(t, err)

		// Try to delete again
		err = st.APIKeyDelete(ctx, apiKey)
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})

	t.Run("succeeds when API key exists", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		keyID := s.CreateAPIKey(t, WithAPIKeyName("dev"), WithAPIKeyTenant(tenantID))

		// Get the full API key
		apiKey, err := st.APIKeyResolve(ctx, store.APIKeyIDResolver, keyID, st.Options().InNamespace(tenantID))
		require.NoError(t, err)

		// Delete
		err = st.APIKeyDelete(ctx, apiKey)
		require.NoError(t, err)

		// Verify deletion
		_, err = st.APIKeyResolve(ctx, store.APIKeyIDResolver, keyID, st.Options().InNamespace(tenantID))
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})
}
