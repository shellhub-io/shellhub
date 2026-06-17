package app

import (
	"context"
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/api/store"
	storemock "github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/cache"
	cachemock "github.com/shellhub-io/shellhub/pkg/cache/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLicenseEvaluatorWiring verifies the license-evaluator factory wiring in server setup.
// It exercises licenseEvaluatorOption, the helper extracted from Setup() so the logic is
// testable without a live database or Redis connection.
func TestLicenseEvaluatorWiring(t *testing.T) {
	t.Cleanup(func() {
		// Reset the package-level factory after all sub-tests so it does not
		// leak into the regression suite.
		services.RegisterLicenseEvaluator(nil)
	})

	t.Run("no factory registered: returns empty option slice", func(t *testing.T) {
		services.RegisterLicenseEvaluator(nil)

		s := &Server{env: &Env{}}
		mockStore := &storemock.Store{}
		mockCache := &cachemock.Cache{}

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
		mockStore := &storemock.Store{}
		mockCache := &cachemock.Cache{}

		opts, err := s.licenseEvaluatorOption(context.Background(), mockStore, mockCache)
		require.NoError(t, err)
		assert.Len(t, opts, 1)
	})

	t.Run("factory returns untyped nil evaluator: option is NOT appended", func(t *testing.T) {
		services.RegisterLicenseEvaluator(func(_ context.Context, _ store.Store, _ cache.Cache) (services.LicenseEvaluator, error) {
			// The real factory returns an untyped nil on its skip path, so a plain
			// nil-check is sufficient — no typed-nil interface can occur.
			return nil, nil
		})

		s := &Server{env: &Env{}}
		mockStore := &storemock.Store{}
		mockCache := &cachemock.Cache{}

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
		mockStore := &storemock.Store{}
		mockCache := &cachemock.Cache{}

		opts, err := s.licenseEvaluatorOption(context.Background(), mockStore, mockCache)
		require.Error(t, err)
		assert.Empty(t, opts)
	})
}

// stubLicenseEvaluator is a minimal in-test implementation of services.LicenseEvaluator.
type stubLicenseEvaluator struct{}

func (s *stubLicenseEvaluator) CanAcceptDevice(_ context.Context) (bool, error) {
	return true, nil
}
