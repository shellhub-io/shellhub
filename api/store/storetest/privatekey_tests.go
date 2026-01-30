package storetest

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *Suite) TestPrivateKeyCreate(t *testing.T) {
	t.Run("succeeds when data is valid", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		fingerprint := s.CreatePrivateKey(t)
		assert.NotEmpty(t, fingerprint)
	})
}

func (s *Suite) TestPrivateKeyGet(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	t.Run("fails when private key is not found", func(t *testing.T) {
		require.NoError(t, s.provider.CleanDatabase(t))

		privKey, err := st.PrivateKeyGet(ctx, "nonexistent")
		assert.ErrorIs(t, err, store.ErrNoDocuments)
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
