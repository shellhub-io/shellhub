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
			name: "full fields",
			model: &models.System{
				Setup: true,
				Authentication: &models.SystemAuthentication{
					Local: &models.SystemAuthenticationLocal{
						Enabled: true,
					},
				},
			},
			check: func(t *testing.T, result *System) {
				assert.True(t, result.Setup)
				assert.True(t, result.Authentication.Local.Enabled)
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
			},
		},
		{
			name: "nil Local",
			model: &models.System{
				Authentication: &models.SystemAuthentication{
					Local: nil,
				},
			},
			check: func(t *testing.T, result *System) {
				assert.False(t, result.Authentication.Local.Enabled)
			},
		},
		{
			name: "local disabled",
			model: &models.System{
				Authentication: &models.SystemAuthentication{
					Local: &models.SystemAuthenticationLocal{Enabled: false},
				},
			},
			check: func(t *testing.T, result *System) {
				assert.False(t, result.Authentication.Local.Enabled)
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
			name: "zero-value authentication",
			entity: &System{
				Setup: true,
			},
			check: func(t *testing.T, result *models.System) {
				assert.True(t, result.Setup)
				require.NotNil(t, result.Authentication)
				require.NotNil(t, result.Authentication.Local)
				assert.False(t, result.Authentication.Local.Enabled)
			},
		},
		{
			name: "local enabled",
			entity: &System{
				Setup: true,
				Authentication: SystemAuthentication{
					Local: SystemAuthenticationLocal{Enabled: true},
				},
			},
			check: func(t *testing.T, result *models.System) {
				assert.True(t, result.Setup)
				require.NotNil(t, result.Authentication)
				require.NotNil(t, result.Authentication.Local)
				assert.True(t, result.Authentication.Local.Enabled)
			},
		},
		{
			name: "local disabled",
			entity: &System{
				Setup: false,
				Authentication: SystemAuthentication{
					Local: SystemAuthenticationLocal{Enabled: false},
				},
			},
			check: func(t *testing.T, result *models.System) {
				assert.False(t, result.Setup)
				require.NotNil(t, result.Authentication)
				require.NotNil(t, result.Authentication.Local)
				assert.False(t, result.Authentication.Local.Enabled)
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
