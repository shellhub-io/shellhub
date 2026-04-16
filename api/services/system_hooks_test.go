package services

import (
	"context"
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/responses"
	"github.com/stretchr/testify/assert"
)

func TestOnGetSystemInfo(t *testing.T) {
	saved := systemInfoHooks
	t.Cleanup(func() { systemInfoHooks = saved })

	t.Run("nil hook panics", func(t *testing.T) {
		systemInfoHooks = nil
		assert.PanicsWithValue(t, "services: OnGetSystemInfo called with nil hook", func() {
			OnGetSystemInfo(nil)
		})
	})

	t.Run("registers hook", func(t *testing.T) {
		systemInfoHooks = nil

		called := false
		OnGetSystemInfo(func(_ context.Context, _ *responses.SystemInfo) error {
			called = true

			return nil
		})

		assert.Len(t, systemInfoHooks, 1)

		err := systemInfoHooks[0](context.Background(), &responses.SystemInfo{})
		assert.NoError(t, err)
		assert.True(t, called)
	})
}

func TestFireGetSystemInfo(t *testing.T) {
	saved := systemInfoHooks
	t.Cleanup(func() { systemInfoHooks = saved })

	ctx := context.Background()
	info := &responses.SystemInfo{
		Authentication: &responses.SystemAuthenticationInfo{Local: true},
	}

	t.Run("no hooks registered", func(t *testing.T) {
		systemInfoHooks = nil
		assert.NoError(t, fireGetSystemInfo(ctx, info))
	})

	t.Run("single hook receives correct args and can mutate info", func(t *testing.T) {
		systemInfoHooks = nil

		OnGetSystemInfo(func(gotCtx context.Context, gotInfo *responses.SystemInfo) error {
			assert.Equal(t, ctx, gotCtx)
			assert.Equal(t, info, gotInfo)
			gotInfo.Authentication.SAML = true

			return nil
		})

		assert.NoError(t, fireGetSystemInfo(ctx, info))
		assert.True(t, info.Authentication.SAML)
	})

	t.Run("error aborts remaining hooks", func(t *testing.T) {
		systemInfoHooks = nil
		hookErr := errors.New("hook failed")

		OnGetSystemInfo(func(context.Context, *responses.SystemInfo) error {
			return hookErr
		})

		secondCalled := false
		OnGetSystemInfo(func(context.Context, *responses.SystemInfo) error {
			secondCalled = true

			return nil
		})

		assert.ErrorIs(t, fireGetSystemInfo(ctx, info), hookErr)
		assert.False(t, secondCalled)
	})

	t.Run("multiple hooks run in order", func(t *testing.T) {
		systemInfoHooks = nil

		var order []int
		OnGetSystemInfo(func(context.Context, *responses.SystemInfo) error {
			order = append(order, 1)

			return nil
		})
		OnGetSystemInfo(func(context.Context, *responses.SystemInfo) error {
			order = append(order, 2)

			return nil
		})

		assert.NoError(t, fireGetSystemInfo(ctx, info))
		assert.Equal(t, []int{1, 2}, order)
	})
}
