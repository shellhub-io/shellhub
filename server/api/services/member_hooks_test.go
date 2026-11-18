package services

import (
	"context"
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHooks_membershipInvited(t *testing.T) {
	hooks := &Hooks{}

	notification := &models.MembershipInvitationNotification{
		Signature:      "ABCDEF123456",
		RecipientEmail: "invitee@test.com",
	}

	t.Run("no hooks registered", func(t *testing.T) {
		*hooks = Hooks{}

		err := hooks.membershipInvited(context.Background(), notification)
		assert.NoError(t, err)
	})

	t.Run("single hook succeeds", func(t *testing.T) {
		*hooks = Hooks{}

		called := false
		hooks.MembershipInvited = append(hooks.MembershipInvited, func(_ context.Context, got *models.MembershipInvitationNotification) error {
			called = true
			assert.Equal(t, notification, got)

			return nil
		})

		err := hooks.membershipInvited(context.Background(), notification)
		require.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("first hook error aborts execution", func(t *testing.T) {
		*hooks = Hooks{}

		errHook := errors.New("hook failed")
		hooks.MembershipInvited = append(hooks.MembershipInvited, func(_ context.Context, _ *models.MembershipInvitationNotification) error {
			return errHook
		})

		secondCalled := false
		hooks.MembershipInvited = append(hooks.MembershipInvited, func(_ context.Context, _ *models.MembershipInvitationNotification) error {
			secondCalled = true

			return nil
		})

		err := hooks.membershipInvited(context.Background(), notification)
		require.ErrorIs(t, err, errHook)
		assert.False(t, secondCalled)
	})

	t.Run("multiple hooks run in order", func(t *testing.T) {
		*hooks = Hooks{}

		var order []int
		hooks.MembershipInvited = append(hooks.MembershipInvited, func(_ context.Context, _ *models.MembershipInvitationNotification) error {
			order = append(order, 1)

			return nil
		})
		hooks.MembershipInvited = append(hooks.MembershipInvited, func(_ context.Context, _ *models.MembershipInvitationNotification) error {
			order = append(order, 2)

			return nil
		})

		err := hooks.membershipInvited(context.Background(), notification)
		require.NoError(t, err)
		assert.Equal(t, []int{1, 2}, order)
	})
}
