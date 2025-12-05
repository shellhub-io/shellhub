package mongo_test

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/clock"
	clockmock "github.com/shellhub-io/shellhub/pkg/clock/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestStore_MembershipInvitationCreate(t *testing.T) {
	mockClock := new(clockmock.Clock)
	clock.DefaultBackend = mockClock

	now := time.Now()
	mockClock.On("Now").Return(now)
	expiresAt := now.Add(7 * 24 * time.Hour)

	cases := []struct {
		description string
		invitation  *models.MembershipInvitation
		fixtures    []string
		expected    map[string]any
	}{
		{
			description: "succeeds creating new invitation",
			invitation: &models.MembershipInvitation{
				TenantID:    "00000000-0000-4000-0000-000000000000",
				UserID:      "6509e169ae6144b2f56bf288",
				InvitedBy:   "507f1f77bcf86cd799439011",
				Role:        authorizer.RoleObserver,
				Status:      models.MembershipInvitationStatusPending,
				ExpiresAt:   &expiresAt,
				Invitations: 1,
			},
			fixtures: []string{},
			expected: map[string]any{
				"tenant_id":         "00000000-0000-4000-0000-000000000000",
				"role":              "observer",
				"status":            "pending",
				"created_at":        primitive.NewDateTimeFromTime(now),
				"updated_at":        primitive.NewDateTimeFromTime(now),
				"status_updated_at": primitive.NewDateTimeFromTime(now),
				"invitations":       int32(1),
			},
		},
		{
			description: "succeeds creating invitation with ID",
			invitation: &models.MembershipInvitation{
				ID:          "507f1f77bcf86cd799439020",
				TenantID:    "00000000-0000-4001-0000-000000000000",
				UserID:      "907f1f77bcf86cd799439022",
				InvitedBy:   "6509e169ae6144b2f56bf288",
				Role:        authorizer.RoleAdministrator,
				Status:      models.MembershipInvitationStatusAccepted,
				ExpiresAt:   &expiresAt,
				Invitations: 2,
			},
			fixtures: []string{},
			expected: map[string]any{
				"tenant_id":         "00000000-0000-4001-0000-000000000000",
				"role":              "administrator",
				"status":            "accepted",
				"created_at":        primitive.NewDateTimeFromTime(now),
				"updated_at":        primitive.NewDateTimeFromTime(now),
				"status_updated_at": primitive.NewDateTimeFromTime(now),
				"invitations":       int32(2),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			ctx := context.Background()

			require.NoError(tt, srv.Apply(tc.fixtures...))
			tt.Cleanup(func() {
				require.NoError(tt, srv.Reset())
			})

			err := s.MembershipInvitationCreate(ctx, tc.invitation)
			require.NoError(tt, err)
			require.NotEmpty(tt, tc.invitation.ID)

			objID, _ := primitive.ObjectIDFromHex(tc.invitation.ID)
			userObjID, _ := primitive.ObjectIDFromHex(tc.invitation.UserID)
			invitedByObjID, _ := primitive.ObjectIDFromHex(tc.invitation.InvitedBy)

			tmpInvitation := make(map[string]any)
			require.NoError(tt, db.Collection("membership_invitations").FindOne(ctx, bson.M{"_id": objID}).Decode(&tmpInvitation))

			require.Equal(tt, objID, tmpInvitation["_id"])
			require.Equal(tt, userObjID, tmpInvitation["user_id"])
			require.Equal(tt, invitedByObjID, tmpInvitation["invited_by"])

			for field, expectedValue := range tc.expected {
				require.Equal(tt, expectedValue, tmpInvitation[field])
			}
		})
	}
}

func TestStore_MembershipInvitationResolve(t *testing.T) {
	type Expected struct {
		invitation *models.MembershipInvitation
		err        error
	}

	cases := []struct {
		description string
		tenantID    string
		userID      string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when invitation not found",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			userID:      "000000000000000000000000",
			fixtures:    []string{fixtureMembershipInvitations, fixtureNamespaces, fixtureUsers},
			expected:    Expected{invitation: nil, err: store.ErrNoDocuments},
		},
		{
			description: "succeeds fetching email from users collection",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			userID:      "6509e169ae6144b2f56bf288",
			fixtures:    []string{fixtureMembershipInvitations, fixtureNamespaces, fixtureUsers},
			expected: Expected{
				invitation: &models.MembershipInvitation{
					ID:            "507f1f77bcf86cd799439012",
					TenantID:      "00000000-0000-4000-0000-000000000000",
					NamespaceName: "namespace-1",
					UserID:        "6509e169ae6144b2f56bf288",
					UserEmail:     "maria.garcia@test.com",
					InvitedBy:     "507f1f77bcf86cd799439011",
					Role:          authorizer.RoleObserver,
					Status:        models.MembershipInvitationStatusPending,
				},
				err: nil,
			},
		},
		{
			description: "succeeds fetching email from user_invitations collection",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			userID:      "507f1f77bcf86cd799439011",
			fixtures:    []string{fixtureMembershipInvitations, fixtureNamespaces, fixtureUserInvitations},
			expected: Expected{
				invitation: &models.MembershipInvitation{
					ID:            "507f1f77bcf86cd799439014",
					TenantID:      "00000000-0000-4000-0000-000000000000",
					NamespaceName: "namespace-1",
					UserID:        "507f1f77bcf86cd799439011",
					UserEmail:     "jane.doe@test.com",
					InvitedBy:     "6509e169ae6144b2f56bf288",
					Role:          authorizer.RoleObserver,
					Status:        models.MembershipInvitationStatusPending,
				},
				err: nil,
			},
		},
		{
			description: "returns most recent when multiple invitations exist",
			tenantID:    "00000000-0000-4001-0000-000000000000",
			userID:      "608f32a2c7351f001f6475e0",
			fixtures:    []string{fixtureMembershipInvitations, fixtureNamespaces, fixtureUsers},
			expected: Expected{
				invitation: &models.MembershipInvitation{
					ID:            "507f1f77bcf86cd799439013",
					TenantID:      "00000000-0000-4001-0000-000000000000",
					NamespaceName: "namespace-2",
					UserID:        "608f32a2c7351f001f6475e0",
					UserEmail:     "jane.smith@test.com",
					InvitedBy:     "6509e169ae6144b2f56bf288",
					Role:          authorizer.RoleAdministrator,
					Status:        models.MembershipInvitationStatusAccepted,
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			ctx := context.Background()

			require.NoError(tt, srv.Apply(tc.fixtures...))
			tt.Cleanup(func() {
				require.NoError(tt, srv.Reset())
			})

			invitation, err := s.MembershipInvitationResolve(ctx, tc.tenantID, tc.userID)

			if tc.expected.err != nil {
				require.Equal(tt, tc.expected.err, err)
				require.Nil(tt, invitation)
			} else {
				require.NoError(tt, err)
				require.NotNil(tt, invitation)
				require.Equal(tt, tc.expected.invitation.ID, invitation.ID)
				require.Equal(tt, tc.expected.invitation.TenantID, invitation.TenantID)
				require.Equal(tt, tc.expected.invitation.NamespaceName, invitation.NamespaceName)
				require.Equal(tt, tc.expected.invitation.UserID, invitation.UserID)
				require.Equal(tt, tc.expected.invitation.UserEmail, invitation.UserEmail)
				require.Equal(tt, tc.expected.invitation.InvitedBy, invitation.InvitedBy)
				require.Equal(tt, tc.expected.invitation.Role, invitation.Role)
				require.Equal(tt, tc.expected.invitation.Status, invitation.Status)
			}
		})
	}
}

func TestStore_MembershipInvitationUpdate(t *testing.T) {
	mockClock := new(clockmock.Clock)
	clock.DefaultBackend = mockClock

	now := time.Now()
	mockClock.On("Now").Return(now)

	type Expected struct {
		err error
	}

	cases := []struct {
		description string
		invitation  *models.MembershipInvitation
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when invitation not found",
			invitation: &models.MembershipInvitation{
				ID:              "000000000000000000000000",
				TenantID:        "00000000-0000-4000-0000-000000000000",
				UserID:          "6509e169ae6144b2f56bf288",
				InvitedBy:       "507f1f77bcf86cd799439011",
				Role:            authorizer.RoleObserver,
				Status:          models.MembershipInvitationStatusPending,
				StatusUpdatedAt: now,
				Invitations:     2,
			},
			fixtures: []string{fixtureMembershipInvitations},
			expected: Expected{err: store.ErrNoDocuments},
		},
		{
			description: "succeeds when invitation found",
			invitation: &models.MembershipInvitation{
				ID:              "507f1f77bcf86cd799439012",
				TenantID:        "00000000-0000-4000-0000-000000000000",
				UserID:          "6509e169ae6144b2f56bf288",
				InvitedBy:       "507f1f77bcf86cd799439011",
				Role:            authorizer.RoleAdministrator,
				Status:          models.MembershipInvitationStatusAccepted,
				StatusUpdatedAt: now,
				Invitations:     3,
			},
			fixtures: []string{fixtureMembershipInvitations},
			expected: Expected{err: nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			ctx := context.Background()

			require.NoError(tt, srv.Apply(tc.fixtures...))
			tt.Cleanup(func() {
				require.NoError(tt, srv.Reset())
			})

			err := s.MembershipInvitationUpdate(ctx, tc.invitation)

			if tc.expected.err != nil {
				require.Equal(tt, tc.expected.err, err)
			} else {
				require.NoError(tt, err)

				objID, _ := primitive.ObjectIDFromHex(tc.invitation.ID)
				updatedInvitation := &models.MembershipInvitation{}
				require.NoError(tt, db.Collection("membership_invitations").FindOne(ctx, bson.M{"_id": objID}).Decode(updatedInvitation))

				require.Equal(tt, tc.invitation.Role, updatedInvitation.Role)
				require.Equal(tt, tc.invitation.Status, updatedInvitation.Status)
				require.Equal(tt, tc.invitation.Invitations, updatedInvitation.Invitations)
				require.Equal(tt, primitive.NewDateTimeFromTime(now), primitive.NewDateTimeFromTime(updatedInvitation.UpdatedAt))
			}
		})
	}
}
