package storetest

import (
	"context"
	"fmt"
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

		expectedSystem := &models.System{
			Setup: true,
			Authentication: &models.SystemAuthentication{
				Local: &models.SystemAuthenticationLocal{
					Enabled: true,
				},
			},
		}

		err := st.SystemSet(ctx, expectedSystem)
		require.NoError(t, err)

		system, err := st.SystemGet(ctx)
		require.NoError(t, err)
		require.NotNil(t, system)

		assert.True(t, system.Setup)
		assert.NotNil(t, system.Authentication)
		assert.True(t, system.Authentication.Local.Enabled)
	})

	t.Run("requires verified device keys when the system written leaves device authentication out", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		require.NoError(t, st.SystemSet(ctx, &models.System{
			Authentication: &models.SystemAuthentication{Local: &models.SystemAuthenticationLocal{Enabled: false}},
		}))

		system, err := st.SystemGet(ctx)
		require.NoError(t, err)
		require.NotNil(t, system.Authentication.Device)
		assert.True(t, system.Authentication.Device.RequireVerifiedKey, "a row first written without device settings belongs to a new instance")
	})

	t.Run("reads back whether device keys must be verified", func(t *testing.T) {
		for _, required := range []bool{true, false} {
			t.Run(fmt.Sprintf("require verified key %t", required), func(t *testing.T) {
				require.NoError(t, s.provider.CleanDatabase(t))

				require.NoError(t, st.SystemSet(ctx, &models.System{
					Setup: true,
					Authentication: &models.SystemAuthentication{
						Local:  &models.SystemAuthenticationLocal{Enabled: true},
						Device: &models.SystemAuthenticationDevice{RequireVerifiedKey: required},
					},
				}))

				system, err := st.SystemGet(ctx)
				require.NoError(t, err)
				require.NotNil(t, system.Authentication.Device)
				assert.Equal(t, required, system.Authentication.Device.RequireVerifiedKey)
			})
		}
	})
}

// TestSystemGetDefault tests that SystemGet returns default config when DB is empty
func (s *Suite) TestSystemGetDefault(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("returns default when no system exists", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		system, err := st.SystemGet(ctx)
		require.NoError(t, err)
		require.NotNil(t, system)

		assert.False(t, system.Setup)
		assert.NotNil(t, system.Authentication)
		assert.NotNil(t, system.Authentication.Local)
		assert.True(t, system.Authentication.Local.Enabled)
		require.NotNil(t, system.Authentication.Device)
		assert.True(t, system.Authentication.Device.RequireVerifiedKey, "a new instance requires verified device keys")
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
			},
		}

		err := st.SystemSet(ctx, system)
		require.NoError(t, err)

		created, err := st.SystemGet(ctx)
		require.NoError(t, err)
		assert.True(t, created.Setup)
	})

	t.Run("updates existing system", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		initialSystem := &models.System{
			Setup: false,
			Authentication: &models.SystemAuthentication{
				Local: &models.SystemAuthenticationLocal{
					Enabled: true,
				},
			},
		}

		err := st.SystemSet(ctx, initialSystem)
		require.NoError(t, err)

		updatedSystem := &models.System{
			Setup: true,
			Authentication: &models.SystemAuthentication{
				Local: &models.SystemAuthenticationLocal{
					Enabled: false,
				},
			},
		}

		err = st.SystemSet(ctx, updatedSystem)
		require.NoError(t, err)

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
			},
		}

		err := st.SystemSet(ctx, system)
		require.NoError(t, err)

		created, err := st.SystemGet(ctx)
		require.NoError(t, err)
		assert.NotNil(t, created)
	})
}
