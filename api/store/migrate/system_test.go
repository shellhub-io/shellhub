package migrate

import (
	"testing"

	"github.com/google/uuid" //nolint:depguard
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertSystem(t *testing.T) {
	t.Run("minimal system", func(t *testing.T) {
		doc := mongoSystem{
			Setup: true,
		}

		result := convertSystem(doc)

		_, err := uuid.Parse(result.ID)
		require.NoError(t, err)
		assert.True(t, result.Setup)
		assert.False(t, result.Authentication.Local.Enabled)
		assert.False(t, result.Authentication.SAML.Enabled)
	})

	t.Run("nil authentication", func(t *testing.T) {
		doc := mongoSystem{
			Setup:          false,
			Authentication: nil,
		}

		result := convertSystem(doc)

		assert.False(t, result.Setup)
		assert.False(t, result.Authentication.Local.Enabled)
	})

	t.Run("local auth only", func(t *testing.T) {
		doc := mongoSystem{
			Setup: true,
			Authentication: &mongoSystemAuth{
				Local: &mongoSystemAuthLocal{Enabled: true},
			},
		}

		result := convertSystem(doc)

		assert.True(t, result.Authentication.Local.Enabled)
		assert.False(t, result.Authentication.SAML.Enabled)
	})

	t.Run("full SAML configuration", func(t *testing.T) {
		doc := mongoSystem{
			Setup: true,
			Authentication: &mongoSystemAuth{
				Local: &mongoSystemAuthLocal{Enabled: true},
				SAML: &mongoSystemAuthSAML{
					Enabled: true,
					Idp: &mongoSystemIdp{
						EntityID:     "https://idp.example.com",
						Certificates: []string{"cert1", "cert2"},
						Mappings:     map[string]string{"email": "mail", "name": "displayName"},
						Binding: &mongoSystemBinding{
							Post:      "https://idp.example.com/sso/post",
							Redirect:  "https://idp.example.com/sso/redirect",
							Preferred: "post",
						},
					},
					Sp: &mongoSystemSp{
						SignAuthRequests: true,
						Certificate:      "sp-cert",
						PrivateKey:       "sp-key",
					},
				},
			},
		}

		result := convertSystem(doc)

		assert.True(t, result.Authentication.Local.Enabled)
		assert.True(t, result.Authentication.SAML.Enabled)
		assert.Equal(t, "https://idp.example.com", result.Authentication.SAML.Idp.EntityID)
		assert.Equal(t, []string{"cert1", "cert2"}, result.Authentication.SAML.Idp.Certificates)
		assert.Equal(t, map[string]string{"email": "mail", "name": "displayName"}, result.Authentication.SAML.Idp.Mappings)
		assert.Equal(t, "https://idp.example.com/sso/post", result.Authentication.SAML.Idp.Binding.Post)
		assert.Equal(t, "https://idp.example.com/sso/redirect", result.Authentication.SAML.Idp.Binding.Redirect)
		assert.Equal(t, "post", result.Authentication.SAML.Idp.Binding.Preferred)
		assert.True(t, result.Authentication.SAML.Sp.SignAuthRequests)
		assert.Equal(t, "sp-cert", result.Authentication.SAML.Sp.Certificate)
		assert.Equal(t, "sp-key", result.Authentication.SAML.Sp.PrivateKey)
	})

	t.Run("SAML without binding and sp", func(t *testing.T) {
		doc := mongoSystem{
			Setup: true,
			Authentication: &mongoSystemAuth{
				SAML: &mongoSystemAuthSAML{
					Enabled: true,
					Idp: &mongoSystemIdp{
						EntityID: "https://idp.example.com",
					},
				},
			},
		}

		result := convertSystem(doc)

		assert.True(t, result.Authentication.SAML.Enabled)
		assert.Equal(t, "https://idp.example.com", result.Authentication.SAML.Idp.EntityID)
		assert.Empty(t, result.Authentication.SAML.Idp.Binding.Post)
		assert.False(t, result.Authentication.SAML.Sp.SignAuthRequests)
	})

	t.Run("unique IDs", func(t *testing.T) {
		doc := mongoSystem{Setup: true}
		a := convertSystem(doc)
		b := convertSystem(doc)
		assert.NotEqual(t, a.ID, b.ID)
	})
}
