package entity

import (
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestPrivateKeyFromModel(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		model    *models.PrivateKey
		expected *PrivateKey
	}{
		{
			name: "full fields",
			model: &models.PrivateKey{
				Fingerprint: "SHA256:abc123",
				Data:        []byte("private-key-data"),
				CreatedAt:   now,
			},
			expected: &PrivateKey{
				Fingerprint: "SHA256:abc123",
				Data:        []byte("private-key-data"),
				CreatedAt:   now,
				UpdatedAt:   time.Time{},
			},
		},
		{
			name: "nil Data",
			model: &models.PrivateKey{
				Fingerprint: "SHA256:def456",
				Data:        nil,
				CreatedAt:   now,
			},
			expected: &PrivateKey{
				Fingerprint: "SHA256:def456",
				Data:        nil,
				CreatedAt:   now,
				UpdatedAt:   time.Time{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PrivateKeyFromModel(tt.model)
			assert.Equal(t, tt.expected.Fingerprint, result.Fingerprint)
			assert.Equal(t, tt.expected.Data, result.Data)
			assert.Equal(t, tt.expected.CreatedAt, result.CreatedAt)
			assert.True(t, result.UpdatedAt.IsZero(), "UpdatedAt should be zero")
		})
	}
}

func TestPrivateKeyToModel(t *testing.T) {
	// Use a non-UTC timezone to verify .UTC() conversion
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
		{
			name: "already UTC",
			entity: &PrivateKey{
				Fingerprint: "SHA256:def456",
				Data:        []byte("other-key"),
				CreatedAt:   time.Date(2024, 6, 15, 5, 30, 0, 0, time.UTC),
			},
			expected: &models.PrivateKey{
				Fingerprint: "SHA256:def456",
				Data:        []byte("other-key"),
				CreatedAt:   time.Date(2024, 6, 15, 5, 30, 0, 0, time.UTC),
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
