package services

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/cache"
	cachemock "github.com/shellhub-io/shellhub/pkg/cache/mocks"
	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/shellhub-io/shellhub/pkg/models"
	storemock "github.com/shellhub-io/shellhub/server/api/store/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBuildInstallOverrides(t *testing.T) {
	cases := []struct {
		description string
		req         *requests.SystemInstallScript
		contains    []string
		excludes    []string
	}{
		{
			description: "injects SERVER_ADDRESS from host and forwarded proto",
			req:         &requests.SystemInstallScript{Host: "cloud.example.com", Scheme: "https"},
			contains:    []string{"\nSERVER_ADDRESS=\"${SERVER_ADDRESS:-https://cloud.example.com}\"\n"},
		},
		{
			description: "defaults the scheme to https when not forwarded",
			req:         &requests.SystemInstallScript{Host: "cloud.example.com"},
			contains:    []string{"SERVER_ADDRESS=\"${SERVER_ADDRESS:-https://cloud.example.com}\""},
		},
		{
			description: "appends a non-standard forwarded port",
			req:         &requests.SystemInstallScript{Host: "localhost", ForwardedPort: "8443", Scheme: "https"},
			contains:    []string{"SERVER_ADDRESS=\"${SERVER_ADDRESS:-https://localhost:8443}\""},
		},
		{
			description: "omits the default port for the scheme",
			req:         &requests.SystemInstallScript{Host: "cloud.example.com", ForwardedPort: "443", Scheme: "https"},
			contains:    []string{"SERVER_ADDRESS=\"${SERVER_ADDRESS:-https://cloud.example.com}\""},
			excludes:    []string{":443"},
		},
		{
			description: "keeps the http scheme when forwarded",
			req:         &requests.SystemInstallScript{Host: "cloud.example.com", Scheme: "http"},
			contains:    []string{"SERVER_ADDRESS=\"${SERVER_ADDRESS:-http://cloud.example.com}\""},
		},
		{
			description: "omits the default http port 80",
			req:         &requests.SystemInstallScript{Host: "cloud.example.com", ForwardedPort: "80", Scheme: "http"},
			contains:    []string{"SERVER_ADDRESS=\"${SERVER_ADDRESS:-http://cloud.example.com}\""},
			excludes:    []string{":80"},
		},
		{
			description: "appends a non-standard http port",
			req:         &requests.SystemInstallScript{Host: "localhost", ForwardedPort: "8080", Scheme: "http"},
			contains:    []string{"SERVER_ADDRESS=\"${SERVER_ADDRESS:-http://localhost:8080}\""},
		},
		{
			description: "keeps a port carried inline on the host (direct access, no forwarded port)",
			req:         &requests.SystemInstallScript{Host: "localhost:8080"},
			contains:    []string{"SERVER_ADDRESS=\"${SERVER_ADDRESS:-https://localhost:8080}\""},
		},
		{
			description: "prefers the forwarded port over an inline host port",
			req:         &requests.SystemInstallScript{Host: "localhost:8080", ForwardedPort: "9000", Scheme: "https"},
			contains:    []string{"SERVER_ADDRESS=\"${SERVER_ADDRESS:-https://localhost:9000}\""},
		},
		{
			description: "appends port 443 when scheme is http (non-default for http)",
			req:         &requests.SystemInstallScript{Host: "localhost", ForwardedPort: "443", Scheme: "http"},
			contains:    []string{"SERVER_ADDRESS=\"${SERVER_ADDRESS:-http://localhost:443}\""},
		},
		{
			description: "injects the optional query overrides when present",
			req: &requests.SystemInstallScript{
				Host:              "cloud.example.com",
				Scheme:            "https",
				TenantID:          "00000000-0000-4000-0000-000000000000",
				PreferredHostname: "my-host",
			},
			contains: []string{
				"TENANT_ID=\"${TENANT_ID:-00000000-0000-4000-0000-000000000000}\"",
				"PREFERRED_HOSTNAME=\"${PREFERRED_HOSTNAME:-my-host}\"",
			},
		},
		{
			description: "omits optional overrides that are absent",
			req:         &requests.SystemInstallScript{Host: "cloud.example.com", Scheme: "https"},
			excludes:    []string{"TENANT_ID=", "PREFERRED_HOSTNAME=", "PREFERRED_IDENTITY="},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			out := buildInstallOverrides(tc.req)

			assert.Equal(tt, "\n", out[:1])

			for _, want := range tc.contains {
				assert.Contains(tt, out, want)
			}

			for _, unwanted := range tc.excludes {
				assert.NotContains(tt, out, unwanted)
			}
		})
	}
}

func TestSystemGet(t *testing.T) {
	setupSystem := &models.System{
		Setup: true,
		Authentication: &models.SystemAuthentication{
			Local: &models.SystemAuthenticationLocal{Enabled: true},
		},
	}

	t.Run("serves a cached row without touching the store", func(t *testing.T) {
		storeMock := storemock.NewMockStore(t)
		cacheMock := cachemock.NewMockCache(t)

		cacheMock.
			On("Get", mock.Anything, cache.SystemKey, mock.AnythingOfType("*models.System")).
			Run(func(args mock.Arguments) {
				arg, ok := args.Get(2).(*models.System)
				require.True(t, ok)
				*arg = *setupSystem
			}).
			Return(nil).
			Once()

		system, err := NewService(storeMock, privateKey, publicKey, cacheMock).systemGet(context.TODO())
		require.NoError(t, err)
		assert.Equal(t, setupSystem, system)
	})

	t.Run("reads through on a miss and caches the result", func(t *testing.T) {
		storeMock := storemock.NewMockStore(t)
		cacheMock := cachemock.NewMockCache(t)

		cacheMock.On("Get", mock.Anything, cache.SystemKey, mock.AnythingOfType("*models.System")).Return(nil).Once()
		storeMock.On("SystemGet", mock.Anything).Return(setupSystem, nil).Once()
		cacheMock.On("Set", mock.Anything, cache.SystemKey, setupSystem, systemCacheTTL).Return(nil).Once()

		system, err := NewService(storeMock, privateKey, publicKey, cacheMock).systemGet(context.TODO())
		require.NoError(t, err)
		assert.Equal(t, setupSystem, system)
	})

	t.Run("does not cache before setup completes", func(t *testing.T) {
		pending := &models.System{Setup: false, Authentication: setupSystem.Authentication}

		storeMock := storemock.NewMockStore(t)
		cacheMock := cachemock.NewMockCache(t)

		cacheMock.On("Get", mock.Anything, cache.SystemKey, mock.AnythingOfType("*models.System")).Return(nil).Once()
		storeMock.On("SystemGet", mock.Anything).Return(pending, nil).Once()

		system, err := NewService(storeMock, privateKey, publicKey, cacheMock).systemGet(context.TODO())
		require.NoError(t, err)
		assert.False(t, system.Setup)
	})

	t.Run("ignores a cached row with no authentication", func(t *testing.T) {
		storeMock := storemock.NewMockStore(t)
		cacheMock := cachemock.NewMockCache(t)

		cacheMock.
			On("Get", mock.Anything, cache.SystemKey, mock.AnythingOfType("*models.System")).
			Run(func(args mock.Arguments) {
				arg, ok := args.Get(2).(*models.System)
				require.True(t, ok)
				*arg = models.System{Setup: true}
			}).
			Return(nil).
			Once()
		storeMock.On("SystemGet", mock.Anything).Return(setupSystem, nil).Once()
		cacheMock.On("Set", mock.Anything, cache.SystemKey, setupSystem, systemCacheTTL).Return(nil).Once()

		system, err := NewService(storeMock, privateKey, publicKey, cacheMock).systemGet(context.TODO())
		require.NoError(t, err)
		require.NotNil(t, system.Authentication)
		assert.True(t, system.Authentication.Local.Enabled)
	})

	t.Run("serves the row when the cache is unreachable", func(t *testing.T) {
		storeMock := storemock.NewMockStore(t)
		cacheMock := cachemock.NewMockCache(t)

		cacheMock.
			On("Get", mock.Anything, cache.SystemKey, mock.AnythingOfType("*models.System")).
			Return(errors.New("connection refused", "", 0)).
			Once()
		storeMock.On("SystemGet", mock.Anything).Return(setupSystem, nil).Once()
		cacheMock.
			On("Set", mock.Anything, cache.SystemKey, setupSystem, systemCacheTTL).
			Return(errors.New("connection refused", "", 0)).
			Once()

		system, err := NewService(storeMock, privateKey, publicKey, cacheMock).systemGet(context.TODO())
		require.NoError(t, err)
		assert.Equal(t, setupSystem, system)
	})
}
