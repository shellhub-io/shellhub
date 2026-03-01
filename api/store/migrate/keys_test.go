package migrate

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConvertPublicKey(t *testing.T) {
	now := time.Now().Truncate(time.Second)

	t.Run("all fields populated", func(t *testing.T) {
		doc := mongoPublicKey{
			Fingerprint: "SHA256:aBcDeFgHiJkLmNoPqRsTuVwXyZ012345678901234",
			TenantID:    "tenant-123",
			Data:        []byte("ssh-rsa AAAA..."),
			CreatedAt:   now,
			Name:        "my-key",
			Username:    "root",
			Filter: mongoPKFilter{
				Hostname: "*.example.com",
			},
		}

		result := convertPublicKey(doc)

		assert.Equal(t, "SHA256:aBcDeFgHiJkLmNoPqRsTuVwXyZ012345678901234", result.Fingerprint)
		assert.Equal(t, "tenant-123", result.NamespaceID)
		assert.Equal(t, now, result.CreatedAt)
		assert.True(t, result.UpdatedAt.IsZero())
		assert.Equal(t, "my-key", result.Name)
		assert.Equal(t, "root", result.Username)
		assert.Equal(t, []byte("ssh-rsa AAAA..."), result.Data)
		assert.Equal(t, "*.example.com", result.FilterHostname)
	})

	t.Run("empty username", func(t *testing.T) {
		doc := mongoPublicKey{
			Fingerprint: "SHA256:abc",
			TenantID:    "tenant-123",
		}

		result := convertPublicKey(doc)

		assert.Empty(t, result.Username)
	})
}

func TestConvertAPIKey(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	updated := now.Add(time.Hour)

	t.Run("all fields populated", func(t *testing.T) {
		doc := mongoAPIKey{
			ID:        "sha256-digest-here",
			Name:      "my-api-key",
			TenantID:  "tenant-123",
			Role:      "admin",
			CreatedBy: "507f1f77bcf86cd799439011",
			CreatedAt: now,
			UpdatedAt: updated,
			ExpiresIn: 3600,
		}

		result := convertAPIKey(doc)

		assert.Equal(t, "sha256-digest-here", result.KeyDigest)
		assert.Equal(t, "tenant-123", result.NamespaceID)
		assert.Equal(t, "my-api-key", result.Name)
		assert.Equal(t, "admin", result.Role)
		assert.Equal(t, ObjectIDToUUID("507f1f77bcf86cd799439011"), result.UserID)
		assert.Equal(t, now, result.CreatedAt)
		assert.Equal(t, updated, result.UpdatedAt)
		assert.Equal(t, int64(3600), result.ExpiresIn)
	})
}
