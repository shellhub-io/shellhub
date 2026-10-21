package storetest

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPrivateKeyGet exercises PrivateKeyGet against the store under test.
func (s *Suite) TestPrivateKeyGet(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when private key is not found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		privKey, err := st.PrivateKeyGet(ctx, "nonexistent")
		require.ErrorIs(t, err, store.ErrNoDocuments)
		assert.Nil(t, privKey)
	})

	t.Run("succeeds when private key is found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		expectedData := []byte("test-data")
		fingerprint := s.CreatePrivateKey(t, WithPrivateKeyData(expectedData))

		privKey, err := st.PrivateKeyGet(ctx, fingerprint)
		require.NoError(t, err)
		require.NotNil(t, privKey)
		assert.Equal(t, fingerprint, privKey.Fingerprint)
		assert.Equal(t, expectedData, privKey.Data)
	})
}
