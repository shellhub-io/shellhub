package services

import (
	"context"
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOnMembershipInvited(t *testing.T) {
	// Save and restore global state.
	original := membershipInvitedHooks
	t.Cleanup(func() { membershipInvitedHooks = original })

	t.Run("panics on nil hook", func(t *testing.T) {
		assert.Panics(t, func() { OnMembershipInvited(nil) })
	})

	t.Run("registers a hook", func(t *testing.T) {
		membershipInvitedHooks = nil

		called := false
		OnMembershipInvited(func(_ context.Context, _ *models.MembershipInvitation, _, _ string) error {
			called = true

			return nil
		})

		require.Len(t, membershipInvitedHooks, 1)

		err := membershipInvitedHooks[0](context.Background(), nil, "", "")
		assert.NoError(t, err)
		assert.True(t, called)
	})
}

func TestFireMembershipInvited(t *testing.T) {
	original := membershipInvitedHooks
	t.Cleanup(func() { membershipInvitedHooks = original })

	inv := &models.MembershipInvitation{TenantID: "tenant", Sig: "ABCDEF123456"}

	t.Run("no hooks registered", func(t *testing.T) {
		membershipInvitedHooks = nil

		err := fireMembershipInvited(context.Background(), inv, "", "")
		assert.NoError(t, err)
	})

	t.Run("single hook succeeds", func(t *testing.T) {
		membershipInvitedHooks = nil

		called := false
		OnMembershipInvited(func(_ context.Context, got *models.MembershipInvitation, _, _ string) error {
			called = true
			assert.Equal(t, inv, got)

			return nil
		})

		err := fireMembershipInvited(context.Background(), inv, "", "")
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("first hook error aborts execution", func(t *testing.T) {
		membershipInvitedHooks = nil

		errHook := errors.New("hook failed")
		OnMembershipInvited(func(_ context.Context, _ *models.MembershipInvitation, _, _ string) error {
			return errHook
		})

		secondCalled := false
		OnMembershipInvited(func(_ context.Context, _ *models.MembershipInvitation, _, _ string) error {
			secondCalled = true

			return nil
		})

		err := fireMembershipInvited(context.Background(), inv, "", "")
		assert.ErrorIs(t, err, errHook)
		assert.False(t, secondCalled)
	})

	t.Run("multiple hooks run in order", func(t *testing.T) {
		membershipInvitedHooks = nil

		var order []int
		OnMembershipInvited(func(_ context.Context, _ *models.MembershipInvitation, _, _ string) error {
			order = append(order, 1)

			return nil
		})
		OnMembershipInvited(func(_ context.Context, _ *models.MembershipInvitation, _, _ string) error {
			order = append(order, 2)

			return nil
		})

		err := fireMembershipInvited(context.Background(), inv, "", "")
		assert.NoError(t, err)
		assert.Equal(t, []int{1, 2}, order)
	})
}
