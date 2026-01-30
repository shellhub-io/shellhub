package storetest

import (
	"context"
	"sort"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *Suite) TestTagCreate(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("succeeds when tag data is valid", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create namespace first
		tenantID := s.CreateNamespace(t)

		// Create tag
		tag := &models.Tag{
			Name:     "staging",
			TenantID: tenantID,
		}
		insertedID, err := st.TagCreate(ctx, tag)
		require.NoError(t, err)
		assert.NotEmpty(t, insertedID)
	})
}

func (s *Suite) TestTagConflicts(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("no conflicts when target is empty", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		s.CreateTag(t, WithTagName("production"), WithTagTenant(tenantID))

		conflicts, has, err := st.TagConflicts(ctx, tenantID, &models.TagConflicts{})
		require.NoError(t, err)
		assert.False(t, has)
		assert.Empty(t, conflicts)
	})

	t.Run("no conflicts with non existing name", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		s.CreateTag(t, WithTagName("production"), WithTagTenant(tenantID))

		conflicts, has, err := st.TagConflicts(ctx, tenantID, &models.TagConflicts{Name: "nonexistent"})
		require.NoError(t, err)
		assert.False(t, has)
		assert.Empty(t, conflicts)
	})

	t.Run("no conflicts when namespace is different", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID1 := s.CreateNamespace(t)
		tenantID2 := s.CreateNamespace(t)
		s.CreateTag(t, WithTagName("production"), WithTagTenant(tenantID1))

		conflicts, has, err := st.TagConflicts(ctx, tenantID2, &models.TagConflicts{Name: "production"})
		require.NoError(t, err)
		assert.False(t, has)
		assert.Empty(t, conflicts)
	})

	t.Run("conflict detected with existing name", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		s.CreateTag(t, WithTagName("production"), WithTagTenant(tenantID))

		conflicts, has, err := st.TagConflicts(ctx, tenantID, &models.TagConflicts{Name: "production"})
		require.NoError(t, err)
		assert.True(t, has)
		assert.Equal(t, []string{"name"}, conflicts)
	})
}

func (s *Suite) TestTagList(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	sortTags := func(tags []models.Tag) {
		sort.Slice(tags, func(i, j int) bool {
			return tags[i].Name < tags[j].Name
		})
	}

	t.Run("succeeds when no filters applied", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create tags in different namespaces
		tenant1 := s.CreateNamespace(t)
		tenant2 := s.CreateNamespace(t)
		s.CreateTag(t, WithTagName("production"), WithTagTenant(tenant1))
		s.CreateTag(t, WithTagName("staging"), WithTagTenant(tenant1))
		s.CreateTag(t, WithTagName("development"), WithTagTenant(tenant2))

		tags, count, err := st.TagList(ctx)
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

		// Create tags in different namespaces
		tenant1 := s.CreateNamespace(t)
		tenant2 := s.CreateNamespace(t)
		s.CreateTag(t, WithTagName("production"), WithTagTenant(tenant1))
		s.CreateTag(t, WithTagName("staging"), WithTagTenant(tenant1))
		s.CreateTag(t, WithTagName("development"), WithTagTenant(tenant2))

		tags, count, err := st.TagList(ctx, st.Options().InNamespace(tenant1))
		require.NoError(t, err)
		assert.Equal(t, 2, count)
		assert.Len(t, tags, 2)

		sortTags(tags)
		assert.Equal(t, "production", tags[0].Name)
		assert.Equal(t, "staging", tags[1].Name)
		// Verify both tags belong to tenant1
		assert.Equal(t, tenant1, tags[0].TenantID)
		assert.Equal(t, tenant1, tags[1].TenantID)
	})
}

func (s *Suite) TestTagResolve(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when tag not found by ID", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create and delete a tag to get a valid but non-existent ID
		tenantID := s.CreateNamespace(t)
		tagID := s.CreateTag(t, WithTagName("temp"), WithTagTenant(tenantID))
		err := st.TagDelete(ctx, &models.Tag{ID: tagID, TenantID: tenantID})
		require.NoError(t, err)

		// Try to resolve the deleted tag
		tag, err := st.TagResolve(ctx, store.TagIDResolver, tagID)
		assert.ErrorIs(t, err, store.ErrNoDocuments)
		assert.Nil(t, tag)
	})

	t.Run("succeeds resolving tag by ID", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		tagID := s.CreateTag(t, WithTagName("production"), WithTagTenant(tenantID))

		tag, err := st.TagResolve(ctx, store.TagIDResolver, tagID)
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

		tag, err := st.TagResolve(ctx, store.TagNameResolver, "nonexistent")
		assert.ErrorIs(t, err, store.ErrNoDocuments)
		assert.Nil(t, tag)
	})

	t.Run("succeeds resolving tag by name with tenant filter", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		tagID := s.CreateTag(t, WithTagName("production"), WithTagTenant(tenantID))

		tag, err := st.TagResolve(ctx, store.TagNameResolver, "production", st.Options().InNamespace(tenantID))
		require.NoError(t, err)
		require.NotNil(t, tag)
		assert.Equal(t, tagID, tag.ID)
		assert.Equal(t, "production", tag.Name)
		assert.Equal(t, tenantID, tag.TenantID)
	})
}

func (s *Suite) TestTagUpdate(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when tag is not found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)

		// Create and delete a tag to get a valid but non-existent ID
		tagID := s.CreateTag(t, WithTagName("temp"), WithTagTenant(tenantID))
		err := st.TagDelete(ctx, &models.Tag{ID: tagID, TenantID: tenantID})
		require.NoError(t, err)

		// Try to update non-existent tag
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

		// Verify update
		updatedTag, err := st.TagResolve(ctx, store.TagIDResolver, tagID)
		require.NoError(t, err)
		assert.Equal(t, "edited-tag", updatedTag.Name)
	})
}

func (s *Suite) TestTagPushToTarget(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when tag does not exist", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		deviceUID := s.CreateDevice(t, WithTenantID(tenantID))

		// Create and delete a tag to get a valid but non-existent ID
		tagID := s.CreateTag(t, WithTagName("temp"), WithTagTenant(tenantID))
		err := st.TagDelete(ctx, &models.Tag{ID: tagID, TenantID: tenantID})
		require.NoError(t, err)

		// Try to push non-existent tag
		err = st.TagPushToTarget(ctx, tagID, store.TagTargetDevice, string(deviceUID))
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})

	t.Run("fails when device does not exist", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		tagID := s.CreateTag(t, WithTagName("staging"), WithTagTenant(tenantID))

		// Create and delete a device to get a valid but non-existent UID
		deviceUID := s.CreateDevice(t, WithTenantID(tenantID))
		device, err := st.DeviceResolve(ctx, store.DeviceUIDResolver, string(deviceUID))
		require.NoError(t, err)
		err = st.DeviceDelete(ctx, device)
		require.NoError(t, err)

		// Try to push tag to non-existent device
		err = st.TagPushToTarget(ctx, tagID, store.TagTargetDevice, string(deviceUID))
		assert.ErrorIs(t, err, store.ErrNoDocuments)
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

func (s *Suite) TestTagPullFromTarget(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when tag does not exist", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		deviceUID := s.CreateDevice(t, WithTenantID(tenantID))

		// Create and delete a tag to get a valid but non-existent ID
		tagID := s.CreateTag(t, WithTagName("temp"), WithTagTenant(tenantID))
		err := st.TagDelete(ctx, &models.Tag{ID: tagID, TenantID: tenantID})
		require.NoError(t, err)

		// Try to pull non-existent tag
		err = st.TagPullFromTarget(ctx, tagID, store.TagTargetDevice, string(deviceUID))
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})

	t.Run("fails when device does not exist", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		tagID := s.CreateTag(t, WithTagName("production"), WithTagTenant(tenantID))

		// Create and delete a device to get a valid but non-existent UID
		deviceUID := s.CreateDevice(t, WithTenantID(tenantID))
		device, err := st.DeviceResolve(ctx, store.DeviceUIDResolver, string(deviceUID))
		require.NoError(t, err)
		err = st.DeviceDelete(ctx, device)
		require.NoError(t, err)

		// Try to pull tag from non-existent device
		err = st.TagPullFromTarget(ctx, tagID, store.TagTargetDevice, string(deviceUID))
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})

	t.Run("succeeds to pull a tag from device", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		deviceUID := s.CreateDevice(t, WithTenantID(tenantID))
		tagID := s.CreateTag(t, WithTagName("production"), WithTagTenant(tenantID))

		// First push the tag to device
		err := st.TagPushToTarget(ctx, tagID, store.TagTargetDevice, string(deviceUID))
		require.NoError(t, err)

		// Then pull it
		err = st.TagPullFromTarget(ctx, tagID, store.TagTargetDevice, string(deviceUID))
		require.NoError(t, err)
	})

	t.Run("succeeds to pull a tag from all targets when no specific targets provided", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		device1 := s.CreateDevice(t, WithTenantID(tenantID))
		device2 := s.CreateDevice(t, WithTenantID(tenantID))
		tagID := s.CreateTag(t, WithTagName("production"), WithTagTenant(tenantID))

		// Push tag to both devices
		err := st.TagPushToTarget(ctx, tagID, store.TagTargetDevice, string(device1))
		require.NoError(t, err)
		err = st.TagPushToTarget(ctx, tagID, store.TagTargetDevice, string(device2))
		require.NoError(t, err)

		// Pull from all targets (no targetIDs provided)
		err = st.TagPullFromTarget(ctx, tagID, store.TagTargetDevice)
		require.NoError(t, err)
	})
}

func (s *Suite) TestTagDelete(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when tag is not found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)

		// Create and delete a tag to get a valid but non-existent ID
		tagID := s.CreateTag(t, WithTagName("temp"), WithTagTenant(tenantID))
		err := st.TagDelete(ctx, &models.Tag{ID: tagID, TenantID: tenantID})
		require.NoError(t, err)

		// Try to delete again (tag already deleted)
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

		// Verify deletion
		_, err = st.TagResolve(ctx, store.TagIDResolver, tagID)
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})
}
