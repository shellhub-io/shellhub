package entity

import (
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestPrivateKeyToModel(t *testing.T) {
	loc := time.FixedZone("UTC+5", 5*60*60)
	createdAt := time.Date(2024, 6, 15, 10, 30, 0, 0, loc)

	tests := []struct {
		name     string
		entity   *PrivateKey
		expected *models.PrivateKey
	}{
		{
			name: "converts CreatedAt to UTC",
			entity: &PrivateKey{
				Fingerprint: "SHA256:abc123",
				Data:        []byte("key-data"),
				CreatedAt:   createdAt,
			},
			expected: &models.PrivateKey{
				Fingerprint: "SHA256:abc123",
				Data:        []byte("key-data"),
				CreatedAt:   createdAt.UTC(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PrivateKeyToModel(tt.entity)
			assert.Equal(t, tt.expected.Fingerprint, result.Fingerprint)
			assert.Equal(t, tt.expected.Data, result.Data)
			assert.Equal(t, tt.expected.CreatedAt, result.CreatedAt)
			assert.Equal(t, time.UTC, result.CreatedAt.Location())
		})
	}
}
