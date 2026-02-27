package migrate

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestConvertTag(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	updated := now.Add(time.Hour)
	oid := primitive.NewObjectID()

	t.Run("all fields populated", func(t *testing.T) {
		doc := mongoTag{
			ID:        oid,
			TenantID:  "tenant-123",
			Name:      "production",
			CreatedAt: now,
			UpdatedAt: updated,
		}

		result := convertTag(doc)

		assert.Equal(t, ObjectIDToUUID(oid.Hex()), result.ID)
		assert.Equal(t, "tenant-123", result.NamespaceID)
		assert.Equal(t, "production", result.Name)
		assert.Equal(t, now, result.CreatedAt)
		assert.Equal(t, updated, result.UpdatedAt)
	})
}
