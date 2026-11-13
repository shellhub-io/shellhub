package services

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type invitationE2E struct {
	*enrollmentE2E
	ownerID string
}

func setupInvitationE2E(t *testing.T) *invitationE2E {
	t.Helper()

	e := setupEnrollmentE2E(t)

	ns, err := e.st.NamespaceResolve(context.Background(), store.NamespaceTenantIDResolver, e.tenantID)
	require.NoError(t, err)

	return &invitationE2E{enrollmentE2E: e, ownerID: ns.Owner}
}

func (e *invitationE2E) user(t *testing.T, name string) string {
	t.Helper()

	id, err := e.st.UserCreate(context.Background(), &models.User{
		Origin:   models.UserOriginLocal,
		Status:   models.UserStatusConfirmed,
		UserData: models.UserData{Name: name, Email: name + "@example.com", Username: name},
		Password: models.UserPassword{Hash: "hash"},
	})
	require.NoError(t, err)

	return id
}

func (e *invitationE2E) member(t *testing.T, name string, role authorizer.Role) string {
	t.Helper()

	id := e.user(t, name)
	require.NoError(t, e.st.NamespaceCreateMembership(context.Background(), scope.MustBounded(e.tenantID),
		&models.Member{ID: id, Role: role, AddedAt: e.clock.Now()}))

	return id
}

func (e *invitationE2E) invite(t *testing.T, inviterID, email string) string {
	t.Helper()

	link, err := e.svc.GenerateInvitationLink(context.Background(), scope.MustBounded(e.tenantID), &requests.GenerateInvitationLink{
		TenantID:    e.tenantID,
		UserID:      inviterID,
		MemberEmail: email,
		MemberRole:  authorizer.RoleAdministrator,
	})
	require.NoError(t, err)

	parsed, err := url.Parse(link)
	require.NoError(t, err)

	return parsed.Query().Get("invite")
}

func (e *invitationE2E) accept(userID, code string) error {
	return e.svc.AcceptInvite(context.Background(), &requests.AcceptInvite{TenantID: e.tenantID, UserID: userID, Code: code})
}

func (e *invitationE2E) register(username, code string) error {
	_, err := e.svc.RegisterUser(context.Background(), requests.RegisterUser{
		Name:     username,
		Username: username,
		Password: "secret123",
		Sig:      code,
	})

	return err
}

func (e *invitationE2E) isMember(t *testing.T, userID string) bool {
	t.Helper()

	ns, err := e.st.NamespaceResolve(context.Background(), store.NamespaceTenantIDResolver, e.tenantID)
	require.NoError(t, err)

	_, ok := ns.FindMember(userID)

	return ok
}

func (e *invitationE2E) invitationStatus(t *testing.T, inviteeID string) models.MembershipInvitationStatus {
	t.Helper()

	invitation, err := e.st.MembershipInvitationResolve(context.Background(), scope.MustBounded(e.tenantID), inviteeID)
	require.NoError(t, err)

	return invitation.Status
}

func (e *invitationE2E) userIDByEmail(t *testing.T, email string) string {
	t.Helper()

	invitation, err := e.st.UserInvitationGet(context.Background(), store.UserInvitationEmailResolver, email)
	require.NoError(t, err)

	return invitation.ID
}

// TestInvitationE2E_inviterLosesAuthority shows an administrator who issues administrator
// invitations and then loses that authority, on every path that takes it, leaves neither
// invitation redeemable.
func TestInvitationE2E_inviterLosesAuthority(t *testing.T) {
	t.Parallel()

	sc := func(e *invitationE2E) scope.Scope { return scope.MustBounded(e.tenantID) }

	cases := []struct {
		description   string
		loseAuthority func(t *testing.T, e *invitationE2E, adminID string)
	}{
		{
			description: "removed by the owner",
			loseAuthority: func(t *testing.T, e *invitationE2E, adminID string) {
				t.Helper()
				_, err := e.svc.RemoveNamespaceMember(context.Background(), sc(e), &requests.NamespaceRemoveMember{
					UserID: e.ownerID, TenantID: e.tenantID, MemberID: adminID,
				})
				require.NoError(t, err)
			},
		},
		{
			description: "leaves the namespace",
			loseAuthority: func(t *testing.T, e *invitationE2E, adminID string) {
				t.Helper()
				_, err := e.svc.LeaveNamespace(context.Background(), &requests.LeaveNamespace{
					UserID: adminID, TenantID: e.tenantID, AuthenticatedTenantID: "another-tenant",
				})
				require.NoError(t, err)
			},
		},
		{
			description: "demoted to operator",
			loseAuthority: func(t *testing.T, e *invitationE2E, adminID string) {
				t.Helper()
				require.NoError(t, e.svc.UpdateNamespaceMember(context.Background(), sc(e), &requests.NamespaceUpdateMember{
					UserID: e.ownerID, TenantID: e.tenantID, MemberID: adminID, MemberRole: authorizer.RoleOperator,
				}))
			},
		},
		{
			description: "demoted to observer",
			loseAuthority: func(t *testing.T, e *invitationE2E, adminID string) {
				t.Helper()
				require.NoError(t, e.svc.UpdateNamespaceMember(context.Background(), sc(e), &requests.NamespaceUpdateMember{
					UserID: e.ownerID, TenantID: e.tenantID, MemberID: adminID, MemberRole: authorizer.RoleObserver,
				}))
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			t.Parallel()

			e := setupInvitationE2E(t)
			adminID := e.member(t, "admin", authorizer.RoleAdministrator)
			altID := e.user(t, "alt")

			altCode := e.invite(t, adminID, "alt@example.com")
			sockCode := e.invite(t, adminID, "sock@example.com")
			sockID := e.userIDByEmail(t, "sock@example.com")

			tc.loseAuthority(t, e, adminID)

			assert.Equal(t, models.MembershipInvitationStatusCancelled, e.invitationStatus(t, altID))
			assert.Equal(t, models.MembershipInvitationStatusCancelled, e.invitationStatus(t, sockID))

			require.ErrorIs(t, e.accept(altID, altCode), ErrNamespaceMemberNotFound)
			require.ErrorIs(t, e.register("sock", sockCode), ErrAuthForbidden)

			assert.False(t, e.isMember(t, altID))
			assert.False(t, e.isMember(t, sockID))
		})
	}
}

// TestInvitationE2E_redemptionRechecksTheInviter shows redemption refuses a pending invitation
// whose inviter no longer holds authority over its role, even when nothing cancelled it, and
// still admits through one whose inviter does.
func TestInvitationE2E_redemptionRechecksTheInviter(t *testing.T) {
	t.Parallel()

	demote := func(t *testing.T, e *invitationE2E, memberID string) {
		t.Helper()
		require.NoError(t, e.st.NamespaceUpdateMembership(context.Background(), scope.MustBounded(e.tenantID),
			&models.Member{ID: memberID, Role: authorizer.RoleOperator}))
	}

	t.Run("refuses both redemptions once the inviter lost authority", func(t *testing.T) {
		t.Parallel()

		e := setupInvitationE2E(t)
		adminID := e.member(t, "admin", authorizer.RoleAdministrator)
		altID := e.user(t, "alt")

		altCode := e.invite(t, adminID, "alt@example.com")
		sockCode := e.invite(t, adminID, "sock@example.com")
		sockID := e.userIDByEmail(t, "sock@example.com")

		demote(t, e, adminID)

		require.Equal(t, models.MembershipInvitationStatusPending, e.invitationStatus(t, altID))
		require.ErrorIs(t, e.accept(altID, altCode), ErrNamespaceMemberNotFound)
		require.ErrorIs(t, e.register("sock", sockCode), ErrAuthForbidden)

		assert.False(t, e.isMember(t, altID))
		assert.False(t, e.isMember(t, sockID))
	})

	t.Run("admits through an inviter who kept authority", func(t *testing.T) {
		t.Parallel()

		e := setupInvitationE2E(t)
		adminID := e.member(t, "admin", authorizer.RoleAdministrator)
		altID := e.user(t, "alt")

		require.NoError(t, e.accept(altID, e.invite(t, adminID, "alt@example.com")))
		require.NoError(t, e.register("sock", e.invite(t, adminID, "sock@example.com")))

		assert.True(t, e.isMember(t, altID))
		assert.True(t, e.isMember(t, e.userIDByEmail(t, "sock@example.com")))
	})
}

// TestInvitationE2E_resendRecordsTheResender shows renewing an expired invitation makes the
// administrator who renewed it its inviter, so it survives the original inviter's removal.
func TestInvitationE2E_resendRecordsTheResender(t *testing.T) {
	t.Parallel()

	e := setupInvitationE2E(t)
	firstID := e.member(t, "first", authorizer.RoleAdministrator)
	secondID := e.member(t, "second", authorizer.RoleAdministrator)
	altID := e.user(t, "alt")

	e.invite(t, firstID, "alt@example.com")
	e.clock.Set(e.clock.Now().Add(8 * 24 * time.Hour))
	code := e.invite(t, secondID, "alt@example.com")

	invitation, err := e.st.MembershipInvitationResolve(context.Background(), scope.MustBounded(e.tenantID), altID)
	require.NoError(t, err)
	assert.Equal(t, secondID, invitation.InvitedBy)

	_, err = e.svc.RemoveNamespaceMember(context.Background(), scope.MustBounded(e.tenantID), &requests.NamespaceRemoveMember{
		UserID: e.ownerID, TenantID: e.tenantID, MemberID: firstID,
	})
	require.NoError(t, err)

	require.NoError(t, e.accept(altID, code))
	assert.True(t, e.isMember(t, altID))
}
