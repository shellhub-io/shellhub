package storetest

import (
	"context"
	"sort"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTagCreate exercises TagCreate against the store under test.
func (s *Suite) TestTagCreate(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("succeeds when tag data is valid", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)

		tag := &models.Tag{
			Name:     "staging",
			TenantID: tenantID,
		}
		insertedID, err := st.TagCreate(ctx, tag)
		require.NoError(t, err)
		assert.NotEmpty(t, insertedID)
	})
}

// TestTagConflicts locks which tag fields are reported as already taken.
func (s *Suite) TestTagConflicts(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("no conflicts when target is empty", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		s.CreateTag(t, WithTagName("production"), WithTagTenant(tenantID))

		conflicts, has, err := st.TagConflicts(ctx, scope.MustBounded(tenantID), &models.TagConflicts{})
		require.NoError(t, err)
		assert.False(t, has)
		assert.Empty(t, conflicts)
	})

	t.Run("no conflicts with non existing name", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		s.CreateTag(t, WithTagName("production"), WithTagTenant(tenantID))

		conflicts, has, err := st.TagConflicts(ctx, scope.MustBounded(tenantID), &models.TagConflicts{Name: "nonexistent"})
		require.NoError(t, err)
		assert.False(t, has)
		assert.Empty(t, conflicts)
	})

	t.Run("no conflicts when namespace is different", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID1 := s.CreateNamespace(t)
		tenantID2 := s.CreateNamespace(t)
		s.CreateTag(t, WithTagName("production"), WithTagTenant(tenantID1))

		conflicts, has, err := st.TagConflicts(ctx, scope.MustBounded(tenantID2), &models.TagConflicts{Name: "production"})
		require.NoError(t, err)
		assert.False(t, has)
		assert.Empty(t, conflicts)
	})

	t.Run("conflict detected with existing name", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		s.CreateTag(t, WithTagName("production"), WithTagTenant(tenantID))

		conflicts, has, err := st.TagConflicts(ctx, scope.MustBounded(tenantID), &models.TagConflicts{Name: "production"})
		require.NoError(t, err)
		assert.True(t, has)
		assert.Equal(t, []string{"name"}, conflicts)
	})
}

// TestTagList exercises TagList against the store under test.
func (s *Suite) TestTagList(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	sortTags := func(tags []models.Tag) {
		sort.Slice(tags, func(i, j int) bool {
			return tags[i].Name < tags[j].Name
		})
	}

	t.Run("succeeds spanning namespaces under an explicit unbounded scope", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenant1 := s.CreateNamespace(t)
		tenant2 := s.CreateNamespace(t)
		s.CreateTag(t, WithTagName("production"), WithTagTenant(tenant1))
		s.CreateTag(t, WithTagName("staging"), WithTagTenant(tenant1))
		s.CreateTag(t, WithTagName("development"), WithTagTenant(tenant2))

		tags, count, err := st.TagList(ctx, scope.NewUnbounded("test: asserting that an unbounded scope really does span namespaces"))
		require.NoError(t, err)
		assert.Equal(t, 3, count)
		assert.Len(t, tags, 3)

		sortTags(tags)
		assert.Equal(t, "development", tags[0].Name)
		assert.Equal(t, "production", tags[1].Name)
		assert.Equal(t, "staging", tags[2].Name)
	})

	t.Run("succeeds when tenant filter applied", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenant1 := s.CreateNamespace(t)
		tenant2 := s.CreateNamespace(t)
		s.CreateTag(t, WithTagName("production"), WithTagTenant(tenant1))
		s.CreateTag(t, WithTagName("staging"), WithTagTenant(tenant1))
		s.CreateTag(t, WithTagName("development"), WithTagTenant(tenant2))

		tags, count, err := st.TagList(ctx, scope.MustBounded(tenant1))
		require.NoError(t, err)
		assert.Equal(t, 2, count)
		assert.Len(t, tags, 2)

		sortTags(tags)
		assert.Equal(t, "production", tags[0].Name)
		assert.Equal(t, "staging", tags[1].Name)
		assert.Equal(t, tenant1, tags[0].TenantID)
		assert.Equal(t, tenant1, tags[1].TenantID)
	})
}

// TestTagResolve exercises TagResolve against the store under test.
func (s *Suite) TestTagResolve(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when tag not found by ID", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		tagID := s.CreateTag(t, WithTagName("temp"), WithTagTenant(tenantID))
		err := st.TagDelete(ctx, &models.Tag{ID: tagID, TenantID: tenantID})
		require.NoError(t, err)

		tag, err := st.TagResolve(ctx, scope.NewUnbounded(reasonTestQueryMechanics), store.TagIDResolver, tagID)
		require.ErrorIs(t, err, store.ErrNoDocuments)
		assert.Nil(t, tag)
	})

	t.Run("succeeds resolving tag by ID", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		tagID := s.CreateTag(t, WithTagName("production"), WithTagTenant(tenantID))

		tag, err := st.TagResolve(ctx, scope.NewUnbounded(reasonTestQueryMechanics), store.TagIDResolver, tagID)
		require.NoError(t, err)
		require.NotNil(t, tag)
		assert.Equal(t, tagID, tag.ID)
		assert.Equal(t, "production", tag.Name)
		assert.Equal(t, tenantID, tag.TenantID)
	})

	t.Run("fails when tag not found by name", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		s.CreateTag(t, WithTagName("production"), WithTagTenant(tenantID))

		tag, err := st.TagResolve(ctx, scope.NewUnbounded(reasonTestQueryMechanics), store.TagNameResolver, "nonexistent")
		require.ErrorIs(t, err, store.ErrNoDocuments)
		assert.Nil(t, tag)
	})

	t.Run("succeeds resolving tag by name with tenant filter", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		tagID := s.CreateTag(t, WithTagName("production"), WithTagTenant(tenantID))

		tag, err := st.TagResolve(ctx, scope.MustBounded(tenantID), store.TagNameResolver, "production")
		require.NoError(t, err)
		require.NotNil(t, tag)
		assert.Equal(t, tagID, tag.ID)
		assert.Equal(t, "production", tag.Name)
		assert.Equal(t, tenantID, tag.TenantID)
	})
}

// TestTagUpdate exercises TagUpdate against the store under test.
func (s *Suite) TestTagUpdate(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when tag is not found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)

		tagID := s.CreateTag(t, WithTagName("temp"), WithTagTenant(tenantID))
		err := st.TagDelete(ctx, &models.Tag{ID: tagID, TenantID: tenantID})
		require.NoError(t, err)

		tag := &models.Tag{
			ID:       tagID,
			TenantID: tenantID,
			Name:     "edited-tag",
		}

		err = st.TagUpdate(ctx, tag)
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})

	t.Run("succeeds when tag is found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		tagID := s.CreateTag(t, WithTagName("production"), WithTagTenant(tenantID))

		tag := &models.Tag{
			ID:       tagID,
			TenantID: tenantID,
			Name:     "edited-tag",
		}

		err := st.TagUpdate(ctx, tag)
		require.NoError(t, err)

		updatedTag, err := st.TagResolve(ctx, scope.NewUnbounded(reasonTestQueryMechanics), store.TagIDResolver, tagID)
		require.NoError(t, err)
		assert.Equal(t, "edited-tag", updatedTag.Name)
	})
}

// TestTagPushToTarget locks attaching a tag to a device or key.
func (s *Suite) TestTagPushToTarget(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when tag does not exist", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		deviceUID := s.CreateDevice(t, WithTenantID(tenantID))

		tagID := s.CreateTag(t, WithTagName("temp"), WithTagTenant(tenantID))
		err := st.TagDelete(ctx, &models.Tag{ID: tagID, TenantID: tenantID})
		require.NoError(t, err)

		err = st.TagPushToTarget(ctx, tagID, store.TagTargetDevice, string(deviceUID))
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})

	t.Run("fails when device does not exist", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		tagID := s.CreateTag(t, WithTagName("staging"), WithTagTenant(tenantID))

		deviceUID := s.CreateDevice(t, WithTenantID(tenantID))
		device, err := st.DeviceResolve(ctx, scope.NewUnbounded(reasonTestQueryMechanics), store.DeviceUIDResolver, string(deviceUID))
		require.NoError(t, err)
		err = st.DeviceDelete(ctx, device)
		require.NoError(t, err)

		err = st.TagPushToTarget(ctx, tagID, store.TagTargetDevice, string(deviceUID))
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})

	t.Run("fails when public key does not exist", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		tagID := s.CreateTag(t, WithTagName("staging"), WithTagTenant(tenantID))

		err := st.TagPushToTarget(ctx, tagID, store.TagTargetPublicKey, "nonexistent-fingerprint")
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})

	t.Run("succeeds to push a tag to public key", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		fingerprint := s.CreatePublicKey(t, WithPublicKeyTenant(tenantID))
		tagID := s.CreateTag(t, WithTagName("staging"), WithTagTenant(tenantID))

		err := st.TagPushToTarget(ctx, tagID, store.TagTargetPublicKey, fingerprint)
		require.NoError(t, err)
	})

	t.Run("succeeds to push a tag to device", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		deviceUID := s.CreateDevice(t, WithTenantID(tenantID))
		tagID := s.CreateTag(t, WithTagName("staging"), WithTagTenant(tenantID))

		err := st.TagPushToTarget(ctx, tagID, store.TagTargetDevice, string(deviceUID))
		require.NoError(t, err)
	})
}

// TestTagPullFromTarget locks detaching a tag while leaving the tag itself in place.
func (s *Suite) TestTagPullFromTarget(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when tag does not exist", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		deviceUID := s.CreateDevice(t, WithTenantID(tenantID))

		tagID := s.CreateTag(t, WithTagName("temp"), WithTagTenant(tenantID))
		err := st.TagDelete(ctx, &models.Tag{ID: tagID, TenantID: tenantID})
		require.NoError(t, err)

		err = st.TagPullFromTarget(ctx, tagID, store.TagTargetDevice, string(deviceUID))
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})

	t.Run("fails when device does not exist", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		tagID := s.CreateTag(t, WithTagName("production"), WithTagTenant(tenantID))

		deviceUID := s.CreateDevice(t, WithTenantID(tenantID))
		device, err := st.DeviceResolve(ctx, scope.NewUnbounded(reasonTestQueryMechanics), store.DeviceUIDResolver, string(deviceUID))
		require.NoError(t, err)
		err = st.DeviceDelete(ctx, device)
		require.NoError(t, err)

		err = st.TagPullFromTarget(ctx, tagID, store.TagTargetDevice, string(deviceUID))
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})

	t.Run("succeeds to pull a tag from public key", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		fingerprint := s.CreatePublicKey(t, WithPublicKeyTenant(tenantID))
		tagID := s.CreateTag(t, WithTagName("production"), WithTagTenant(tenantID))

		err := st.TagPushToTarget(ctx, tagID, store.TagTargetPublicKey, fingerprint)
		require.NoError(t, err)

		err = st.TagPullFromTarget(ctx, tagID, store.TagTargetPublicKey, fingerprint)
		require.NoError(t, err)
	})

	t.Run("succeeds to pull tag from all public keys", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		fp1 := s.CreatePublicKey(t, WithPublicKeyTenant(tenantID))
		fp2 := s.CreatePublicKey(t, WithPublicKeyTenant(tenantID))
		tagID := s.CreateTag(t, WithTagName("production"), WithTagTenant(tenantID))

		err := st.TagPushToTarget(ctx, tagID, store.TagTargetPublicKey, fp1)
		require.NoError(t, err)
		err = st.TagPushToTarget(ctx, tagID, store.TagTargetPublicKey, fp2)
		require.NoError(t, err)

		err = st.TagPullFromTarget(ctx, tagID, store.TagTargetPublicKey)
		require.NoError(t, err)
	})

	t.Run("succeeds to pull a tag from device", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		deviceUID := s.CreateDevice(t, WithTenantID(tenantID))
		tagID := s.CreateTag(t, WithTagName("production"), WithTagTenant(tenantID))

		err := st.TagPushToTarget(ctx, tagID, store.TagTargetDevice, string(deviceUID))
		require.NoError(t, err)

		err = st.TagPullFromTarget(ctx, tagID, store.TagTargetDevice, string(deviceUID))
		require.NoError(t, err)
	})

	t.Run("succeeds to pull a tag from all targets when no specific targets provided", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		device1 := s.CreateDevice(t, WithTenantID(tenantID))
		device2 := s.CreateDevice(t, WithTenantID(tenantID))
		tagID := s.CreateTag(t, WithTagName("production"), WithTagTenant(tenantID))

		err := st.TagPushToTarget(ctx, tagID, store.TagTargetDevice, string(device1))
		require.NoError(t, err)
		err = st.TagPushToTarget(ctx, tagID, store.TagTargetDevice, string(device2))
		require.NoError(t, err)

		err = st.TagPullFromTarget(ctx, tagID, store.TagTargetDevice)
		require.NoError(t, err)
	})
}

// TestTagDelete exercises TagDelete against the store under test.
func (s *Suite) TestTagDelete(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when tag is not found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)

		tagID := s.CreateTag(t, WithTagName("temp"), WithTagTenant(tenantID))
		err := st.TagDelete(ctx, &models.Tag{ID: tagID, TenantID: tenantID})
		require.NoError(t, err)

		tag := &models.Tag{
			ID:       tagID,
			TenantID: tenantID,
		}

		err = st.TagDelete(ctx, tag)
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})

	t.Run("succeeds when tag is found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		tagID := s.CreateTag(t, WithTagName("production"), WithTagTenant(tenantID))

		tag := &models.Tag{
			ID:       tagID,
			TenantID: tenantID,
		}

		err := st.TagDelete(ctx, tag)
		require.NoError(t, err)

		_, err = st.TagResolve(ctx, scope.NewUnbounded(reasonTestQueryMechanics), store.TagIDResolver, tagID)
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})
}
