package storetest

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSystemGet tests getting system configuration
func (s *Suite) TestSystemGet(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("returns existing system configuration", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create a system configuration
		expectedSystem := &models.System{
			Setup: true,
			Authentication: &models.SystemAuthentication{
				Local: &models.SystemAuthenticationLocal{
					Enabled: true,
				},
				SAML: &models.SystemAuthenticationSAML{
					Enabled: false,
					Idp:     &models.SystemIdpSAML{Binding: &models.SystemAuthenticationBinding{}},
					Sp:      &models.SystemSpSAML{},
				},
			},
		}

		err := st.SystemSet(ctx, expectedSystem)
		require.NoError(t, err)

		// Get the system
		system, err := st.SystemGet(ctx)
		require.NoError(t, err)
		require.NotNil(t, system)

		assert.True(t, system.Setup)
		assert.NotNil(t, system.Authentication)
		assert.True(t, system.Authentication.Local.Enabled)
	})
}

// TestSystemSet tests setting system configuration
func (s *Suite) TestSystemSet(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("creates system when none exists", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		system := &models.System{
			Setup: true,
			Authentication: &models.SystemAuthentication{
				Local: &models.SystemAuthenticationLocal{
					Enabled: true,
				},
				SAML: &models.SystemAuthenticationSAML{
					Enabled: false,
					Idp:     &models.SystemIdpSAML{Binding: &models.SystemAuthenticationBinding{}},
					Sp:      &models.SystemSpSAML{},
				},
			},
		}

		err := st.SystemSet(ctx, system)
		require.NoError(t, err)

		// Verify it was created
		created, err := st.SystemGet(ctx)
		require.NoError(t, err)
		assert.True(t, created.Setup)
	})

	t.Run("updates existing system", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create initial system
		initialSystem := &models.System{
			Setup: false,
			Authentication: &models.SystemAuthentication{
				Local: &models.SystemAuthenticationLocal{
					Enabled: true,
				},
				SAML: &models.SystemAuthenticationSAML{
					Enabled: false,
					Idp:     &models.SystemIdpSAML{Binding: &models.SystemAuthenticationBinding{}},
					Sp:      &models.SystemSpSAML{},
				},
			},
		}

		err := st.SystemSet(ctx, initialSystem)
		require.NoError(t, err)

		// Update system
		updatedSystem := &models.System{
			Setup: true,
			Authentication: &models.SystemAuthentication{
				Local: &models.SystemAuthenticationLocal{
					Enabled: false,
				},
				SAML: &models.SystemAuthenticationSAML{
					Enabled: true,
					Idp:     &models.SystemIdpSAML{Binding: &models.SystemAuthenticationBinding{}},
					Sp:      &models.SystemSpSAML{},
				},
			},
		}

		err = st.SystemSet(ctx, updatedSystem)
		require.NoError(t, err)

		// Verify update
		system, err := st.SystemGet(ctx)
		require.NoError(t, err)
		assert.True(t, system.Setup)
		assert.False(t, system.Authentication.Local.Enabled)
	})

	t.Run("succeeds with minimal system data", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		system := &models.System{
			Setup: true,
			Authentication: &models.SystemAuthentication{
				Local: &models.SystemAuthenticationLocal{
					Enabled: true,
				},
				SAML: &models.SystemAuthenticationSAML{
					Enabled: false,
					Idp:     &models.SystemIdpSAML{Binding: &models.SystemAuthenticationBinding{}},
					Sp:      &models.SystemSpSAML{},
				},
			},
		}

		err := st.SystemSet(ctx, system)
		require.NoError(t, err)

		// Verify
		created, err := st.SystemGet(ctx)
		require.NoError(t, err)
		assert.NotNil(t, created)
	})
}
