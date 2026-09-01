package storetest

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNamespaceCreateMembership exercises NamespaceCreateMembership against the store under test.
func (s *Suite) TestNamespaceCreateMembership(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when tenant is not found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		userID := s.CreateUser(t)

		member := &models.Member{
			ID:   userID,
			Role: authorizer.RoleObserver,
		}
		err := st.NamespaceCreateMembership(ctx, scope.MustBounded("99999999-9999-4999-9999-999999999999"), member)
		assert.Error(t, err)
	})

	t.Run("succeeds when tenant is found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		userID := s.CreateUser(t)

		member := &models.Member{
			ID:   userID,
			Role: authorizer.RoleObserver,
		}
		err := st.NamespaceCreateMembership(ctx, scope.MustBounded(tenantID), member)
		assert.NoError(t, err)
	})
}

// TestNamespaceUpdateMembership exercises NamespaceUpdateMembership against the store under test.
func (s *Suite) TestNamespaceUpdateMembership(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when user is not found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		userID := s.CreateUser(t)
		member := &models.Member{
			ID:   userID,
			Role: authorizer.RoleObserver,
		}
		err := st.NamespaceCreateMembership(ctx, scope.MustBounded(tenantID), member)
		require.NoError(t, err)

		nonExistentMember := &models.Member{
			ID:   "99999999-9999-4999-9999-999999999999",
			Role: authorizer.RoleAdministrator,
		}
		err = st.NamespaceUpdateMembership(ctx, scope.MustBounded(tenantID), nonExistentMember)
		assert.Error(t, err)
	})

	t.Run("succeeds when tenant and user is found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		userID := s.CreateUser(t)
		member := &models.Member{
			ID:   userID,
			Role: authorizer.RoleObserver,
		}
		err := st.NamespaceCreateMembership(ctx, scope.MustBounded(tenantID), member)
		require.NoError(t, err)

		member.Role = authorizer.RoleAdministrator
		err = st.NamespaceUpdateMembership(ctx, scope.MustBounded(tenantID), member)
		assert.NoError(t, err)
	})
}

// TestNamespaceDeleteMembership exercises NamespaceDeleteMembership against the store under test.
func (s *Suite) TestNamespaceDeleteMembership(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when tenant is not found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		userID := s.CreateUser(t)
		member := &models.Member{
			ID:   userID,
			Role: authorizer.RoleObserver,
		}
		err := st.NamespaceCreateMembership(ctx, scope.MustBounded(tenantID), member)
		require.NoError(t, err)

		ns, err := st.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, tenantID)
		require.NoError(t, err)
		err = st.NamespaceDelete(ctx, ns)
		require.NoError(t, err)

		err = st.NamespaceDeleteMembership(ctx, scope.MustBounded(tenantID), member)
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})

	t.Run("succeeds when tenant and user is found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		userID := s.CreateUser(t)
		member := &models.Member{
			ID:   userID,
			Role: authorizer.RoleObserver,
		}
		err := st.NamespaceCreateMembership(ctx, scope.MustBounded(tenantID), member)
		require.NoError(t, err)

		err = st.NamespaceDeleteMembership(ctx, scope.MustBounded(tenantID), member)
		assert.NoError(t, err)
	})

	t.Run("clears preferred namespace on delete", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		userID := s.CreateUser(t)
		tenantID := s.CreateNamespace(t, WithOwner(userID))
		s.CreateMembership(t, tenantID, userID, "observer")

		require.NoError(t, st.UserUpdatePreferredNamespace(ctx, userID, tenantID))

		user, err := st.UserResolve(ctx, store.UserIDResolver, userID)
		require.NoError(t, err)
		assert.Equal(t, tenantID, user.Preferences.PreferredNamespace)

		member := &models.Member{
			ID:   userID,
			Role: authorizer.RoleObserver,
		}
		err = st.NamespaceDeleteMembership(ctx, scope.MustBounded(tenantID), member)
		require.NoError(t, err)

		user, err = st.UserResolve(ctx, store.UserIDResolver, userID)
		require.NoError(t, err)
		assert.Empty(t, user.Preferences.PreferredNamespace)
	})
}
