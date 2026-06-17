package services

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLicenseEvaluator(t *testing.T) {
	t.Run("factory is nil before registration", func(t *testing.T) {
		// Reset the package-level var before the test.
		licenseEvaluatorFactory = nil

		got := LicenseEvaluatorFactory()
		assert.Nil(t, got)
	})

	t.Run("factory is non-nil after registration", func(t *testing.T) {
		licenseEvaluatorFactory = nil

		stub := func(_ context.Context, _ store.Store, _ cache.Cache) (LicenseEvaluator, error) {
			return nil, nil
		}

		RegisterLicenseEvaluator(stub)

		got := LicenseEvaluatorFactory()
		require.NotNil(t, got)
	})

	t.Run("LicenseEvaluatorFactory returns the registered function", func(t *testing.T) {
		licenseEvaluatorFactory = nil

		var called bool

		stub := func(_ context.Context, _ store.Store, _ cache.Cache) (LicenseEvaluator, error) {
			called = true

			return nil, nil
		}

		RegisterLicenseEvaluator(stub)

		got := LicenseEvaluatorFactory()
		require.NotNil(t, got)

		// Invoke the returned factory to confirm it is the same function.
		//nolint:errcheck
		got(context.Background(), nil, nil)
		assert.True(t, called)
	})
}
