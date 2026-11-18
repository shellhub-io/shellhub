package services

import (
	"context"
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHooks_deviceMerge(t *testing.T) {
	hooks := &Hooks{}

	ctx := context.Background()
	oldDev := &models.Device{UID: "old-uid"}
	newDev := &models.Device{UID: "new-uid"}
	tenant := "tenant-id"

	t.Run("no hooks registered", func(t *testing.T) {
		*hooks = Hooks{}
		assert.NoError(t, hooks.deviceMerge(ctx, tenant, oldDev, newDev))
	})

	t.Run("single hook called with correct args", func(t *testing.T) {
		*hooks = Hooks{}

		var called bool
		hooks.DeviceMerge = append(hooks.DeviceMerge, func(gotCtx context.Context, tid string, o, n *models.Device) error {
			called = true
			assert.Equal(t, ctx, gotCtx)
			assert.Equal(t, tenant, tid)
			assert.Equal(t, oldDev, o)
			assert.Equal(t, newDev, n)

			return nil
		})

		require.NoError(t, hooks.deviceMerge(ctx, tenant, oldDev, newDev))
		assert.True(t, called)
	})

	t.Run("error aborts remaining hooks", func(t *testing.T) {
		*hooks = Hooks{}
		hookErr := errors.New("hook failed")

		hooks.DeviceMerge = append(hooks.DeviceMerge, func(context.Context, string, *models.Device, *models.Device) error {
			return hookErr
		})

		var secondCalled bool
		hooks.DeviceMerge = append(hooks.DeviceMerge, func(context.Context, string, *models.Device, *models.Device) error {
			secondCalled = true

			return nil
		})

		require.ErrorIs(t, hooks.deviceMerge(ctx, tenant, oldDev, newDev), hookErr)
		assert.False(t, secondCalled)
	})

	t.Run("multiple hooks run in order", func(t *testing.T) {
		*hooks = Hooks{}

		var order []int
		hooks.DeviceMerge = append(hooks.DeviceMerge, func(context.Context, string, *models.Device, *models.Device) error {
			order = append(order, 1)

			return nil
		})
		hooks.DeviceMerge = append(hooks.DeviceMerge, func(context.Context, string, *models.Device, *models.Device) error {
			order = append(order, 2)

			return nil
		})

		require.NoError(t, hooks.deviceMerge(ctx, tenant, oldDev, newDev))
		assert.Equal(t, []int{1, 2}, order)
	})
}
