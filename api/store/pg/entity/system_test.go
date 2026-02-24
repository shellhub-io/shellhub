package entity

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSystemFromModel(t *testing.T) {
	tests := []struct {
		name  string
		model *models.System
		check func(t *testing.T, result *System)
	}{
		{
			name:  "nil input",
			model: nil,
			check: func(t *testing.T, result *System) {
				require.NotNil(t, result)
				assert.False(t, result.Setup)
			},
		},
		{
			name: "full fields with all nesting",
			model: &models.System{
				Setup: true,
				Authentication: &models.SystemAuthentication{
					Local: &models.SystemAuthenticationLocal{
						Enabled: true,
					},
					SAML: &models.SystemAuthenticationSAML{
						Enabled: true,
						Idp: &models.SystemIdpSAML{
							EntityID:     "https://idp.example.com",
							Certificates: []string{"cert1", "cert2"},
							Mappings:     map[string]string{"email": "user.email"},
							Binding: &models.SystemAuthenticationBinding{
								Post:      "https://idp.example.com/post",
								Redirect:  "https://idp.example.com/redirect",
								Preferred: "post",
							},
						},
						Sp: &models.SystemSpSAML{
							SignAuthRequests: true,
							Certificate:      "sp-cert",
							PrivateKey:       "sp-key",
						},
					},
				},
			},
			check: func(t *testing.T, result *System) {
				assert.True(t, result.Setup)
				assert.True(t, result.Authentication.Local.Enabled)
				assert.True(t, result.Authentication.SAML.Enabled)
				assert.Equal(t, "https://idp.example.com", result.Authentication.SAML.Idp.EntityID)
				assert.Equal(t, []string{"cert1", "cert2"}, result.Authentication.SAML.Idp.Certificates)
				assert.Equal(t, map[string]string{"email": "user.email"}, result.Authentication.SAML.Idp.Mappings)
				assert.Equal(t, "https://idp.example.com/post", result.Authentication.SAML.Idp.Binding.Post)
				assert.Equal(t, "https://idp.example.com/redirect", result.Authentication.SAML.Idp.Binding.Redirect)
				assert.Equal(t, "post", result.Authentication.SAML.Idp.Binding.Preferred)
				assert.True(t, result.Authentication.SAML.Sp.SignAuthRequests)
				assert.Equal(t, "sp-cert", result.Authentication.SAML.Sp.Certificate)
				assert.Equal(t, "sp-key", result.Authentication.SAML.Sp.PrivateKey)
			},
		},
		{
			name: "nil Authentication",
			model: &models.System{
				Setup:          true,
				Authentication: nil,
			},
			check: func(t *testing.T, result *System) {
				assert.True(t, result.Setup)
				assert.False(t, result.Authentication.Local.Enabled)
				assert.False(t, result.Authentication.SAML.Enabled)
			},
		},
		{
			name: "nil Local",
			model: &models.System{
				Authentication: &models.SystemAuthentication{
					Local: nil,
					SAML:  &models.SystemAuthenticationSAML{Enabled: true},
				},
			},
			check: func(t *testing.T, result *System) {
				assert.False(t, result.Authentication.Local.Enabled)
				assert.True(t, result.Authentication.SAML.Enabled)
			},
		},
		{
			name: "nil SAML",
			model: &models.System{
				Authentication: &models.SystemAuthentication{
					Local: &models.SystemAuthenticationLocal{Enabled: true},
					SAML:  nil,
				},
			},
			check: func(t *testing.T, result *System) {
				assert.True(t, result.Authentication.Local.Enabled)
				assert.False(t, result.Authentication.SAML.Enabled)
			},
		},
		{
			name: "nil Idp",
			model: &models.System{
				Authentication: &models.SystemAuthentication{
					SAML: &models.SystemAuthenticationSAML{
						Enabled: true,
						Idp:     nil,
						Sp:      &models.SystemSpSAML{Certificate: "cert"},
					},
				},
			},
			check: func(t *testing.T, result *System) {
				assert.True(t, result.Authentication.SAML.Enabled)
				assert.Equal(t, "", result.Authentication.SAML.Idp.EntityID)
				assert.Equal(t, "cert", result.Authentication.SAML.Sp.Certificate)
			},
		},
		{
			name: "nil Idp.Binding",
			model: &models.System{
				Authentication: &models.SystemAuthentication{
					SAML: &models.SystemAuthenticationSAML{
						Enabled: true,
						Idp: &models.SystemIdpSAML{
							EntityID: "entity-1",
							Binding:  nil,
						},
					},
				},
			},
			check: func(t *testing.T, result *System) {
				assert.Equal(t, "entity-1", result.Authentication.SAML.Idp.EntityID)
				assert.Equal(t, "", result.Authentication.SAML.Idp.Binding.Post)
				assert.Equal(t, "", result.Authentication.SAML.Idp.Binding.Redirect)
			},
		},
		{
			name: "nil Sp",
			model: &models.System{
				Authentication: &models.SystemAuthentication{
					SAML: &models.SystemAuthenticationSAML{
						Enabled: true,
						Idp:     &models.SystemIdpSAML{EntityID: "entity-1"},
						Sp:      nil,
					},
				},
			},
			check: func(t *testing.T, result *System) {
				assert.Equal(t, "entity-1", result.Authentication.SAML.Idp.EntityID)
				assert.False(t, result.Authentication.SAML.Sp.SignAuthRequests)
				assert.Equal(t, "", result.Authentication.SAML.Sp.Certificate)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SystemFromModel(tt.model)
			tt.check(t, result)
		})
	}
}

func TestSystemToModel(t *testing.T) {
	tests := []struct {
		name   string
		entity *System
		check  func(t *testing.T, result *models.System)
	}{
		{
			name:   "nil input",
			entity: nil,
			check: func(t *testing.T, result *models.System) {
				require.NotNil(t, result)
				assert.False(t, result.Setup)
				assert.Nil(t, result.Authentication)
			},
		},
		{
			name: "zero-value authentication sub-structs",
			entity: &System{
				Setup: true,
			},
			check: func(t *testing.T, result *models.System) {
				assert.True(t, result.Setup)
				require.NotNil(t, result.Authentication)
				require.NotNil(t, result.Authentication.Local)
				assert.False(t, result.Authentication.Local.Enabled)
				require.NotNil(t, result.Authentication.SAML)
				assert.False(t, result.Authentication.SAML.Enabled)
				require.NotNil(t, result.Authentication.SAML.Idp)
				assert.Equal(t, "", result.Authentication.SAML.Idp.EntityID)
				require.NotNil(t, result.Authentication.SAML.Idp.Binding)
				assert.Equal(t, "", result.Authentication.SAML.Idp.Binding.Post)
				require.NotNil(t, result.Authentication.SAML.Sp)
				assert.False(t, result.Authentication.SAML.Sp.SignAuthRequests)
			},
		},
		{
			name: "only SAML enabled",
			entity: &System{
				Setup: true,
				Authentication: SystemAuthentication{
					Local: SystemAuthenticationLocal{Enabled: false},
					SAML: SystemAuthenticationSAML{
						Enabled: true,
						Idp: SystemIdpSAML{
							EntityID: "https://idp.example.com",
						},
					},
				},
			},
			check: func(t *testing.T, result *models.System) {
				assert.True(t, result.Setup)
				require.NotNil(t, result.Authentication)
				require.NotNil(t, result.Authentication.Local)
				assert.False(t, result.Authentication.Local.Enabled)
				require.NotNil(t, result.Authentication.SAML)
				assert.True(t, result.Authentication.SAML.Enabled)
				assert.Equal(t, "https://idp.example.com", result.Authentication.SAML.Idp.EntityID)
				assert.False(t, result.Authentication.SAML.Sp.SignAuthRequests)
			},
		},
		{
			name: "full fields",
			entity: &System{
				Setup: true,
				Authentication: SystemAuthentication{
					Local: SystemAuthenticationLocal{Enabled: true},
					SAML: SystemAuthenticationSAML{
						Enabled: true,
						Idp: SystemIdpSAML{
							EntityID:     "https://idp.example.com",
							Certificates: []string{"cert1"},
							Mappings:     map[string]string{"email": "user.email"},
							Binding: SystemAuthenticationBinding{
								Post:      "https://idp.example.com/post",
								Redirect:  "https://idp.example.com/redirect",
								Preferred: "post",
							},
						},
						Sp: SystemSpSAML{
							SignAuthRequests: true,
							Certificate:      "sp-cert",
							PrivateKey:       "sp-key",
						},
					},
				},
			},
			check: func(t *testing.T, result *models.System) {
				assert.True(t, result.Setup)
				require.NotNil(t, result.Authentication)
				require.NotNil(t, result.Authentication.Local)
				assert.True(t, result.Authentication.Local.Enabled)
				require.NotNil(t, result.Authentication.SAML)
				assert.True(t, result.Authentication.SAML.Enabled)
				require.NotNil(t, result.Authentication.SAML.Idp)
				assert.Equal(t, "https://idp.example.com", result.Authentication.SAML.Idp.EntityID)
				require.NotNil(t, result.Authentication.SAML.Idp.Binding)
				assert.Equal(t, "https://idp.example.com/post", result.Authentication.SAML.Idp.Binding.Post)
				require.NotNil(t, result.Authentication.SAML.Sp)
				assert.True(t, result.Authentication.SAML.Sp.SignAuthRequests)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SystemToModel(tt.entity)
			tt.check(t, result)
		})
	}
}
