package migrate

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestConvertUser(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	oid := primitive.NewObjectID()

	t.Run("all fields populated", func(t *testing.T) {
		doc := mongoUser{
			ID:             oid,
			Origin:         "saml",
			ExternalID:     "ext-123",
			Status:         "active",
			MaxNamespaces:  5,
			CreatedAt:      now,
			LastLogin:      now.Add(time.Hour),
			EmailMarketing: true,
			Name:           "John Doe",
			Username:       "johndoe",
			Email:          "john@example.com",
			RecoveryEmail:  "recovery@example.com",
			Password:       "$2a$12$hashedpassword",
			Admin:          true,
			Preferences: mongoUserPrefs{
				PreferredNamespace: "00000000-0000-4000-0000-000000000000",
				AuthMethods:        []string{"saml", "local"},
			},
		}

		result := convertUser(doc)

		assert.Equal(t, ObjectIDToUUID(oid.Hex()), result.ID)
		assert.Equal(t, now, result.CreatedAt)
		assert.True(t, result.UpdatedAt.IsZero())
		assert.Equal(t, now.Add(time.Hour), result.LastLogin)
		assert.Equal(t, "saml", result.Origin)
		assert.Equal(t, "ext-123", result.ExternalID)
		assert.Equal(t, "active", result.Status)
		assert.Equal(t, "John Doe", result.Name)
		assert.Equal(t, "johndoe", result.Username)
		assert.Equal(t, "john@example.com", result.Email)
		assert.Equal(t, "$2a$12$hashedpassword", result.PasswordDigest)
		assert.True(t, result.Admin)
		assert.Equal(t, "00000000-0000-4000-0000-000000000000", result.Preferences.PreferredNamespace)
		assert.Equal(t, []string{"saml", "local"}, result.Preferences.AuthMethods)
		assert.Equal(t, "recovery@example.com", result.Preferences.SecurityEmail)
		assert.Equal(t, 5, result.Preferences.MaxNamespaces)
		assert.True(t, result.Preferences.EmailMarketing)
	})

	t.Run("defaults for empty fields", func(t *testing.T) {
		doc := mongoUser{
			ID: oid,
		}

		result := convertUser(doc)

		assert.Equal(t, "local", result.Origin)
		assert.Equal(t, "confirmed", result.Status)
		assert.Equal(t, []string{"local"}, result.Preferences.AuthMethods)
	})

	t.Run("preserves explicit origin and status", func(t *testing.T) {
		doc := mongoUser{
			ID:     oid,
			Origin: "github",
			Status: "disabled",
			Preferences: mongoUserPrefs{
				AuthMethods: []string{"github"},
			},
		}

		result := convertUser(doc)

		assert.Equal(t, "github", result.Origin)
		assert.Equal(t, "disabled", result.Status)
		assert.Equal(t, []string{"github"}, result.Preferences.AuthMethods)
	})
}
