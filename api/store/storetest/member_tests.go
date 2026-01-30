package storetest

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *Suite) TestNamespaceCreateMembership(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when tenant is not found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create a user but not a namespace
		userID := s.CreateUser(t)

		// Try to add membership to non-existent namespace
		member := &models.Member{
			ID:   userID,
			Role: authorizer.RoleObserver,
		}
		err := st.NamespaceCreateMembership(ctx, "99999999-9999-4999-9999-999999999999", member)
		assert.Error(t, err)
	})

	t.Run("succeeds when tenant is found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create namespace and user
		tenantID := s.CreateNamespace(t)
		userID := s.CreateUser(t)

		// Add user as member to namespace
		member := &models.Member{
			ID:   userID,
			Role: authorizer.RoleObserver,
		}
		err := st.NamespaceCreateMembership(ctx, tenantID, member)
		assert.NoError(t, err)
	})
}

func (s *Suite) TestNamespaceUpdateMembership(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when user is not found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create namespace with a member
		tenantID := s.CreateNamespace(t)
		userID := s.CreateUser(t)
		member := &models.Member{
			ID:   userID,
			Role: authorizer.RoleObserver,
		}
		err := st.NamespaceCreateMembership(ctx, tenantID, member)
		require.NoError(t, err)

		// Try to update membership for non-existent user
		nonExistentMember := &models.Member{
			ID:   "99999999-9999-4999-9999-999999999999",
			Role: authorizer.RoleAdministrator,
		}
		err = st.NamespaceUpdateMembership(ctx, tenantID, nonExistentMember)
		assert.Error(t, err)
	})

	t.Run("succeeds when tenant and user is found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create namespace with a member
		tenantID := s.CreateNamespace(t)
		userID := s.CreateUser(t)
		member := &models.Member{
			ID:   userID,
			Role: authorizer.RoleObserver,
		}
		err := st.NamespaceCreateMembership(ctx, tenantID, member)
		require.NoError(t, err)

		// Update member role
		member.Role = authorizer.RoleAdministrator
		err = st.NamespaceUpdateMembership(ctx, tenantID, member)
		assert.NoError(t, err)
	})
}

func (s *Suite) TestNamespaceDeleteMembership(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when tenant is not found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create namespace with a member, then delete the namespace
		tenantID := s.CreateNamespace(t)
		userID := s.CreateUser(t)
		member := &models.Member{
			ID:   userID,
			Role: authorizer.RoleObserver,
		}
		err := st.NamespaceCreateMembership(ctx, tenantID, member)
		require.NoError(t, err)

		// Delete the namespace
		ns, err := st.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, tenantID)
		require.NoError(t, err)
		err = st.NamespaceDelete(ctx, ns)
		require.NoError(t, err)

		// Try to delete membership from deleted namespace
		err = st.NamespaceDeleteMembership(ctx, tenantID, member)
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})

	t.Run("succeeds when tenant and user is found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create namespace with a member
		tenantID := s.CreateNamespace(t)
		userID := s.CreateUser(t)
		member := &models.Member{
			ID:   userID,
			Role: authorizer.RoleObserver,
		}
		err := st.NamespaceCreateMembership(ctx, tenantID, member)
		require.NoError(t, err)

		// Delete membership
		err = st.NamespaceDeleteMembership(ctx, tenantID, member)
		assert.NoError(t, err)
	})
}
