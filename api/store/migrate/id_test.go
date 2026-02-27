package migrate

import (
	"testing"

	"github.com/google/uuid" //nolint:depguard
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestObjectIDToUUID(t *testing.T) {
	t.Run("deterministic", func(t *testing.T) {
		hex := "507f1f77bcf86cd799439011"
		first := ObjectIDToUUID(hex)
		second := ObjectIDToUUID(hex)
		assert.Equal(t, first, second)
	})

	t.Run("different inputs produce different outputs", func(t *testing.T) {
		a := ObjectIDToUUID("507f1f77bcf86cd799439011")
		b := ObjectIDToUUID("507f1f77bcf86cd799439012")
		assert.NotEqual(t, a, b)
	})

	t.Run("valid UUID format", func(t *testing.T) {
		result := ObjectIDToUUID("507f1f77bcf86cd799439011")
		_, err := uuid.Parse(result)
		require.NoError(t, err)
	})

	t.Run("UUID v5 SHA1", func(t *testing.T) {
		result, err := uuid.Parse(ObjectIDToUUID("507f1f77bcf86cd799439011"))
		require.NoError(t, err)
		assert.Equal(t, uuid.Version(5), result.Version())
	})

	t.Run("known value", func(t *testing.T) {
		hex := "507f1f77bcf86cd799439011"
		expected := uuid.NewSHA1(migrationNamespace, []byte(hex)).String()
		assert.Equal(t, expected, ObjectIDToUUID(hex))
	})

	t.Run("empty string", func(t *testing.T) {
		result := ObjectIDToUUID("")
		_, err := uuid.Parse(result)
		require.NoError(t, err)
	})
}
