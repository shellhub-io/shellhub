package storetest

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *Suite) TestMembershipInvitationCreate(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("succeeds creating new invitation", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		invitedBy := s.CreateUser(t)
		invitedUser := s.CreateUser(t)

		expiresAt := clock.Now().Add(7 * 24 * time.Hour)
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

		tenantID := s.CreateNamespace(t)
		userID := s.CreateUser(t)

		invitation, err := st.MembershipInvitationResolve(ctx, tenantID, userID)
		assert.ErrorIs(t, err, store.ErrNoDocuments)
		assert.Nil(t, invitation)
	})

	t.Run("succeeds when invitation is found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		invitedBy := s.CreateUser(t)
		invitedUser := s.CreateUser(t)

		expiresAt := clock.Now().Add(7 * 24 * time.Hour)
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

		invitation, err := st.MembershipInvitationResolve(ctx, tenantID, invitedUser)
		require.NoError(t, err)
		require.NotNil(t, invitation)
		assert.Equal(t, tenantID, invitation.TenantID)
		assert.Equal(t, invitedUser, invitation.UserID)
		assert.Equal(t, authorizer.RoleObserver, invitation.Role)
		assert.Equal(t, models.MembershipInvitationStatusPending, invitation.Status)
	})

	// Regression: cancelling a pending invitation while a previously cancelled
	// invitation exists for the same (tenant, user) used to return the older
	// cancelled row when the store ordered by primary key. Resolve must return
	// the most recently created invitation regardless of status.
	t.Run("returns most recent invitation when multiple exist for the same user", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		invitedBy := s.CreateUser(t)
		invitedUser := s.CreateUser(t)

		expiresAt := clock.Now().Add(7 * 24 * time.Hour)

		cancelled := &models.MembershipInvitation{
			TenantID:    tenantID,
			UserID:      invitedUser,
			InvitedBy:   invitedBy,
			Role:        authorizer.RoleObserver,
			Status:      models.MembershipInvitationStatusCancelled,
			ExpiresAt:   &expiresAt,
			Invitations: 1,
		}
		require.NoError(t, st.MembershipInvitationCreate(ctx, cancelled))

		// Force a distinguishable created_at on the next row (PG stores microseconds,
		// so a couple of milliseconds is plenty).
		time.Sleep(2 * time.Millisecond)

		pending := &models.MembershipInvitation{
			TenantID:    tenantID,
			UserID:      invitedUser,
			InvitedBy:   invitedBy,
			Role:        authorizer.RoleObserver,
			Status:      models.MembershipInvitationStatusPending,
			ExpiresAt:   &expiresAt,
			Invitations: 1,
		}
		require.NoError(t, st.MembershipInvitationCreate(ctx, pending))

		invitation, err := st.MembershipInvitationResolve(ctx, tenantID, invitedUser)
		require.NoError(t, err)
		require.NotNil(t, invitation)
		assert.Equal(t, pending.ID, invitation.ID)
		assert.Equal(t, models.MembershipInvitationStatusPending, invitation.Status)
	})
}

func (s *Suite) TestMembershipInvitationResolveBySig(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when no invitation carries the signature", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		invitation, err := st.MembershipInvitationResolveBySig(ctx, "MISSINGSIG12")
		assert.ErrorIs(t, err, store.ErrNoDocuments)
		assert.Nil(t, invitation)
	})

	t.Run("succeeds resolving an unexpired invitation by signature", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		invitedBy := s.CreateUser(t)
		invitedUser := s.CreateUser(t)

		expiresAt := clock.Now().Add(7 * 24 * time.Hour)
		invitation := &models.MembershipInvitation{
			TenantID:    tenantID,
			UserID:      invitedUser,
			InvitedBy:   invitedBy,
			Role:        authorizer.RoleObserver,
			Status:      models.MembershipInvitationStatusPending,
			ExpiresAt:   &expiresAt,
			Sig:         "VALIDSIG1234",
			Invitations: 1,
		}
		require.NoError(t, st.MembershipInvitationCreate(ctx, invitation))

		resolved, err := st.MembershipInvitationResolveBySig(ctx, "VALIDSIG1234")
		require.NoError(t, err)
		require.NotNil(t, resolved)
		assert.Equal(t, invitation.ID, resolved.ID)
		assert.Equal(t, tenantID, resolved.TenantID)
		assert.Equal(t, invitedUser, resolved.UserID)
	})

	t.Run("fails when the invitation carrying the signature has expired", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		invitedBy := s.CreateUser(t)
		invitedUser := s.CreateUser(t)

		expiredAt := clock.Now().Add(-1 * time.Hour)
		invitation := &models.MembershipInvitation{
			TenantID:    tenantID,
			UserID:      invitedUser,
			InvitedBy:   invitedBy,
			Role:        authorizer.RoleObserver,
			Status:      models.MembershipInvitationStatusPending,
			ExpiresAt:   &expiredAt,
			Sig:         "EXPIREDSIG12",
			Invitations: 1,
		}
		require.NoError(t, st.MembershipInvitationCreate(ctx, invitation))

		resolved, err := st.MembershipInvitationResolveBySig(ctx, "EXPIREDSIG12")
		assert.ErrorIs(t, err, store.ErrNoDocuments)
		assert.Nil(t, resolved)
	})

	// A cancelled invitation keeps its sig and expiry (the row is retained for the
	// namespace listing), so the pending filter is what stops it from being consumed.
	t.Run("fails when the invitation carrying the signature was cancelled", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		invitedBy := s.CreateUser(t)
		invitedUser := s.CreateUser(t)

		expiresAt := clock.Now().Add(7 * 24 * time.Hour)
		invitation := &models.MembershipInvitation{
			TenantID:    tenantID,
			UserID:      invitedUser,
			InvitedBy:   invitedBy,
			Role:        authorizer.RoleObserver,
			Status:      models.MembershipInvitationStatusCancelled,
			ExpiresAt:   &expiresAt,
			Sig:         "CANCELLEDSIG",
			Invitations: 1,
		}
		require.NoError(t, st.MembershipInvitationCreate(ctx, invitation))

		resolved, err := st.MembershipInvitationResolveBySig(ctx, "CANCELLEDSIG")
		assert.ErrorIs(t, err, store.ErrNoDocuments)
		assert.Nil(t, resolved)
	})
}

func (s *Suite) TestMembershipInvitationUpdate(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when invitation not found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// All UUID fields must be valid for PG compatibility.
		nonExistentInvitation := &models.MembershipInvitation{
			ID:          "99999999-9999-4999-9999-999999999999",
			TenantID:    "99999999-9999-4999-9999-999999999998",
			UserID:      "99999999-9999-4999-9999-999999999997",
			InvitedBy:   "99999999-9999-4999-9999-999999999996",
			Role:        authorizer.RoleAdministrator,
			Status:      models.MembershipInvitationStatusAccepted,
			Invitations: 2,
		}
		err := st.MembershipInvitationUpdate(ctx, nonExistentInvitation)
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})

	t.Run("succeeds when invitation is found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		invitedBy := s.CreateUser(t)
		invitedUser := s.CreateUser(t)

		expiresAt := clock.Now().Add(7 * 24 * time.Hour)
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

		invitation.Role = authorizer.RoleAdministrator
		invitation.Status = models.MembershipInvitationStatusAccepted
		invitation.Invitations = 3
		err = st.MembershipInvitationUpdate(ctx, invitation)
		assert.NoError(t, err)
	})
}

func (s *Suite) TestMembershipInvitationDelete(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when invitation not found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		nonExistentInvitation := &models.MembershipInvitation{
			ID:        "99999999-9999-4999-9999-999999999999",
			TenantID:  "99999999-9999-4999-9999-999999999998",
			UserID:    "99999999-9999-4999-9999-999999999997",
			InvitedBy: "99999999-9999-4999-9999-999999999996",
		}
		err := st.MembershipInvitationDelete(ctx, nonExistentInvitation)
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})

	t.Run("succeeds deleting an existing invitation", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		tenantID := s.CreateNamespace(t)
		invitedBy := s.CreateUser(t)
		invitedUser := s.CreateUser(t)

		expiresAt := clock.Now().Add(7 * 24 * time.Hour)
		invitation := &models.MembershipInvitation{
			TenantID:    tenantID,
			UserID:      invitedUser,
			InvitedBy:   invitedBy,
			Role:        authorizer.RoleObserver,
			Status:      models.MembershipInvitationStatusPending,
			ExpiresAt:   &expiresAt,
			Invitations: 1,
		}
		require.NoError(t, st.MembershipInvitationCreate(ctx, invitation))

		err := st.MembershipInvitationDelete(ctx, invitation)
		assert.NoError(t, err)

		resolved, err := st.MembershipInvitationResolve(ctx, tenantID, invitedUser)
		assert.ErrorIs(t, err, store.ErrNoDocuments)
		assert.Nil(t, resolved)
	})
}

func (s *Suite) TestUserMembershipInvitationList(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()
	require.NoError(t, s.provider.CleanDatabase(t))

	userID := s.CreateUser(t)
	tenantID := s.CreateNamespace(t)
	inviterID := s.CreateUser(t)

	invite := &models.MembershipInvitation{
		UserID:    userID,
		TenantID:  tenantID,
		InvitedBy: inviterID,
		Status:    models.MembershipInvitationStatusPending,
		Role:      authorizer.RoleObserver,
	}
	require.NoError(t, st.MembershipInvitationCreate(ctx, invite))

	invitations, count, err := st.UserMembershipInvitationList(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
	require.Len(t, invitations, 1)
	assert.Equal(t, userID, invitations[0].UserID)
	assert.Equal(t, tenantID, invitations[0].TenantID)
}

func (s *Suite) TestNamespaceMembershipInvitationList(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()
	require.NoError(t, s.provider.CleanDatabase(t))

	userID1 := s.CreateUser(t)
	userID2 := s.CreateUser(t)
	tenantID := s.CreateNamespace(t)
	inviterID := s.CreateUser(t)

	for _, uid := range []string{userID1, userID2} {
		invite := &models.MembershipInvitation{
			UserID:    uid,
			TenantID:  tenantID,
			InvitedBy: inviterID,
			Status:    models.MembershipInvitationStatusPending,
			Role:      authorizer.RoleObserver,
		}
		require.NoError(t, st.MembershipInvitationCreate(ctx, invite))
	}

	invitations, count, err := st.NamespaceMembershipInvitationList(ctx, tenantID)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
	assert.Len(t, invitations, 2)
}

func (s *Suite) TestNamespaceMembershipInvitationListWithStatusFilter(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()
	require.NoError(t, s.provider.CleanDatabase(t))

	userID1 := s.CreateUser(t)
	userID2 := s.CreateUser(t)
	userID3 := s.CreateUser(t)
	tenantID := s.CreateNamespace(t)
	inviterID := s.CreateUser(t)

	invitations := []*models.MembershipInvitation{
		{
			UserID:    userID1,
			TenantID:  tenantID,
			InvitedBy: inviterID,
			Status:    models.MembershipInvitationStatusPending,
			Role:      authorizer.RoleObserver,
		},
		{
			UserID:    userID2,
			TenantID:  tenantID,
			InvitedBy: inviterID,
			Status:    models.MembershipInvitationStatusAccepted,
			Role:      authorizer.RoleObserver,
		},
		{
			UserID:    userID3,
			TenantID:  tenantID,
			InvitedBy: inviterID,
			Status:    models.MembershipInvitationStatusCancelled,
			Role:      authorizer.RoleObserver,
		},
	}

	for _, invite := range invitations {
		require.NoError(t, st.MembershipInvitationCreate(ctx, invite))
	}

	cases := []struct {
		status     models.MembershipInvitationStatus
		wantUserID string
	}{
		{models.MembershipInvitationStatusCancelled, userID3},
		{models.MembershipInvitationStatusPending, userID1},
		{models.MembershipInvitationStatusAccepted, userID2},
	}

	for _, tc := range cases {
		t.Run("filter by "+string(tc.status)+" status", func(t *testing.T) {
			result, count, err := st.NamespaceMembershipInvitationList(
				ctx,
				tenantID,
				st.Options().Match(&query.Filters{Data: []query.Filter{
					{Type: query.FilterTypeProperty, Params: &query.FilterProperty{Name: "status", Operator: "eq", Value: string(tc.status)}},
				}}),
			)
			require.NoError(t, err)
			assert.Equal(t, int64(1), count)
			require.Len(t, result, 1)
			assert.Equal(t, tc.wantUserID, result[0].UserID)
			assert.Equal(t, tc.status, result[0].Status)
		})
	}
}

func (s *Suite) TestUserInvitationsUpsert(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("creates a pending invitation on first upsert", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		id, err := st.UserInvitationsUpsert(ctx, "invitee@test.com")
		require.NoError(t, err)
		assert.NotEmpty(t, id)

		invitation, err := st.UserInvitationGet(ctx, store.UserInvitationEmailResolver, "invitee@test.com")
		require.NoError(t, err)
		require.NotNil(t, invitation)
		assert.Equal(t, models.UserInvitationStatusPending, invitation.Status)
		assert.Equal(t, 1, invitation.Invitations)
	})

	t.Run("increments the counter and keeps the id on a repeated upsert", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		firstID, err := st.UserInvitationsUpsert(ctx, "invitee@test.com")
		require.NoError(t, err)

		secondID, err := st.UserInvitationsUpsert(ctx, "invitee@test.com")
		require.NoError(t, err)
		assert.Equal(t, firstID, secondID)

		invitation, err := st.UserInvitationGet(ctx, store.UserInvitationIDResolver, firstID)
		require.NoError(t, err)
		require.NotNil(t, invitation)
		assert.Equal(t, 2, invitation.Invitations)
	})

	t.Run("normalizes the email to lowercase", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		_, err := st.UserInvitationsUpsert(ctx, "Mixed@Case.com")
		require.NoError(t, err)

		invitation, err := st.UserInvitationGet(ctx, store.UserInvitationEmailResolver, "mixed@case.com")
		require.NoError(t, err)
		require.NotNil(t, invitation)
		assert.Equal(t, "mixed@case.com", invitation.Email)
	})
}

func (s *Suite) TestUserInvitationGet(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when invitation not found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		invitation, err := st.UserInvitationGet(ctx, store.UserInvitationEmailResolver, "missing@test.com")
		assert.ErrorIs(t, err, store.ErrNoDocuments)
		assert.Nil(t, invitation)
	})

	t.Run("resolves by id and by email", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		id, err := st.UserInvitationsUpsert(ctx, "invitee@test.com")
		require.NoError(t, err)

		byID, err := st.UserInvitationGet(ctx, store.UserInvitationIDResolver, id)
		require.NoError(t, err)
		require.NotNil(t, byID)
		assert.Equal(t, "invitee@test.com", byID.Email)

		byEmail, err := st.UserInvitationGet(ctx, store.UserInvitationEmailResolver, "invitee@test.com")
		require.NoError(t, err)
		require.NotNil(t, byEmail)
		assert.Equal(t, id, byEmail.ID)
	})
}

func (s *Suite) TestUserInvitationUpdate(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when invitation not found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		nonExistent := &models.UserInvitation{
			ID:     "99999999-9999-4999-9999-999999999999",
			Email:  "missing@test.com",
			Status: models.UserInvitationStatusAccepted,
		}
		err := st.UserInvitationUpdate(ctx, nonExistent)
		assert.ErrorIs(t, err, store.ErrNoDocuments)
	})

	t.Run("succeeds updating an existing invitation", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		id, err := st.UserInvitationsUpsert(ctx, "invitee@test.com")
		require.NoError(t, err)

		invitation, err := st.UserInvitationGet(ctx, store.UserInvitationIDResolver, id)
		require.NoError(t, err)

		invitation.Status = models.UserInvitationStatusAccepted
		require.NoError(t, st.UserInvitationUpdate(ctx, invitation))

		updated, err := st.UserInvitationGet(ctx, store.UserInvitationIDResolver, id)
		require.NoError(t, err)
		assert.Equal(t, models.UserInvitationStatusAccepted, updated.Status)
	})
}
