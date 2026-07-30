package services

import (
	"context"
	"testing"

	cachemock "github.com/shellhub-io/shellhub/pkg/cache/mocks"
	storemock "github.com/shellhub-io/shellhub/server/api/store/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// mockLicenseEvaluator stands in for LicenseEvaluator. It lives here rather than in the
// generated mocks package because this package's own tests use it: services/mocks
// imports services, so an internal test importing it would form a cycle.
type mockLicenseEvaluator struct {
	mock.Mock
}

func (m *mockLicenseEvaluator) CanAcceptDevice(ctx context.Context) (bool, error) {
	args := m.Called(ctx)

	return args.Bool(0), args.Error(1)
}

func (m *mockLicenseEvaluator) CanConnectDevice(ctx context.Context) (bool, error) {
	args := m.Called(ctx)

	return args.Bool(0), args.Error(1)
}

func TestWithLicenseEvaluator(t *testing.T) {
	t.Run("licenseEvaluator is nil without the option", func(t *testing.T) {
		store := storemock.NewMockStore(t)
		cache := cachemock.NewMockCache(t)

		svc := NewService(store, privateKey, publicKey, cache)
		require.NotNil(t, svc)

		assert.Nil(t, svc.licenseEvaluator)
	})

	t.Run("WithLicenseEvaluator injects a non-nil evaluator", func(t *testing.T) {
		store := storemock.NewMockStore(t)
		cache := cachemock.NewMockCache(t)

		evaluator := &mockLicenseEvaluator{}
		svc := NewService(store, privateKey, publicKey, cache,
			WithLicenseEvaluator(evaluator),
		)
		require.NotNil(t, svc)

		assert.NotNil(t, svc.licenseEvaluator)
		assert.Equal(t, evaluator, svc.licenseEvaluator)
	})
}
