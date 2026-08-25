package storetest

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
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

		conflicts, has, err := st.APIKeyConflicts(ctx, scope.MustBounded(tenantID), &models.APIKeyConflicts{})
		require.NoError(t, err)
		assert.False(t, has)
		assert.Empty(t, conflicts)
	})

	t.Run("no conflicts with non existing name", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		s.CreateAPIKey(t, WithAPIKeyName("dev"), WithAPIKeyTenant(tenantID))

		conflicts, has, err := st.APIKeyConflicts(ctx, scope.MustBounded(tenantID), &models.APIKeyConflicts{Name: "nonexistent"})
		require.NoError(t, err)
		assert.False(t, has)
		assert.Empty(t, conflicts)
	})

	t.Run("no conflict detected with existing attribute but different tenant id", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenant1 := s.CreateNamespace(t)
		tenant2 := s.CreateNamespace(t)
		s.CreateAPIKey(t, WithAPIKeyName("dev"), WithAPIKeyTenant(tenant1))

		conflicts, has, err := st.APIKeyConflicts(ctx, scope.MustBounded(tenant2), &models.APIKeyConflicts{Name: "dev"})
		require.NoError(t, err)
		assert.False(t, has)
		assert.Empty(t, conflicts)
	})

	t.Run("conflict detected with existing name", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		s.CreateAPIKey(t, WithAPIKeyName("dev"), WithAPIKeyTenant(tenantID))

		conflicts, has, err := st.APIKeyConflicts(ctx, scope.MustBounded(tenantID), &models.APIKeyConflicts{Name: "dev"})
		require.NoError(t, err)
		assert.True(t, has)
		assert.Equal(t, []string{"name"}, conflicts)
	})

	t.Run("conflict detected with existing id", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		keyID := s.CreateAPIKey(t, WithAPIKeyName("dev"), WithAPIKeyTenant(tenantID))

		conflicts, has, err := st.APIKeyConflicts(ctx, scope.MustBounded(tenantID), &models.APIKeyConflicts{ID: keyID})
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

		apiKey, err := st.APIKeyResolve(ctx, scope.MustBounded(tenantID), store.APIKeyIDResolver, "nonexistent-id")
		assert.ErrorIs(t, err, store.ErrNoDocuments)
		assert.Nil(t, apiKey)
	})

	// namespace_id is uuid-typed; a malformed tenant ID must return ErrNoDocuments
	// without reaching the database (avoids SQLSTATE 22P02 on Postgres).
	t.Run("fails with malformed (non-UUID) namespace ID", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// 'l' in position 13 makes this an invalid UUID (not a hex digit).
		malformedTenantID := "83176492-e6cl-43d7-922e-ee01c3693e26"

		apiKey, err := st.APIKeyResolve(ctx, scope.MustBounded(malformedTenantID), store.APIKeyIDResolver, "any-key-id")
		assert.ErrorIs(t, err, store.ErrNoDocuments)
		assert.Nil(t, apiKey)
	})

	t.Run("succeeds resolving API key by ID", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		keyID := s.CreateAPIKey(t, WithAPIKeyName("dev"), WithAPIKeyTenant(tenantID), WithAPIKeyRole("administrator"))

		apiKey, err := st.APIKeyResolve(ctx, scope.MustBounded(tenantID), store.APIKeyIDResolver, keyID)
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

		apiKey, err := st.APIKeyResolve(ctx, scope.MustBounded(tenantID), store.APIKeyNameResolver, "nonexistent")
		assert.ErrorIs(t, err, store.ErrNoDocuments)
		assert.Nil(t, apiKey)
	})

	t.Run("succeeds resolving API key by name", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		keyID := s.CreateAPIKey(t, WithAPIKeyName("dev"), WithAPIKeyTenant(tenantID), WithAPIKeyRole("administrator"))

		apiKey, err := st.APIKeyResolve(ctx, scope.MustBounded(tenantID), store.APIKeyNameResolver, "dev")
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

		apiKey, err := st.APIKeyResolve(ctx, scope.MustBounded(tenant2), store.APIKeyNameResolver, "dev")
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

		apiKeys, count, err := st.APIKeyList(ctx, scope.MustBounded(tenantID),
			st.Options().Sort(&query.Sorter{By: "expires_in", Order: query.OrderAsc}),
			st.Options().Paginate(&query.Paginator{Page: 1, PerPage: 10}))
		require.NoError(t, err)
		assert.Equal(t, 0, count)
		assert.Empty(t, apiKeys)
	})

	t.Run("succeeds when there are api keys", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		s.CreateAPIKey(t, WithAPIKeyName("key1"), WithAPIKeyTenant(tenantID))
		s.CreateAPIKey(t, WithAPIKeyName("key2"), WithAPIKeyTenant(tenantID))

		apiKeys, count, err := st.APIKeyList(ctx, scope.MustBounded(tenantID),
			st.Options().Sort(&query.Sorter{By: "expires_in", Order: query.OrderAsc}),
			st.Options().Paginate(&query.Paginator{Page: 1, PerPage: 10}))
		require.NoError(t, err)
		assert.Equal(t, 2, count)
		assert.Len(t, apiKeys, 2)
	})

	t.Run("succeeds when there are api keys and pagination", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		s.CreateAPIKey(t, WithAPIKeyName("key1"), WithAPIKeyTenant(tenantID))
		s.CreateAPIKey(t, WithAPIKeyName("key2"), WithAPIKeyTenant(tenantID))

		apiKeys, count, err := st.APIKeyList(ctx, scope.MustBounded(tenantID),
			st.Options().Sort(&query.Sorter{By: "expires_in", Order: query.OrderAsc}),
			st.Options().Paginate(&query.Paginator{Page: 1, PerPage: 1}))
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
		apiKey, err := st.APIKeyResolve(ctx, scope.MustBounded(tenantID), store.APIKeyIDResolver, keyID)
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
		apiKey, err := st.APIKeyResolve(ctx, scope.MustBounded(tenantID), store.APIKeyIDResolver, keyID)
		require.NoError(t, err)

		apiKey.Name = "updated-dev"
		err = st.APIKeyUpdate(ctx, apiKey)
		require.NoError(t, err)

		// Verify update
		updatedKey, err := st.APIKeyResolve(ctx, scope.MustBounded(tenantID), store.APIKeyIDResolver, keyID)
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
		apiKey, err := st.APIKeyResolve(ctx, scope.MustBounded(tenantID), store.APIKeyIDResolver, keyID)
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
		apiKey, err := st.APIKeyResolve(ctx, scope.MustBounded(tenantID), store.APIKeyIDResolver, keyID)
		require.NoError(t, err)

		err = st.APIKeyDelete(ctx, apiKey)
		require.NoError(t, err)

		// Verify deletion
		_, err = st.APIKeyResolve(ctx, scope.MustBounded(tenantID), store.APIKeyIDResolver, keyID)
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})
}

func (s *Suite) TestAPIKeyDeleteAllByCreator(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("succeeds and is a no-op when the creator has no keys", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)

		err := st.APIKeyDeleteAllByCreator(ctx, tenantID, s.CreateUser(t))
		require.NoError(t, err)
	})

	t.Run("deletes only the creator's keys within the tenant", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		otherTenantID := s.CreateNamespace(t)
		creator := s.CreateUser(t)
		other := s.CreateUser(t)

		s.CreateAPIKey(t, WithAPIKeyName("creator-1"), WithAPIKeyTenant(tenantID), WithAPIKeyCreatedBy(creator))
		s.CreateAPIKey(t, WithAPIKeyName("creator-2"), WithAPIKeyTenant(tenantID), WithAPIKeyCreatedBy(creator))
		s.CreateAPIKey(t, WithAPIKeyName("other-user"), WithAPIKeyTenant(tenantID), WithAPIKeyCreatedBy(other))
		s.CreateAPIKey(t, WithAPIKeyName("other-tenant"), WithAPIKeyTenant(otherTenantID), WithAPIKeyCreatedBy(creator))

		err := st.APIKeyDeleteAllByCreator(ctx, tenantID, creator)
		require.NoError(t, err)

		remaining, count, err := st.APIKeyList(ctx, scope.MustBounded(tenantID),
			st.Options().Sort(&query.Sorter{By: "expires_in", Order: query.OrderAsc}),
			st.Options().Paginate(&query.Paginator{Page: 1, PerPage: 10}))
		require.NoError(t, err)
		assert.Equal(t, 1, count)
		require.Len(t, remaining, 1)
		assert.Equal(t, "other-user", remaining[0].Name)

		// The same creator's key in another tenant must be untouched.
		otherRemaining, otherCount, err := st.APIKeyList(ctx, scope.MustBounded(otherTenantID),
			st.Options().Sort(&query.Sorter{By: "expires_in", Order: query.OrderAsc}),
			st.Options().Paginate(&query.Paginator{Page: 1, PerPage: 10}))
		require.NoError(t, err)
		assert.Equal(t, 1, otherCount)
		require.Len(t, otherRemaining, 1)
		assert.Equal(t, "other-tenant", otherRemaining[0].Name)
	})
}
