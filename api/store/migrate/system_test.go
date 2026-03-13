package migrate

import (
	"testing"

	"github.com/google/uuid" //nolint:depguard
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertSystem(t *testing.T) {
	t.Run("minimal system", func(t *testing.T) {
		doc := mongoSystem{
			Setup: true,
		}

		result := convertSystem(doc)

		_, err := uuid.Parse(result.ID)
		require.NoError(t, err)
		assert.True(t, result.Setup)
		assert.False(t, result.Authentication.Local.Enabled)
	})

	t.Run("nil authentication", func(t *testing.T) {
		doc := mongoSystem{
			Setup:          false,
			Authentication: nil,
		}

		result := convertSystem(doc)

		assert.False(t, result.Setup)
		assert.False(t, result.Authentication.Local.Enabled)
	})

	t.Run("local auth only", func(t *testing.T) {
		doc := mongoSystem{
			Setup: true,
			Authentication: &mongoSystemAuth{
				Local: &mongoSystemAuthLocal{Enabled: true},
			},
		}

		result := convertSystem(doc)

		assert.True(t, result.Authentication.Local.Enabled)
	})

	t.Run("unique IDs", func(t *testing.T) {
		doc := mongoSystem{Setup: true}
		a := convertSystem(doc)
		b := convertSystem(doc)
		assert.NotEqual(t, a.ID, b.ID)
	})
}
