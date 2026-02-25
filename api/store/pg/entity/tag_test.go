package entity

import (
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestTagFromModel(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		model    *models.Tag
		expected *Tag
	}{
		{
			name: "full fields",
			model: &models.Tag{
				ID:        "tag-id-1",
				TenantID:  "tenant-id-1",
				Name:      "production",
				CreatedAt: now,
				UpdatedAt: now.Add(time.Hour),
			},
			expected: &Tag{
				ID:          "tag-id-1",
				NamespaceID: "tenant-id-1",
				Name:        "production",
				CreatedAt:   now,
				UpdatedAt:   now.Add(time.Hour),
			},
		},
		{
			name: "zero-value times",
			model: &models.Tag{
				ID:       "tag-id-2",
				TenantID: "tenant-id-2",
				Name:     "staging",
			},
			expected: &Tag{
				ID:          "tag-id-2",
				NamespaceID: "tenant-id-2",
				Name:        "staging",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TagFromModel(tt.model)
			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.NamespaceID, result.NamespaceID)
			assert.Equal(t, tt.expected.Name, result.Name)
			assert.Equal(t, tt.expected.CreatedAt, result.CreatedAt)
			assert.Equal(t, tt.expected.UpdatedAt, result.UpdatedAt)
		})
	}
}

func TestTagToModel(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		entity   *Tag
		expected *models.Tag
	}{
		{
			name: "full fields",
			entity: &Tag{
				ID:          "tag-id-1",
				NamespaceID: "tenant-id-1",
				Name:        "production",
				CreatedAt:   now,
				UpdatedAt:   now.Add(time.Hour),
			},
			expected: &models.Tag{
				ID:        "tag-id-1",
				TenantID:  "tenant-id-1",
				Name:      "production",
				CreatedAt: now,
				UpdatedAt: now.Add(time.Hour),
			},
		},
		{
			name: "zero-value times",
			entity: &Tag{
				ID:          "tag-id-2",
				NamespaceID: "tenant-id-2",
				Name:        "staging",
			},
			expected: &models.Tag{
				ID:       "tag-id-2",
				TenantID: "tenant-id-2",
				Name:     "staging",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TagToModel(tt.entity)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewDeviceTag(t *testing.T) {
	result := NewDeviceTag("tag-123", "device-456")
	assert.Equal(t, "tag-123", result.TagID)
	assert.Equal(t, "device-456", result.DeviceID)
}

func TestNewPublicKeyTag(t *testing.T) {
	result := NewPublicKeyTag("tag-123", "fingerprint-456")
	assert.Equal(t, "tag-123", result.TagID)
	assert.Equal(t, "fingerprint-456", result.PublicKeyFingerprint)
}
