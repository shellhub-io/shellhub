package entity

import (
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestUserFromModel(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		model    *models.User
		expected *User
	}{
		{
			name: "full fields",
			model: &models.User{
				ID:             "user-id-1",
				Origin:         models.UserOriginLocal,
				ExternalID:     "ext-id-1",
				Status:         models.UserStatusConfirmed,
				MaxNamespaces:  5,
				CreatedAt:      now,
				LastLogin:      now.Add(-time.Hour),
				EmailMarketing: true,
				UserData: models.UserData{
					Name:          "John Doe",
					Username:      "johndoe",
					Email:         "john@example.com",
					RecoveryEmail: "recovery@example.com",
				},
				Password: models.UserPassword{
					Hash: "hashed-password-123",
				},
				MFA: models.UserMFA{
					Enabled:       true,
					Secret:        "mfa-secret",
					RecoveryCodes: []string{"code1", "code2"},
				},
				Preferences: models.UserPreferences{
					PreferredNamespace: "ns-id-1",
					AuthMethods:        []models.UserAuthMethod{models.UserAuthMethodLocal, models.UserAuthMethodSAML},
				},
				Admin: true,
			},
			expected: &User{
				ID:             "user-id-1",
				Origin:         "local",
				ExternalID:     "ext-id-1",
				Status:         "confirmed",
				Name:           "John Doe",
				Username:       "johndoe",
				Email:          "john@example.com",
				PasswordDigest: "hashed-password-123",
				Admin:          true,
				CreatedAt:      now,
				LastLogin:      now.Add(-time.Hour),
				Preferences: UserPreferences{
					PreferredNamespace: "ns-id-1",
					AuthMethods:        []string{"local", "saml"},
					SecurityEmail:      "recovery@example.com",
					MaxNamespaces:      5,
					EmailMarketing:     true,
				},
				MFA: UserMFA{
					Enabled:       true,
					Secret:        "mfa-secret",
					RecoveryCodes: []string{"code1", "code2"},
				},
			},
		},
		{
			name: "empty Origin defaults to local",
			model: &models.User{
				ID:     "user-id-2",
				Origin: "",
				Status: models.UserStatusConfirmed,
			},
			expected: &User{
				ID:     "user-id-2",
				Origin: "local",
				Status: "confirmed",
				Preferences: UserPreferences{
					AuthMethods: []string{},
				},
			},
		},
		{
			name: "empty Status defaults to confirmed",
			model: &models.User{
				ID:     "user-id-3",
				Origin: models.UserOriginLocal,
				Status: "",
			},
			expected: &User{
				ID:     "user-id-3",
				Origin: "local",
				Status: "confirmed",
				Preferences: UserPreferences{
					AuthMethods: []string{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UserFromModel(tt.model)
			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.Origin, result.Origin)
			assert.Equal(t, tt.expected.ExternalID, result.ExternalID)
			assert.Equal(t, tt.expected.Status, result.Status)
			assert.Equal(t, tt.expected.Name, result.Name)
			assert.Equal(t, tt.expected.Username, result.Username)
			assert.Equal(t, tt.expected.Email, result.Email)
			assert.Equal(t, tt.expected.PasswordDigest, result.PasswordDigest)
			assert.Equal(t, tt.expected.Admin, result.Admin)
			assert.Equal(t, tt.expected.CreatedAt, result.CreatedAt)
			assert.Equal(t, tt.expected.LastLogin, result.LastLogin)
			assert.Equal(t, tt.expected.Preferences.PreferredNamespace, result.Preferences.PreferredNamespace)
			assert.Equal(t, tt.expected.Preferences.AuthMethods, result.Preferences.AuthMethods)
			assert.Equal(t, tt.expected.Preferences.SecurityEmail, result.Preferences.SecurityEmail)
			assert.Equal(t, tt.expected.Preferences.MaxNamespaces, result.Preferences.MaxNamespaces)
			assert.Equal(t, tt.expected.Preferences.EmailMarketing, result.Preferences.EmailMarketing)
			assert.Equal(t, tt.expected.MFA.Enabled, result.MFA.Enabled)
			assert.Equal(t, tt.expected.MFA.Secret, result.MFA.Secret)
			assert.Equal(t, tt.expected.MFA.RecoveryCodes, result.MFA.RecoveryCodes)
			assert.True(t, result.UpdatedAt.IsZero(), "UpdatedAt should be zero")
		})
	}
}

func TestUserToModel(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		entity   *User
		expected *models.User
	}{
		{
			name: "full fields with UserData disaggregation",
			entity: &User{
				ID:             "user-id-1",
				Origin:         "local",
				ExternalID:     "ext-id-1",
				Status:         "confirmed",
				Name:           "John Doe",
				Username:       "johndoe",
				Email:          "john@example.com",
				PasswordDigest: "hashed-password-123",
				Admin:          true,
				CreatedAt:      now,
				LastLogin:      now.Add(-time.Hour),
				Preferences: UserPreferences{
					PreferredNamespace: "ns-id-1",
					AuthMethods:        []string{"local", "saml"},
					SecurityEmail:      "recovery@example.com",
					MaxNamespaces:      5,
					EmailMarketing:     true,
				},
				MFA: UserMFA{
					Enabled:       true,
					Secret:        "mfa-secret",
					RecoveryCodes: []string{"code1", "code2"},
				},
			},
			expected: &models.User{
				ID:             "user-id-1",
				Origin:         models.UserOriginLocal,
				ExternalID:     "ext-id-1",
				Status:         models.UserStatusConfirmed,
				MaxNamespaces:  5,
				CreatedAt:      now,
				LastLogin:      now.Add(-time.Hour),
				EmailMarketing: true,
				Admin:          true,
				UserData: models.UserData{
					Name:          "John Doe",
					Username:      "johndoe",
					Email:         "john@example.com",
					RecoveryEmail: "recovery@example.com",
				},
				Password: models.UserPassword{
					Hash: "hashed-password-123",
				},
				MFA: models.UserMFA{
					Enabled:       true,
					Secret:        "mfa-secret",
					RecoveryCodes: []string{"code1", "code2"},
				},
				Preferences: models.UserPreferences{
					PreferredNamespace: "ns-id-1",
					AuthMethods:        []models.UserAuthMethod{"local", "saml"},
				},
			},
		},
		{
			name: "SAML origin with non-confirmed status and empty auth methods",
			entity: &User{
				ID:     "user-id-2",
				Origin: "SAML",
				Status: "not-confirmed",
				Preferences: UserPreferences{
					AuthMethods: []string{},
				},
			},
			expected: &models.User{
				ID:       "user-id-2",
				Origin:   models.UserOriginSAML,
				Status:   models.UserStatusNotConfirmed,
				UserData: models.UserData{},
				Password: models.UserPassword{},
				MFA:      models.UserMFA{},
				Preferences: models.UserPreferences{
					AuthMethods: []models.UserAuthMethod{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UserToModel(tt.entity)
			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.Origin, result.Origin)
			assert.Equal(t, tt.expected.ExternalID, result.ExternalID)
			assert.Equal(t, tt.expected.Status, result.Status)
			assert.Equal(t, tt.expected.MaxNamespaces, result.MaxNamespaces)
			assert.Equal(t, tt.expected.CreatedAt, result.CreatedAt)
			assert.Equal(t, tt.expected.LastLogin, result.LastLogin)
			assert.Equal(t, tt.expected.EmailMarketing, result.EmailMarketing)
			assert.Equal(t, tt.expected.Admin, result.Admin)
			// UserData disaggregation
			assert.Equal(t, tt.expected.Name, result.Name)
			assert.Equal(t, tt.expected.Username, result.Username)
			assert.Equal(t, tt.expected.Email, result.Email)
			assert.Equal(t, tt.expected.RecoveryEmail, result.RecoveryEmail)
			// Password
			assert.Equal(t, tt.expected.Password.Hash, result.Password.Hash)
			// MFA
			assert.Equal(t, tt.expected.MFA, result.MFA)
			// Preferences
			assert.Equal(t, tt.expected.Preferences, result.Preferences)
			// entity.Namespaces is scanonly (populated by DB queries) and intentionally not mapped by UserToModel.
		})
	}
}
