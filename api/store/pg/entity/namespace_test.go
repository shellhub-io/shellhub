package entity

import (
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNamespaceFromModel(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		model    *models.Namespace
		expected *Namespace
	}{
		{
			name: "full fields with Settings and Members",
			model: &models.Namespace{
				TenantID: "ns-id-1",
				Name:     "my-namespace",
				Owner:    "owner-id-1",
				Type:     models.TypeTeam,
				Settings: &models.NamespaceSettings{
					SessionRecord:          true,
					ConnectionAnnouncement: "Welcome!",
				},
				Members: []models.Member{
					{
						ID:      "user-id-1",
						AddedAt: now,
						Role:    authorizer.RoleAdministrator,
					},
					{
						ID:      "user-id-2",
						AddedAt: now,
						Role:    authorizer.RoleObserver,
					},
				},
				MaxDevices:           10,
				DevicesAcceptedCount: 5,
				DevicesPendingCount:  2,
				DevicesRejectedCount: 1,
				DevicesRemovedCount:  3,
				CreatedAt:            now,
			},
			expected: &Namespace{
				ID:      "ns-id-1",
				Name:    "my-namespace",
				OwnerID: "owner-id-1",
				Type:    "team",
				Settings: NamespaceSettings{
					MaxDevices:             10,
					SessionRecord:          true,
					ConnectionAnnouncement: "Welcome!",
				},
				Memberships: []Membership{
					{
						UserID:      "user-id-1",
						NamespaceID: "ns-id-1",
						CreatedAt:   now,
						Role:        "administrator",
					},
					{
						UserID:      "user-id-2",
						NamespaceID: "ns-id-1",
						CreatedAt:   now,
						Role:        "observer",
					},
				},
				DevicesAcceptedCount: 5,
				DevicesPendingCount:  2,
				DevicesRejectedCount: 1,
				DevicesRemovedCount:  3,
				CreatedAt:            now,
			},
		},
		{
			name: "empty Type defaults to personal",
			model: &models.Namespace{
				TenantID: "ns-id-2",
				Name:     "personal-ns",
				Owner:    "owner-id-2",
				Type:     "",
				Members:  []models.Member{},
			},
			expected: &Namespace{
				ID:          "ns-id-2",
				Name:        "personal-ns",
				OwnerID:     "owner-id-2",
				Type:        "personal",
				Memberships: []Membership{},
			},
		},
		{
			name: "nil Settings with MaxDevices still set",
			model: &models.Namespace{
				TenantID:   "ns-id-3",
				Name:       "no-settings",
				Owner:      "owner-id-3",
				Type:       models.TypePersonal,
				Settings:   nil,
				MaxDevices: 15,
				Members:    []models.Member{},
			},
			expected: &Namespace{
				ID:      "ns-id-3",
				Name:    "no-settings",
				OwnerID: "owner-id-3",
				Type:    "personal",
				Settings: NamespaceSettings{
					MaxDevices: 15,
				},
				Memberships: []Membership{},
			},
		},
		{
			name: "member with empty Role stores empty string",
			model: &models.Namespace{
				TenantID: "ns-id-5",
				Name:     "empty-role-ns",
				Owner:    "owner-id-5",
				Type:     models.TypeTeam,
				Members: []models.Member{
					{
						ID:   "user-id-5",
						Role: "",
					},
				},
			},
			expected: &Namespace{
				ID:      "ns-id-5",
				Name:    "empty-role-ns",
				OwnerID: "owner-id-5",
				Type:    "team",
				Memberships: []Membership{
					{
						UserID:      "user-id-5",
						NamespaceID: "ns-id-5",
						Role:        "",
					},
				},
			},
		},
		{
			name: "empty Members",
			model: &models.Namespace{
				TenantID: "ns-id-4",
				Name:     "empty-members",
				Owner:    "owner-id-4",
				Type:     models.TypeTeam,
				Members:  []models.Member{},
			},
			expected: &Namespace{
				ID:          "ns-id-4",
				Name:        "empty-members",
				OwnerID:     "owner-id-4",
				Type:        "team",
				Memberships: []Membership{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NamespaceFromModel(tt.model)
			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.Name, result.Name)
			assert.Equal(t, tt.expected.OwnerID, result.OwnerID)
			assert.Equal(t, tt.expected.Type, result.Type)
			assert.Equal(t, tt.expected.Settings.MaxDevices, result.Settings.MaxDevices)
			assert.Equal(t, tt.expected.Settings.SessionRecord, result.Settings.SessionRecord)
			assert.Equal(t, tt.expected.Settings.ConnectionAnnouncement, result.Settings.ConnectionAnnouncement)
			assert.Equal(t, tt.expected.DevicesAcceptedCount, result.DevicesAcceptedCount)
			assert.Equal(t, tt.expected.DevicesPendingCount, result.DevicesPendingCount)
			assert.Equal(t, tt.expected.DevicesRejectedCount, result.DevicesRejectedCount)
			assert.Equal(t, tt.expected.DevicesRemovedCount, result.DevicesRemovedCount)
			assert.Equal(t, tt.expected.CreatedAt, result.CreatedAt)
			assert.True(t, result.UpdatedAt.IsZero(), "UpdatedAt should be zero")
			require.Len(t, result.Memberships, len(tt.expected.Memberships))
			for i, m := range tt.expected.Memberships {
				assert.Equal(t, m.UserID, result.Memberships[i].UserID)
				assert.Equal(t, m.NamespaceID, result.Memberships[i].NamespaceID)
				assert.Equal(t, m.Role, result.Memberships[i].Role)
				assert.Equal(t, m.CreatedAt, result.Memberships[i].CreatedAt)
			}
		})
	}
}

func TestNamespaceToModel(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		entity   *Namespace
		expected *models.Namespace
	}{
		{
			name: "full fields",
			entity: &Namespace{
				ID:      "ns-id-1",
				Name:    "my-namespace",
				OwnerID: "owner-id-1",
				Type:    "team",
				Settings: NamespaceSettings{
					MaxDevices:             10,
					SessionRecord:          true,
					ConnectionAnnouncement: "Hello!",
				},
				Memberships: []Membership{
					{
						UserID:    "user-id-1",
						CreatedAt: now,
						Role:      "administrator",
					},
				},
				DevicesAcceptedCount: 5,
				DevicesPendingCount:  2,
				DevicesRejectedCount: 1,
				DevicesRemovedCount:  3,
				CreatedAt:            now,
			},
			expected: &models.Namespace{
				TenantID: "ns-id-1",
				Name:     "my-namespace",
				Owner:    "owner-id-1",
				Type:     models.TypeTeam,
				Settings: &models.NamespaceSettings{
					SessionRecord:          true,
					ConnectionAnnouncement: "Hello!",
				},
				MaxDevices: 10,
				Members: []models.Member{
					{
						ID:      "user-id-1",
						AddedAt: now,
						Role:    authorizer.RoleAdministrator,
					},
				},
				DevicesAcceptedCount: 5,
				DevicesPendingCount:  2,
				DevicesRejectedCount: 1,
				DevicesRemovedCount:  3,
				CreatedAt:            now,
			},
		},
		{
			name: "empty Memberships",
			entity: &Namespace{
				ID:          "ns-id-2",
				Name:        "empty-ns",
				OwnerID:     "owner-id-2",
				Type:        "personal",
				Memberships: []Membership{},
			},
			expected: &models.Namespace{
				TenantID: "ns-id-2",
				Name:     "empty-ns",
				Owner:    "owner-id-2",
				Type:     models.TypePersonal,
				Settings: &models.NamespaceSettings{},
				Members:  []models.Member{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NamespaceToModel(tt.entity)
			assert.Equal(t, tt.expected.TenantID, result.TenantID)
			assert.Equal(t, tt.expected.Name, result.Name)
			assert.Equal(t, tt.expected.Owner, result.Owner)
			assert.Equal(t, tt.expected.Type, result.Type)
			assert.Equal(t, tt.expected.MaxDevices, result.MaxDevices)
			require.NotNil(t, result.Settings, "Settings should never be nil")
			assert.Equal(t, tt.expected.Settings.SessionRecord, result.Settings.SessionRecord)
			assert.Equal(t, tt.expected.Settings.ConnectionAnnouncement, result.Settings.ConnectionAnnouncement)
			assert.Equal(t, tt.expected.DevicesAcceptedCount, result.DevicesAcceptedCount)
			assert.Equal(t, tt.expected.DevicesPendingCount, result.DevicesPendingCount)
			assert.Equal(t, tt.expected.DevicesRejectedCount, result.DevicesRejectedCount)
			assert.Equal(t, tt.expected.DevicesRemovedCount, result.DevicesRemovedCount)
			assert.Equal(t, tt.expected.CreatedAt, result.CreatedAt)
			require.Len(t, result.Members, len(tt.expected.Members))
			for i, m := range tt.expected.Members {
				assert.Equal(t, m.ID, result.Members[i].ID)
				assert.Equal(t, m.Role, result.Members[i].Role)
				assert.Equal(t, m.AddedAt, result.Members[i].AddedAt)
			}
		})
	}
}
