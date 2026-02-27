package migrate

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConvertNamespace(t *testing.T) {
	now := time.Now().Truncate(time.Second)

	t.Run("all fields populated", func(t *testing.T) {
		doc := mongoNamespace{
			TenantID:             "00000000-0000-4000-0000-000000000000",
			Name:                 "my-namespace",
			Owner:                "507f1f77bcf86cd799439011",
			Type:                 "team",
			MaxDevices:           10,
			CreatedAt:            now,
			DevicesAcceptedCount: 5,
			DevicesPendingCount:  2,
			DevicesRejectedCount: 1,
			DevicesRemovedCount:  3,
			Settings: &mongoNSSettings{
				SessionRecord:          true,
				ConnectionAnnouncement: "Welcome!",
			},
		}

		result := convertNamespace(doc)

		assert.Equal(t, "00000000-0000-4000-0000-000000000000", result.ID)
		assert.Equal(t, now, result.CreatedAt)
		assert.True(t, result.UpdatedAt.IsZero())
		assert.Equal(t, "team", result.Type)
		assert.Equal(t, "my-namespace", result.Name)
		assert.Equal(t, ObjectIDToUUID("507f1f77bcf86cd799439011"), result.OwnerID)
		assert.Equal(t, int64(5), result.DevicesAcceptedCount)
		assert.Equal(t, int64(2), result.DevicesPendingCount)
		assert.Equal(t, int64(1), result.DevicesRejectedCount)
		assert.Equal(t, int64(3), result.DevicesRemovedCount)
		assert.Equal(t, 10, result.Settings.MaxDevices)
		assert.True(t, result.Settings.SessionRecord)
		assert.Equal(t, "Welcome!", result.Settings.ConnectionAnnouncement)
	})

	t.Run("defaults for empty type", func(t *testing.T) {
		doc := mongoNamespace{
			TenantID: "00000000-0000-4000-0000-000000000000",
			Owner:    "507f1f77bcf86cd799439011",
		}

		result := convertNamespace(doc)

		assert.Equal(t, "personal", result.Type)
	})

	t.Run("nil settings", func(t *testing.T) {
		doc := mongoNamespace{
			TenantID:   "00000000-0000-4000-0000-000000000000",
			Owner:      "507f1f77bcf86cd799439011",
			MaxDevices: 3,
		}

		result := convertNamespace(doc)

		assert.Equal(t, 3, result.Settings.MaxDevices)
		assert.False(t, result.Settings.SessionRecord)
		assert.Empty(t, result.Settings.ConnectionAnnouncement)
	})
}

func TestConvertMembership(t *testing.T) {
	now := time.Now().Truncate(time.Second)

	t.Run("all fields populated", func(t *testing.T) {
		member := mongoMember{
			ID:      "507f1f77bcf86cd799439011",
			AddedAt: now,
			Role:    "administrator",
		}

		result := convertMembership("tenant-123", member)

		assert.Equal(t, ObjectIDToUUID("507f1f77bcf86cd799439011"), result.UserID)
		assert.Equal(t, "tenant-123", result.NamespaceID)
		assert.Equal(t, now, result.CreatedAt)
		assert.True(t, result.UpdatedAt.IsZero())
		assert.Equal(t, "administrator", result.Role)
	})

	t.Run("defaults for empty role", func(t *testing.T) {
		member := mongoMember{
			ID: "507f1f77bcf86cd799439011",
		}

		result := convertMembership("tenant-123", member)

		assert.Equal(t, "observer", result.Role)
	})
}
