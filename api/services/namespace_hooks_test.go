package services

import (
	"context"
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestFireNamespaceDelete(t *testing.T) {
	// Save and restore global hooks so tests don't leak.
	saved := namespaceDeleteHooks
	t.Cleanup(func() { namespaceDeleteHooks = saved })

	ctx := context.Background()
	ns := &models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000", Name: "test"}

	t.Run("nil hook panics", func(t *testing.T) {
		namespaceDeleteHooks = nil
		assert.PanicsWithValue(t, "services: OnNamespaceDelete called with nil hook", func() {
			OnNamespaceDelete(nil)
		})
	})

	t.Run("no hooks registered", func(t *testing.T) {
		namespaceDeleteHooks = nil
		assert.NoError(t, fireNamespaceDelete(ctx, ns))
	})

	t.Run("single hook called with correct args", func(t *testing.T) {
		namespaceDeleteHooks = nil

		var called bool
		OnNamespaceDelete(func(gotCtx context.Context, gotNS *models.Namespace) error {
			called = true
			assert.Equal(t, ctx, gotCtx)
			assert.Equal(t, ns, gotNS)

			return nil
		})

		assert.NoError(t, fireNamespaceDelete(ctx, ns))
		assert.True(t, called)
	})

	t.Run("error aborts remaining hooks", func(t *testing.T) {
		namespaceDeleteHooks = nil
		hookErr := errors.New("hook failed")

		OnNamespaceDelete(func(context.Context, *models.Namespace) error {
			return hookErr
		})

		var secondCalled bool
		OnNamespaceDelete(func(context.Context, *models.Namespace) error {
			secondCalled = true

			return nil
		})

		assert.ErrorIs(t, fireNamespaceDelete(ctx, ns), hookErr)
		assert.False(t, secondCalled)
	})

	t.Run("multiple hooks run in order", func(t *testing.T) {
		namespaceDeleteHooks = nil

		var order []int
		OnNamespaceDelete(func(context.Context, *models.Namespace) error {
			order = append(order, 1)

			return nil
		})
		OnNamespaceDelete(func(context.Context, *models.Namespace) error {
			order = append(order, 2)

			return nil
		})

		assert.NoError(t, fireNamespaceDelete(ctx, ns))
		assert.Equal(t, []int{1, 2}, order)
	})
}
