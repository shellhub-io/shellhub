package storetest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *Suite) TestUserInvitationsUpsert(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("succeeds creating new invitation", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create new user invitation
		email := "john.doe@test.com"
		upsertedID, err := st.UserInvitationsUpsert(ctx, email)
		assert.NoError(t, err)
		assert.NotEmpty(t, upsertedID)
	})

	t.Run("succeeds updating existing invitation", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		// Create initial invitation
		email := "jane.doe@test.com"
		firstID, err := st.UserInvitationsUpsert(ctx, email)
		require.NoError(t, err)
		require.NotEmpty(t, firstID)

		// Upsert the same email again (should update, not create)
		secondID, err := st.UserInvitationsUpsert(ctx, email)
		assert.NoError(t, err)
		assert.NotEmpty(t, secondID)
		// For upsert operations, the ID should remain the same
		assert.Equal(t, firstID, secondID)
	})
}
