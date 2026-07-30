package services

import (
	"context"
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestFireDeviceMerge(t *testing.T) {
	// Save and restore global hooks so tests don't leak.
	saved := deviceMergeHooks
	t.Cleanup(func() { deviceMergeHooks = saved })

	ctx := context.Background()
	oldDev := &models.Device{UID: "old-uid"}
	newDev := &models.Device{UID: "new-uid"}
	tenant := "tenant-id"

	t.Run("nil hook panics", func(t *testing.T) {
		deviceMergeHooks = nil
		assert.PanicsWithValue(t, "services: OnDeviceMerge called with nil hook", func() {
			OnDeviceMerge(nil)
		})
	})

	t.Run("no hooks registered", func(t *testing.T) {
		deviceMergeHooks = nil
		assert.NoError(t, fireDeviceMerge(ctx, tenant, oldDev, newDev))
	})

	t.Run("single hook called with correct args", func(t *testing.T) {
		deviceMergeHooks = nil

		var called bool
		OnDeviceMerge(func(gotCtx context.Context, tid string, o, n *models.Device) error {
			called = true
			assert.Equal(t, ctx, gotCtx)
			assert.Equal(t, tenant, tid)
			assert.Equal(t, oldDev, o)
			assert.Equal(t, newDev, n)

			return nil
		})

		assert.NoError(t, fireDeviceMerge(ctx, tenant, oldDev, newDev))
		assert.True(t, called)
	})

	t.Run("error aborts remaining hooks", func(t *testing.T) {
		deviceMergeHooks = nil
		hookErr := errors.New("hook failed")

		OnDeviceMerge(func(context.Context, string, *models.Device, *models.Device) error {
			return hookErr
		})

		var secondCalled bool
		OnDeviceMerge(func(context.Context, string, *models.Device, *models.Device) error {
			secondCalled = true

			return nil
		})

		assert.ErrorIs(t, fireDeviceMerge(ctx, tenant, oldDev, newDev), hookErr)
		assert.False(t, secondCalled)
	})

	t.Run("multiple hooks run in order", func(t *testing.T) {
		deviceMergeHooks = nil

		var order []int
		OnDeviceMerge(func(context.Context, string, *models.Device, *models.Device) error {
			order = append(order, 1)

			return nil
		})
		OnDeviceMerge(func(context.Context, string, *models.Device, *models.Device) error {
			order = append(order, 2)

			return nil
		})

		assert.NoError(t, fireDeviceMerge(ctx, tenant, oldDev, newDev))
		assert.Equal(t, []int{1, 2}, order)
	})
}
