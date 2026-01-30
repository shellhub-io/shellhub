package storetest

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *Suite) TestMembershipInvitationCreate(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("succeeds creating new invitation", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create namespace and users
		tenantID := s.CreateNamespace(t)
		invitedBy := s.CreateUser(t)
		invitedUser := s.CreateUser(t)

		expiresAt := time.Now().Add(7 * 24 * time.Hour)
		invitation := &models.MembershipInvitation{
			TenantID:    tenantID,
			UserID:      invitedUser,
			InvitedBy:   invitedBy,
			Role:        authorizer.RoleObserver,
			Status:      models.MembershipInvitationStatusPending,
			ExpiresAt:   &expiresAt,
			Invitations: 1,
		}

		err := st.MembershipInvitationCreate(ctx, invitation)
		assert.NoError(t, err)
		assert.NotEmpty(t, invitation.ID)
	})
}

func (s *Suite) TestMembershipInvitationResolve(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when invitation not found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create namespace and user but no invitation
		tenantID := s.CreateNamespace(t)
		userID := s.CreateUser(t)

		invitation, err := st.MembershipInvitationResolve(ctx, tenantID, userID)
		assert.ErrorIs(t, err, store.ErrNoDocuments)
		assert.Nil(t, invitation)
	})

	t.Run("succeeds when invitation is found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create namespace and users
		tenantID := s.CreateNamespace(t)
		invitedBy := s.CreateUser(t)
		invitedUser := s.CreateUser(t)

		// Create invitation
		expiresAt := time.Now().Add(7 * 24 * time.Hour)
		createdInvitation := &models.MembershipInvitation{
			TenantID:    tenantID,
			UserID:      invitedUser,
			InvitedBy:   invitedBy,
			Role:        authorizer.RoleObserver,
			Status:      models.MembershipInvitationStatusPending,
			ExpiresAt:   &expiresAt,
			Invitations: 1,
		}
		err := st.MembershipInvitationCreate(ctx, createdInvitation)
		require.NoError(t, err)

		// Resolve invitation
		invitation, err := st.MembershipInvitationResolve(ctx, tenantID, invitedUser)
		require.NoError(t, err)
		require.NotNil(t, invitation)
		assert.Equal(t, tenantID, invitation.TenantID)
		assert.Equal(t, invitedUser, invitation.UserID)
		assert.Equal(t, authorizer.RoleObserver, invitation.Role)
		assert.Equal(t, models.MembershipInvitationStatusPending, invitation.Status)
	})
}

func (s *Suite) TestMembershipInvitationUpdate(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when invitation not found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create invitation, then try to update with non-existent ID
		tenantID := s.CreateNamespace(t)
		invitedBy := s.CreateUser(t)
		invitedUser := s.CreateUser(t)

		// Try to update non-existent invitation
		nonExistentInvitation := &models.MembershipInvitation{
			ID:          "99999999-9999-4999-9999-999999999999",
			TenantID:    tenantID,
			UserID:      invitedUser,
			InvitedBy:   invitedBy,
			Role:        authorizer.RoleAdministrator,
			Status:      models.MembershipInvitationStatusAccepted,
			Invitations: 2,
		}
		err := st.MembershipInvitationUpdate(ctx, nonExistentInvitation)
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})

	t.Run("succeeds when invitation is found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create namespace and users
		tenantID := s.CreateNamespace(t)
		invitedBy := s.CreateUser(t)
		invitedUser := s.CreateUser(t)

		// Create invitation
		expiresAt := time.Now().Add(7 * 24 * time.Hour)
		invitation := &models.MembershipInvitation{
			TenantID:    tenantID,
			UserID:      invitedUser,
			InvitedBy:   invitedBy,
			Role:        authorizer.RoleObserver,
			Status:      models.MembershipInvitationStatusPending,
			ExpiresAt:   &expiresAt,
			Invitations: 1,
		}
		err := st.MembershipInvitationCreate(ctx, invitation)
		require.NoError(t, err)

		// Update invitation
		invitation.Role = authorizer.RoleAdministrator
		invitation.Status = models.MembershipInvitationStatusAccepted
		invitation.Invitations = 3
		err = st.MembershipInvitationUpdate(ctx, invitation)
		assert.NoError(t, err)
	})
}
