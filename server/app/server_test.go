package app

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/cache"
	cachemock "github.com/shellhub-io/shellhub/pkg/cache/mocks"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/services"
	"github.com/shellhub-io/shellhub/server/api/store"
	storemock "github.com/shellhub-io/shellhub/server/api/store/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestLicenseEvaluatorWiring verifies the license-evaluator factory wiring in server setup.
// It exercises licenseEvaluatorOption, the helper extracted from Setup() so the logic is
// testable without a live database or Redis connection.
func TestLicenseEvaluatorWiring(t *testing.T) {
	original := services.LicenseEvaluatorFactory()
	t.Cleanup(func() { services.RegisterLicenseEvaluator(original) })

	t.Run("no factory registered: returns empty option slice", func(t *testing.T) {
		services.RegisterLicenseEvaluator(nil)

		s := &Server{env: &Env{}}
		mockStore := storemock.NewMockStore(t)
		mockCache := cachemock.NewMockCache(t)

		opts, err := s.licenseEvaluatorOption(context.Background(), mockStore, mockCache)
		require.NoError(t, err)
		assert.Empty(t, opts)
	})

	t.Run("factory returns non-nil evaluator: option is appended", func(t *testing.T) {
		stub := &stubLicenseEvaluator{}

		services.RegisterLicenseEvaluator(func(_ context.Context, _ store.Store, _ cache.Cache) (services.LicenseEvaluator, error) {
			return stub, nil
		})

		s := &Server{env: &Env{}}
		mockStore := storemock.NewMockStore(t)
		mockCache := cachemock.NewMockCache(t)

		opts, err := s.licenseEvaluatorOption(context.Background(), mockStore, mockCache)
		require.NoError(t, err)
		assert.Len(t, opts, 1)
	})

	t.Run("factory returns untyped nil evaluator: option is NOT appended", func(t *testing.T) {
		services.RegisterLicenseEvaluator(func(_ context.Context, _ store.Store, _ cache.Cache) (services.LicenseEvaluator, error) {
			return nil, nil
		})

		s := &Server{env: &Env{}}
		mockStore := storemock.NewMockStore(t)
		mockCache := cachemock.NewMockCache(t)

		opts, err := s.licenseEvaluatorOption(context.Background(), mockStore, mockCache)
		require.NoError(t, err)
		assert.Empty(t, opts)
	})

	t.Run("factory returns error: error is propagated", func(t *testing.T) {
		factoryErr := errors.New("license init failed")

		services.RegisterLicenseEvaluator(func(_ context.Context, _ store.Store, _ cache.Cache) (services.LicenseEvaluator, error) {
			return nil, factoryErr
		})

		s := &Server{env: &Env{}}
		mockStore := storemock.NewMockStore(t)
		mockCache := cachemock.NewMockCache(t)

		opts, err := s.licenseEvaluatorOption(context.Background(), mockStore, mockCache)
		require.Error(t, err)
		assert.Empty(t, opts)
	})
}

type stubLicenseEvaluator struct{}

func (s *stubLicenseEvaluator) CanAcceptDevice(_ context.Context) (bool, error) {
	return true, nil
}

func (s *stubLicenseEvaluator) CanConnectDevice(_ context.Context) (bool, error) {
	return true, nil
}

func TestInstanceIssuer(t *testing.T) {
	cases := []struct {
		description string
		env         *Env
		expected    string
	}{
		{
			description: "derives http from the domain when the gateway serves plain HTTP",
			env:         &Env{Domain: "localhost"},
			expected:    "http://localhost",
		},
		{
			description: "derives https from the domain when the gateway terminates TLS",
			env:         &Env{Domain: "shellhub.example.com", AutoSSL: true},
			expected:    "https://shellhub.example.com",
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			s := &Server{env: tc.env}

			assert.Equal(t, tc.expected, s.instanceIssuer())
		})
	}
}

func TestVerifiedDeviceKeyOption(t *testing.T) {
	system := func(required bool) *models.System {
		return &models.System{Authentication: &models.SystemAuthentication{
			Device: &models.SystemAuthenticationDevice{RequireVerifiedKey: required},
		}}
	}

	authUnsigned := func(t *testing.T, st *storemock.MockStore, opts []services.Option) error {
		t.Helper()

		key, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		service := services.NewService(st, key, &key.PublicKey, cachemock.NewMockCache(t), opts...)

		_, err = service.AuthDevice(context.Background(), requests.DeviceAuth{TenantID: "00000000-0000-4000-0000-000000000000", PublicKey: "key"})

		return err
	}

	assertRefusesUnsigned := func(t *testing.T, opts []services.Option) {
		t.Helper()

		assert.ErrorIs(t, authUnsigned(t, storemock.NewMockStore(t), opts), services.ErrDeviceKeyVerificationRefused)
	}

	assertAcceptsUnsigned := func(t *testing.T, opts []services.Option) {
		t.Helper()

		st := storemock.NewMockStore(t)
		st.On("NamespaceResolve", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("past the verification")).Once()

		assert.NotErrorIs(t, authUnsigned(t, st, opts), services.ErrDeviceKeyVerificationRefused)
	}

	t.Run("the environment requires it without reading the instance", func(t *testing.T) {
		s := &Server{env: &Env{RequireVerifiedDeviceKey: true}}

		opts, err := s.verifiedDeviceKeyOption(context.Background(), storemock.NewMockStore(t))
		require.NoError(t, err)
		assertRefusesUnsigned(t, opts)
	})

	t.Run("an instance set up with the requirement requires it", func(t *testing.T) {
		mockStore := storemock.NewMockStore(t)
		mockStore.On("SystemGet", context.Background()).Return(system(true), nil).Once()

		opts, err := (&Server{env: &Env{}}).verifiedDeviceKeyOption(context.Background(), mockStore)
		require.NoError(t, err)
		assertRefusesUnsigned(t, opts)
	})

	t.Run("an upgraded instance does not, until the environment asks", func(t *testing.T) {
		mockStore := storemock.NewMockStore(t)
		mockStore.On("SystemGet", context.Background()).Return(system(false), nil).Once()

		opts, err := (&Server{env: &Env{}}).verifiedDeviceKeyOption(context.Background(), mockStore)
		require.NoError(t, err)
		assertAcceptsUnsigned(t, opts)
	})

	t.Run("an instance whose settings leave device authentication out requires it, as the store does", func(t *testing.T) {
		mockStore := storemock.NewMockStore(t)
		mockStore.On("SystemGet", context.Background()).Return(&models.System{}, nil).Once()

		opts, err := (&Server{env: &Env{}}).verifiedDeviceKeyOption(context.Background(), mockStore)
		require.NoError(t, err)
		assertRefusesUnsigned(t, opts)
	})

	t.Run("reports a store failure", func(t *testing.T) {
		down := errors.New("store down")

		mockStore := storemock.NewMockStore(t)
		mockStore.On("SystemGet", context.Background()).Return(nil, down).Once()

		_, err := (&Server{env: &Env{}}).verifiedDeviceKeyOption(context.Background(), mockStore)
		require.ErrorIs(t, err, down)
		assert.ErrorContains(t, err, "init device key verification")
	})
}

func TestTunnelPingInterval(t *testing.T) {
	parse := func(values map[string]string) (*sshEnv, error) {
		return envs.ParseWithPrefix[sshEnv](envs.Map(values), "SSH_")
	}

	t.Run("is zero when unset, leaving each transport its own interval", func(t *testing.T) {
		for _, values := range []map[string]string{{}, {"SHELLHUB_TUNNEL_PING_INTERVAL": ""}} {
			env, err := parse(values)
			require.NoError(t, err)
			assert.Zero(t, env.TunnelPingInterval)
		}
	})

	for value, interval := range map[string]time.Duration{"15s": 15 * time.Second, "59s": 59 * time.Second} {
		t.Run("honours "+value, func(t *testing.T) {
			env, err := parse(map[string]string{"SHELLHUB_TUNNEL_PING_INTERVAL": value})
			require.NoError(t, err)
			assert.Equal(t, interval, time.Duration(env.TunnelPingInterval))
		})
	}

	for _, value := range []string{"0s", "-1s", "soon", "1m", "3m"} {
		t.Run("refuses "+value+" naming the variable", func(t *testing.T) {
			_, err := parse(map[string]string{"SHELLHUB_TUNNEL_PING_INTERVAL": value})
			require.ErrorIs(t, err, envs.ErrParseWithPrefix)
			assert.ErrorContains(t, err, "SHELLHUB_TUNNEL_PING_INTERVAL")
		})
	}
}
