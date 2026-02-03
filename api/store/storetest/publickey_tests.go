package storetest

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *Suite) TestPublicKeyResolve(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when public key is not found due to fingerprint", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)

		pubKey, err := st.PublicKeyResolve(ctx, store.PublicKeyFingerprintResolver, "nonexistent-fingerprint", st.Options().InNamespace(tenantID))
		assert.ErrorIs(t, err, store.ErrNoDocuments)
		assert.Nil(t, pubKey)
	})

	t.Run("fails when public key is not found due to tenant", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenant1 := s.CreateNamespace(t)
		tenant2 := s.CreateNamespace(t)
		fingerprint := s.CreatePublicKey(t, WithPublicKeyName("key1"), WithPublicKeyTenant(tenant1))

		pubKey, err := st.PublicKeyResolve(ctx, store.PublicKeyFingerprintResolver, fingerprint, st.Options().InNamespace(tenant2))
		assert.ErrorIs(t, err, store.ErrNoDocuments)
		assert.Nil(t, pubKey)
	})

	t.Run("succeeds when public key is found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		fingerprint := s.CreatePublicKey(t,
			WithPublicKeyName("public_key"),
			WithPublicKeyTenant(tenantID),
			WithPublicKeyHostname(".*"),
		)

		pubKey, err := st.PublicKeyResolve(ctx, store.PublicKeyFingerprintResolver, fingerprint, st.Options().InNamespace(tenantID))
		require.NoError(t, err)
		require.NotNil(t, pubKey)
		assert.Equal(t, fingerprint, pubKey.Fingerprint)
		assert.Equal(t, tenantID, pubKey.TenantID)
		assert.Equal(t, "public_key", pubKey.Name)
		assert.Equal(t, ".*", pubKey.Filter.Hostname)
	})
}

func (s *Suite) TestPublicKeyList(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("succeeds when public key list is empty", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		pubKeys, count, err := st.PublicKeyList(ctx, st.Options().Paginate(&query.Paginator{Page: -1, PerPage: -1}))
		require.NoError(t, err)
		assert.Equal(t, 0, count)
		assert.Empty(t, pubKeys)
	})

	t.Run("succeeds when public key list len is greater than 1", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		s.CreatePublicKey(t, WithPublicKeyName("key1"))

		pubKeys, count, err := st.PublicKeyList(ctx, st.Options().Paginate(&query.Paginator{Page: -1, PerPage: -1}))
		require.NoError(t, err)
		assert.Equal(t, 1, count)
		assert.Len(t, pubKeys, 1)
	})
}

func (s *Suite) TestPublicKeyCreate(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("succeeds when data is valid", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		fingerprint := s.CreatePublicKey(t, WithPublicKeyName("public_key"))
		assert.NotEmpty(t, fingerprint)
	})

	t.Run("succeeds with tag filters", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create tenant and tags first
		tenantID := s.CreateNamespace(t)
		tagID1 := s.CreateTag(t, WithTagName("tag1"), WithTagTenant(tenantID))
		tagID2 := s.CreateTag(t, WithTagName("tag2"), WithTagTenant(tenantID))

		// Create public key with tag filters
		fingerprint := s.CreatePublicKey(t,
			WithPublicKeyName("key-with-tags"),
			WithPublicKeyTenant(tenantID),
			WithPublicKeyTags([]string{tagID1, tagID2}),
		)
		assert.NotEmpty(t, fingerprint)

		// Verify the key was created with tags
		pubKey, err := st.PublicKeyResolve(ctx, store.PublicKeyFingerprintResolver, fingerprint, st.Options().InNamespace(tenantID))
		require.NoError(t, err)
		assert.Len(t, pubKey.Filter.TagIDs, 2)
	})

	t.Run("succeeds with empty tag filter", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create public key with explicitly empty tags
		fingerprint := s.CreatePublicKey(t,
			WithPublicKeyName("key-no-tags"),
			WithPublicKeyTags([]string{}),
		)
		assert.NotEmpty(t, fingerprint)
	})

	t.Run("succeeds with single tag", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create tenant and a single tag
		tenantID := s.CreateNamespace(t)
		tagID := s.CreateTag(t, WithTagName("single-tag"), WithTagTenant(tenantID))

		// Create public key with single tag filter
		fingerprint := s.CreatePublicKey(t,
			WithPublicKeyName("key-with-single-tag"),
			WithPublicKeyTenant(tenantID),
			WithPublicKeyTags([]string{tagID}),
		)
		assert.NotEmpty(t, fingerprint)

		// Verify the key was created
		pubKey, err := st.PublicKeyResolve(ctx, store.PublicKeyFingerprintResolver, fingerprint, st.Options().InNamespace(tenantID))
		require.NoError(t, err)
		assert.Len(t, pubKey.Filter.TagIDs, 1)
		assert.Contains(t, pubKey.Filter.TagIDs, tagID)
	})
}

func (s *Suite) TestPublicKeyUpdate(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when public key is not found due to fingerprint", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)

		// Create and delete a public key to get a valid but non-existent fingerprint
		fingerprint := s.CreatePublicKey(t, WithPublicKeyName("temp"), WithPublicKeyTenant(tenantID))
		pubKey, err := st.PublicKeyResolve(ctx, store.PublicKeyFingerprintResolver, fingerprint, st.Options().InNamespace(tenantID))
		require.NoError(t, err)
		err = st.PublicKeyDelete(ctx, pubKey)
		require.NoError(t, err)

		// Try to update the deleted public key
		pubKey.Name = "edited_name"
		err = st.PublicKeyUpdate(ctx, pubKey)
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})

	t.Run("fails when public key is not found due to tenant", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenant1 := s.CreateNamespace(t)
		tenant2 := s.CreateNamespace(t)
		fingerprint := s.CreatePublicKey(t, WithPublicKeyName("key1"), WithPublicKeyTenant(tenant1))

		// Get the public key from tenant1
		pubKey, err := st.PublicKeyResolve(ctx, store.PublicKeyFingerprintResolver, fingerprint, st.Options().InNamespace(tenant1))
		require.NoError(t, err)

		// Try to update it with tenant2 ID (should fail)
		pubKey.TenantID = tenant2
		pubKey.Name = "edited_name"
		err = st.PublicKeyUpdate(ctx, pubKey)
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})

	t.Run("succeeds when public key is found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		tag1 := s.CreateTag(t, WithTagName("tag1"), WithTagTenant(tenantID))
		tag2 := s.CreateTag(t, WithTagName("tag2"), WithTagTenant(tenantID))
		fingerprint := s.CreatePublicKey(t, WithPublicKeyName("key1"), WithPublicKeyTenant(tenantID))

		// Get the full public key
		pubKey, err := st.PublicKeyResolve(ctx, store.PublicKeyFingerprintResolver, fingerprint, st.Options().InNamespace(tenantID))
		require.NoError(t, err)

		// Update name and tags
		pubKey.Name = "edited_key"
		pubKey.Filter.TagIDs = []string{tag1, tag2}
		err = st.PublicKeyUpdate(ctx, pubKey)
		require.NoError(t, err)

		// Verify update
		updatedKey, err := st.PublicKeyResolve(ctx, store.PublicKeyFingerprintResolver, fingerprint, st.Options().InNamespace(tenantID))
		require.NoError(t, err)
		assert.Equal(t, "edited_key", updatedKey.Name)
	})
}

func (s *Suite) TestPublicKeyDelete(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when public key is not found due to fingerprint", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)

		// Create and delete a public key to get a valid but non-existent fingerprint
		fingerprint := s.CreatePublicKey(t, WithPublicKeyName("temp"), WithPublicKeyTenant(tenantID))
		pubKey, err := st.PublicKeyResolve(ctx, store.PublicKeyFingerprintResolver, fingerprint, st.Options().InNamespace(tenantID))
		require.NoError(t, err)
		err = st.PublicKeyDelete(ctx, pubKey)
		require.NoError(t, err)

		// Try to delete again
		err = st.PublicKeyDelete(ctx, pubKey)
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})

	t.Run("fails when public key is not found due to tenant", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenant1 := s.CreateNamespace(t)
		tenant2 := s.CreateNamespace(t)
		fingerprint := s.CreatePublicKey(t, WithPublicKeyName("key1"), WithPublicKeyTenant(tenant1))

		// Get the public key from tenant1
		pubKey, err := st.PublicKeyResolve(ctx, store.PublicKeyFingerprintResolver, fingerprint, st.Options().InNamespace(tenant1))
		require.NoError(t, err)

		// Try to delete with wrong tenant ID
		pubKey.TenantID = tenant2
		err = st.PublicKeyDelete(ctx, pubKey)
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})

	t.Run("succeeds when public key is found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		fingerprint := s.CreatePublicKey(t, WithPublicKeyName("key1"), WithPublicKeyTenant(tenantID))

		// Get the full public key
		pubKey, err := st.PublicKeyResolve(ctx, store.PublicKeyFingerprintResolver, fingerprint, st.Options().InNamespace(tenantID))
		require.NoError(t, err)

		// Delete
		err = st.PublicKeyDelete(ctx, pubKey)
		require.NoError(t, err)

		// Verify deletion
		_, err = st.PublicKeyResolve(ctx, store.PublicKeyFingerprintResolver, fingerprint, st.Options().InNamespace(tenantID))
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})
}
