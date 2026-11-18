package services

import (
	"context"
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHooks_namespaceDelete(t *testing.T) {
	hooks := &Hooks{}

	ctx := context.Background()
	ns := &models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000", Name: "test"}

	t.Run("no hooks registered", func(t *testing.T) {
		*hooks = Hooks{}
		assert.NoError(t, hooks.namespaceDelete(ctx, ns))
	})

	t.Run("single hook called with correct args", func(t *testing.T) {
		*hooks = Hooks{}

		var called bool
		hooks.NamespaceDelete = append(hooks.NamespaceDelete, func(gotCtx context.Context, gotNS *models.Namespace) error {
			called = true
			assert.Equal(t, ctx, gotCtx)
			assert.Equal(t, ns, gotNS)

			return nil
		})

		require.NoError(t, hooks.namespaceDelete(ctx, ns))
		assert.True(t, called)
	})

	t.Run("error aborts remaining hooks", func(t *testing.T) {
		*hooks = Hooks{}
		hookErr := errors.New("hook failed")

		hooks.NamespaceDelete = append(hooks.NamespaceDelete, func(context.Context, *models.Namespace) error {
			return hookErr
		})

		var secondCalled bool
		hooks.NamespaceDelete = append(hooks.NamespaceDelete, func(context.Context, *models.Namespace) error {
			secondCalled = true

			return nil
		})

		require.ErrorIs(t, hooks.namespaceDelete(ctx, ns), hookErr)
		assert.False(t, secondCalled)
	})

	t.Run("multiple hooks run in order", func(t *testing.T) {
		*hooks = Hooks{}

		var order []int
		hooks.NamespaceDelete = append(hooks.NamespaceDelete, func(context.Context, *models.Namespace) error {
			order = append(order, 1)

			return nil
		})
		hooks.NamespaceDelete = append(hooks.NamespaceDelete, func(context.Context, *models.Namespace) error {
			order = append(order, 2)

			return nil
		})

		require.NoError(t, hooks.namespaceDelete(ctx, ns))
		assert.Equal(t, []int{1, 2}, order)
	})
}
