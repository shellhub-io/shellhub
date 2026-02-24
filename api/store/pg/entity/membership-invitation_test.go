package entity

import (
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestMembershipInvitationFromModel(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	tests := []struct {
		name     string
		model    *models.MembershipInvitation
		expected *MembershipInvitation
	}{
		{
			name: "full fields",
			model: &models.MembershipInvitation{
				ID:              "inv-id-1",
				TenantID:        "tenant-id-1",
				UserID:          "user-id-1",
				InvitedBy:       "admin-id-1",
				Role:            authorizer.RoleOperator,
				Status:          models.MembershipInvitationStatusAccepted,
				StatusUpdatedAt: now,
				ExpiresAt:       &expiresAt,
				Invitations:     3,
				CreatedAt:       now,
				UpdatedAt:       now,
			},
			expected: &MembershipInvitation{
				ID:              "inv-id-1",
				TenantID:        "tenant-id-1",
				UserID:          "user-id-1",
				InvitedBy:       "admin-id-1",
				Role:            "operator",
				Status:          "accepted",
				StatusUpdatedAt: now,
				ExpiresAt:       &expiresAt,
				Invitations:     3,
				CreatedAt:       now,
				UpdatedAt:       now,
			},
		},
		{
			name: "empty Role defaults to observer",
			model: &models.MembershipInvitation{
				ID:       "inv-id-2",
				TenantID: "tenant-id-2",
				UserID:   "user-id-2",
				Role:     "",
				Status:   models.MembershipInvitationStatusPending,
			},
			expected: &MembershipInvitation{
				ID:       "inv-id-2",
				TenantID: "tenant-id-2",
				UserID:   "user-id-2",
				Role:     "observer",
				Status:   "pending",
			},
		},
		{
			name: "empty Status defaults to pending",
			model: &models.MembershipInvitation{
				ID:       "inv-id-3",
				TenantID: "tenant-id-3",
				UserID:   "user-id-3",
				Role:     authorizer.RoleAdministrator,
				Status:   "",
			},
			expected: &MembershipInvitation{
				ID:       "inv-id-3",
				TenantID: "tenant-id-3",
				UserID:   "user-id-3",
				Role:     "administrator",
				Status:   "pending",
			},
		},
		{
			name: "nil ExpiresAt",
			model: &models.MembershipInvitation{
				ID:        "inv-id-4",
				TenantID:  "tenant-id-4",
				UserID:    "user-id-4",
				Role:      authorizer.RoleObserver,
				Status:    models.MembershipInvitationStatusPending,
				ExpiresAt: nil,
			},
			expected: &MembershipInvitation{
				ID:        "inv-id-4",
				TenantID:  "tenant-id-4",
				UserID:    "user-id-4",
				Role:      "observer",
				Status:    "pending",
				ExpiresAt: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MembershipInvitationFromModel(tt.model)
			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.TenantID, result.TenantID)
			assert.Equal(t, tt.expected.UserID, result.UserID)
			assert.Equal(t, tt.expected.InvitedBy, result.InvitedBy)
			assert.Equal(t, tt.expected.Role, result.Role)
			assert.Equal(t, tt.expected.Status, result.Status)
			assert.Equal(t, tt.expected.ExpiresAt, result.ExpiresAt)
			assert.Equal(t, tt.expected.Invitations, result.Invitations)
			assert.Equal(t, tt.expected.StatusUpdatedAt, result.StatusUpdatedAt)
			assert.Equal(t, tt.expected.CreatedAt, result.CreatedAt)
			assert.Equal(t, tt.expected.UpdatedAt, result.UpdatedAt)
		})
	}
}

func TestMembershipInvitationToModel(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	tests := []struct {
		name     string
		entity   *MembershipInvitation
		expected *models.MembershipInvitation
	}{
		{
			name: "Namespace and User loaded",
			entity: &MembershipInvitation{
				ID:              "inv-id-1",
				TenantID:        "tenant-id-1",
				UserID:          "user-id-1",
				InvitedBy:       "admin-id-1",
				Role:            "administrator",
				Status:          "accepted",
				StatusUpdatedAt: now,
				ExpiresAt:       &expiresAt,
				Invitations:     2,
				CreatedAt:       now,
				UpdatedAt:       now,
				Namespace:       &Namespace{Name: "my-namespace"},
				User:            &User{Email: "user@example.com"},
			},
			expected: &models.MembershipInvitation{
				ID:              "inv-id-1",
				TenantID:        "tenant-id-1",
				UserID:          "user-id-1",
				InvitedBy:       "admin-id-1",
				Role:            authorizer.RoleAdministrator,
				Status:          models.MembershipInvitationStatusAccepted,
				StatusUpdatedAt: now,
				ExpiresAt:       &expiresAt,
				Invitations:     2,
				CreatedAt:       now,
				UpdatedAt:       now,
				NamespaceName:   "my-namespace",
				UserEmail:       "user@example.com",
			},
		},
		{
			name: "nil Namespace",
			entity: &MembershipInvitation{
				ID:        "inv-id-2",
				TenantID:  "tenant-id-2",
				UserID:    "user-id-2",
				Role:      "observer",
				Status:    "pending",
				Namespace: nil,
				User:      &User{Email: "user2@example.com"},
			},
			expected: &models.MembershipInvitation{
				ID:            "inv-id-2",
				TenantID:      "tenant-id-2",
				UserID:        "user-id-2",
				Role:          authorizer.RoleObserver,
				Status:        models.MembershipInvitationStatusPending,
				NamespaceName: "",
				UserEmail:     "user2@example.com",
			},
		},
		{
			name: "nil User with UserInvitation fallback",
			entity: &MembershipInvitation{
				ID:             "inv-id-3",
				TenantID:       "tenant-id-3",
				UserID:         "user-id-3",
				Role:           "operator",
				Status:         "pending",
				Namespace:      &Namespace{Name: "test-ns"},
				User:           nil,
				UserInvitation: &UserInvitation{Email: "invited@example.com"},
			},
			expected: &models.MembershipInvitation{
				ID:            "inv-id-3",
				TenantID:      "tenant-id-3",
				UserID:        "user-id-3",
				Role:          authorizer.RoleOperator,
				Status:        models.MembershipInvitationStatusPending,
				NamespaceName: "test-ns",
				UserEmail:     "invited@example.com",
			},
		},
		{
			name: "nil Namespace and nil User and nil UserInvitation",
			entity: &MembershipInvitation{
				ID:             "inv-id-4",
				TenantID:       "tenant-id-4",
				UserID:         "user-id-4",
				Role:           "observer",
				Status:         "pending",
				Namespace:      nil,
				User:           nil,
				UserInvitation: nil,
			},
			expected: &models.MembershipInvitation{
				ID:            "inv-id-4",
				TenantID:      "tenant-id-4",
				UserID:        "user-id-4",
				Role:          authorizer.RoleObserver,
				Status:        models.MembershipInvitationStatusPending,
				NamespaceName: "",
				UserEmail:     "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MembershipInvitationToModel(tt.entity)
			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.TenantID, result.TenantID)
			assert.Equal(t, tt.expected.UserID, result.UserID)
			assert.Equal(t, tt.expected.InvitedBy, result.InvitedBy)
			assert.Equal(t, tt.expected.Role, result.Role)
			assert.Equal(t, tt.expected.Status, result.Status)
			assert.Equal(t, tt.expected.ExpiresAt, result.ExpiresAt)
			assert.Equal(t, tt.expected.Invitations, result.Invitations)
			assert.Equal(t, tt.expected.NamespaceName, result.NamespaceName)
			assert.Equal(t, tt.expected.UserEmail, result.UserEmail)
			assert.Equal(t, tt.expected.StatusUpdatedAt, result.StatusUpdatedAt)
			assert.Equal(t, tt.expected.CreatedAt, result.CreatedAt)
			assert.Equal(t, tt.expected.UpdatedAt, result.UpdatedAt)
		})
	}
}
