package entity

import (
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestMembershipFromModel(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		namespaceID string
		member      *models.Member
		expected    *Membership
	}{
		{
			name:        "full fields",
			namespaceID: "ns-id-1",
			member: &models.Member{
				ID:      "user-id-1",
				AddedAt: now,
				Role:    authorizer.RoleAdministrator,
			},
			expected: &Membership{
				UserID:      "user-id-1",
				NamespaceID: "ns-id-1",
				CreatedAt:   now,
				UpdatedAt:   time.Time{},
				Role:        "administrator",
			},
		},
		{
			name:        "empty Role defaults to observer",
			namespaceID: "ns-id-2",
			member: &models.Member{
				ID:      "user-id-2",
				AddedAt: now,
				Role:    "",
			},
			expected: &Membership{
				UserID:      "user-id-2",
				NamespaceID: "ns-id-2",
				CreatedAt:   now,
				UpdatedAt:   time.Time{},
				Role:        "observer",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MembershipFromModel(tt.namespaceID, tt.member)
			assert.Equal(t, tt.expected.UserID, result.UserID)
			assert.Equal(t, tt.expected.NamespaceID, result.NamespaceID)
			assert.Equal(t, tt.expected.CreatedAt, result.CreatedAt)
			assert.True(t, result.UpdatedAt.IsZero(), "UpdatedAt should be zero")
			assert.Equal(t, tt.expected.Role, result.Role)
		})
	}
}

func TestMembershipToModel(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		entity   *Membership
		expected *models.Member
	}{
		{
			name: "with User loaded",
			entity: &Membership{
				UserID:    "user-id-1",
				CreatedAt: now,
				Role:      "administrator",
				User: &User{
					Email: "user@example.com",
				},
			},
			expected: &models.Member{
				ID:      "user-id-1",
				AddedAt: now,
				Role:    authorizer.RoleAdministrator,
				Email:   "user@example.com",
			},
		},
		{
			name: "nil User",
			entity: &Membership{
				UserID:    "user-id-2",
				CreatedAt: now,
				Role:      "observer",
				User:      nil,
			},
			expected: &models.Member{
				ID:      "user-id-2",
				AddedAt: now,
				Role:    authorizer.RoleObserver,
				Email:   "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MembershipToModel(tt.entity)
			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.AddedAt, result.AddedAt)
			assert.Equal(t, tt.expected.Role, result.Role)
			assert.Equal(t, tt.expected.Email, result.Email)
		})
	}
}
