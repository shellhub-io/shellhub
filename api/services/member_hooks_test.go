package services

import (
	"context"
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOnMemberAdd(t *testing.T) {
	// Save and restore global state.
	original := memberAddHooks
	t.Cleanup(func() { memberAddHooks = original })

	t.Run("panics on nil hook", func(t *testing.T) {
		assert.Panics(t, func() { OnMemberAdd(nil) })
	})

	t.Run("registers a hook", func(t *testing.T) {
		memberAddHooks = nil

		called := false
		OnMemberAdd(func(_ context.Context, _ *models.Namespace, _ *requests.NamespaceAddMember) error {
			called = true

			return nil
		})

		require.Len(t, memberAddHooks, 1)

		err := memberAddHooks[0](context.Background(), nil, nil)
		assert.NoError(t, err)
		assert.True(t, called)
	})
}

func TestFireMemberAdd(t *testing.T) {
	original := memberAddHooks
	t.Cleanup(func() { memberAddHooks = original })

	ns := &models.Namespace{TenantID: "tenant"}
	req := &requests.NamespaceAddMember{
		TenantID:    "tenant",
		MemberEmail: "user@test.com",
		MemberRole:  authorizer.RoleObserver,
	}

	t.Run("no hooks registered", func(t *testing.T) {
		memberAddHooks = nil

		err := fireMemberAdd(context.Background(), ns, req)
		assert.NoError(t, err)
	})

	t.Run("single hook succeeds", func(t *testing.T) {
		memberAddHooks = nil

		called := false
		OnMemberAdd(func(_ context.Context, gotNS *models.Namespace, gotReq *requests.NamespaceAddMember) error {
			called = true
			assert.Equal(t, ns, gotNS)
			assert.Equal(t, req, gotReq)

			return nil
		})

		err := fireMemberAdd(context.Background(), ns, req)
		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("first hook error aborts execution", func(t *testing.T) {
		memberAddHooks = nil

		errHook := errors.New("hook failed")
		OnMemberAdd(func(_ context.Context, _ *models.Namespace, _ *requests.NamespaceAddMember) error {
			return errHook
		})

		secondCalled := false
		OnMemberAdd(func(_ context.Context, _ *models.Namespace, _ *requests.NamespaceAddMember) error {
			secondCalled = true

			return nil
		})

		err := fireMemberAdd(context.Background(), ns, req)
		assert.ErrorIs(t, err, errHook)
		assert.False(t, secondCalled)
	})

	t.Run("multiple hooks run in order", func(t *testing.T) {
		memberAddHooks = nil

		var order []int
		OnMemberAdd(func(_ context.Context, _ *models.Namespace, _ *requests.NamespaceAddMember) error {
			order = append(order, 1)

			return nil
		})
		OnMemberAdd(func(_ context.Context, _ *models.Namespace, _ *requests.NamespaceAddMember) error {
			order = append(order, 2)

			return nil
		})

		err := fireMemberAdd(context.Background(), ns, req)
		assert.NoError(t, err)
		assert.Equal(t, []int{1, 2}, order)
	})
}
