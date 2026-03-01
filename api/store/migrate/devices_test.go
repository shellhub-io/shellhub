package migrate

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConvertDevice(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	disconnected := now.Add(-time.Hour)
	removed := now.Add(-2 * time.Hour)

	t.Run("all fields populated", func(t *testing.T) {
		doc := mongoDevice{
			UID:             "device-uid-123",
			CreatedAt:       now,
			RemovedAt:       &removed,
			Name:            "my-device",
			Identity:        &mongoDeviceID{MAC: "aa:bb:cc:dd:ee:ff"},
			Info:            &mongoDeviceInfo{ID: "id-1", PrettyName: "My Device", Version: "0.14.0", Arch: "amd64", Platform: "linux"},
			PublicKey:       "ssh-rsa AAAA...",
			TenantID:        "tenant-123",
			LastSeen:        now,
			DisconnectedAt:  &disconnected,
			Status:          "accepted",
			StatusUpdatedAt: now,
			Position:        &mongoDevicePos{Latitude: -23.5, Longitude: -46.6},
		}

		result := convertDevice(doc)

		assert.Equal(t, "device-uid-123", result.ID)
		assert.Equal(t, "tenant-123", result.NamespaceID)
		assert.Equal(t, now, result.CreatedAt)
		assert.True(t, result.UpdatedAt.IsZero())
		assert.Equal(t, &removed, result.RemovedAt)
		assert.Equal(t, now, result.LastSeen)
		assert.Equal(t, "accepted", result.Status)
		assert.Equal(t, now, result.StatusUpdatedAt)
		assert.Equal(t, "my-device", result.Name)
		assert.Equal(t, "ssh-rsa AAAA...", result.PublicKey)
		assert.Equal(t, disconnected, result.DisconnectedAt)
		assert.Equal(t, "aa:bb:cc:dd:ee:ff", result.MAC)
		assert.Equal(t, "id-1", result.Identifier)
		assert.Equal(t, "My Device", result.PrettyName)
		assert.Equal(t, "0.14.0", result.Version)
		assert.Equal(t, "amd64", result.Arch)
		assert.Equal(t, "linux", result.Platform)
		assert.InDelta(t, -46.6, result.Longitude, 0.001)
		assert.InDelta(t, -23.5, result.Latitude, 0.001)
	})

	t.Run("nil optional fields", func(t *testing.T) {
		doc := mongoDevice{
			UID:      "device-uid-456",
			TenantID: "tenant-123",
			Status:   "accepted",
		}

		result := convertDevice(doc)

		assert.True(t, result.DisconnectedAt.IsZero())
		assert.Empty(t, result.MAC)
		assert.Empty(t, result.Identifier)
		assert.Empty(t, result.PrettyName)
		assert.Zero(t, result.Longitude)
		assert.Zero(t, result.Latitude)
		assert.Nil(t, result.RemovedAt)
	})

	t.Run("defaults for empty status", func(t *testing.T) {
		doc := mongoDevice{
			UID:      "device-uid-789",
			TenantID: "tenant-123",
		}

		result := convertDevice(doc)

		assert.Equal(t, "pending", result.Status)
	})
}
