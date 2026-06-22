package services

import (
	"context"
	"testing"

	storemock "github.com/shellhub-io/shellhub/api/store/mocks"
	cachemock "github.com/shellhub-io/shellhub/pkg/cache/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockLicenseEvaluator is a minimal in-test stub of LicenseEvaluator.
type mockLicenseEvaluator struct{}

func (m *mockLicenseEvaluator) CanAcceptDevice(_ context.Context) (bool, error) {
	return true, nil
}

func TestWithLicenseEvaluator(t *testing.T) {
	t.Run("licenseEvaluator is nil without the option", func(t *testing.T) {
		store := storemock.NewMockStore(t)
		cache := cachemock.NewMockCache(t)

		svc := NewService(store, privateKey, publicKey, cache, clientMock)
		require.NotNil(t, svc)

		assert.Nil(t, svc.licenseEvaluator)
	})

	t.Run("WithLicenseEvaluator injects a non-nil evaluator", func(t *testing.T) {
		store := storemock.NewMockStore(t)
		cache := cachemock.NewMockCache(t)

		evaluator := &mockLicenseEvaluator{}
		svc := NewService(store, privateKey, publicKey, cache, clientMock,
			WithLicenseEvaluator(evaluator),
		)
		require.NotNil(t, svc)

		assert.NotNil(t, svc.licenseEvaluator)
		assert.Equal(t, evaluator, svc.licenseEvaluator)
	})
}
