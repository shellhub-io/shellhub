package entity

import (
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublicKeyFromModel(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name  string
		model *models.PublicKey
		check func(t *testing.T, result *PublicKey)
	}{
		{
			name: "full fields with Tags",
			model: &models.PublicKey{
				TenantID:    "tenant-id-1",
				Fingerprint: "SHA256:abc123",
				Data:        []byte("ssh-rsa AAAA..."),
				CreatedAt:   now,
				PublicKeyFields: models.PublicKeyFields{
					Name: "my-key",
					Filter: models.PublicKeyFilter{
						Hostname: "*.example.com",
						Taggable: models.Taggable{
							Tags: []models.Tag{
								{ID: "tag-1", Name: "prod", TenantID: "t1"},
								{ID: "tag-2", Name: "staging", TenantID: "t1"},
							},
						},
					},
				},
			},
			check: func(t *testing.T, result *PublicKey) {
				assert.Equal(t, "tenant-id-1", result.NamespaceID)
				assert.Equal(t, "SHA256:abc123", result.Fingerprint)
				assert.Equal(t, []byte("ssh-rsa AAAA..."), result.Data)
				assert.Equal(t, "my-key", result.Name)
				assert.Equal(t, "*.example.com", result.FilterHostname)
				assert.Equal(t, now, result.CreatedAt)
				require.Len(t, result.Tags, 2)
				assert.Equal(t, "tag-1", result.Tags[0].ID)
				assert.Equal(t, "prod", result.Tags[0].Name)
				assert.True(t, result.UpdatedAt.IsZero())
			},
		},
		{
			name: "Tags from TagIDs",
			model: &models.PublicKey{
				Fingerprint: "SHA256:def456",
				PublicKeyFields: models.PublicKeyFields{
					Name: "id-key",
					Filter: models.PublicKeyFilter{
						Hostname: "host",
						Taggable: models.Taggable{
							TagIDs: []string{"tag-1", "tag-2"},
						},
					},
				},
			},
			check: func(t *testing.T, result *PublicKey) {
				require.Len(t, result.Tags, 2)
				assert.Equal(t, "tag-1", result.Tags[0].ID)
				assert.Equal(t, "", result.Tags[0].Name)
			},
		},
		{
			name: "no tags",
			model: &models.PublicKey{
				Fingerprint: "SHA256:ghi789",
				PublicKeyFields: models.PublicKeyFields{
					Name: "no-tag-key",
					Filter: models.PublicKeyFilter{
						Hostname: "host",
					},
				},
			},
			check: func(t *testing.T, result *PublicKey) {
				assert.Empty(t, result.Tags)
			},
		},
		{
			name: "Data as []byte",
			model: &models.PublicKey{
				Fingerprint: "SHA256:jkl012",
				Data:        []byte{0x00, 0x01, 0x02},
				PublicKeyFields: models.PublicKeyFields{
					Name:   "binary-key",
					Filter: models.PublicKeyFilter{},
				},
			},
			check: func(t *testing.T, result *PublicKey) {
				assert.Equal(t, []byte{0x00, 0x01, 0x02}, result.Data)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PublicKeyFromModel(tt.model)
			tt.check(t, result)
		})
	}
}

func TestPublicKeyToModel(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name   string
		entity *PublicKey
		check  func(t *testing.T, result *models.PublicKey)
	}{
		{
			name: "full fields with Tags",
			entity: &PublicKey{
				NamespaceID:    "tenant-id-1",
				Fingerprint:    "SHA256:abc123",
				Data:           []byte("ssh-rsa AAAA..."),
				CreatedAt:      now,
				Name:           "my-key",
				FilterHostname: "*.example.com",
				Tags: []*Tag{
					{ID: "tag-1", NamespaceID: "t1", Name: "prod"},
					{ID: "tag-2", NamespaceID: "t1", Name: "staging"},
				},
			},
			check: func(t *testing.T, result *models.PublicKey) {
				assert.Equal(t, "tenant-id-1", result.TenantID)
				assert.Equal(t, "SHA256:abc123", result.Fingerprint)
				assert.Equal(t, []byte("ssh-rsa AAAA..."), result.Data)
				assert.Equal(t, now, result.CreatedAt)
				assert.Equal(t, "my-key", result.PublicKeyFields.Name)
				assert.Equal(t, "", result.PublicKeyFields.Username)
				assert.Equal(t, "*.example.com", result.Filter.Hostname)
				require.Len(t, result.Filter.Tags, 2)
				assert.Equal(t, "tag-1", result.Filter.Tags[0].ID)
				require.Len(t, result.Filter.TagIDs, 2)
				assert.Equal(t, "tag-1", result.Filter.TagIDs[0])
			},
		},
		{
			name: "no Tags",
			entity: &PublicKey{
				NamespaceID:    "tenant-id-2",
				Fingerprint:    "SHA256:def456",
				Name:           "empty-key",
				FilterHostname: "host",
				Tags:           []*Tag{},
			},
			check: func(t *testing.T, result *models.PublicKey) {
				assert.Empty(t, result.Filter.Tags)
				assert.Nil(t, result.Filter.TagIDs)
				assert.Equal(t, "host", result.Filter.Hostname)
				assert.Equal(t, "empty-key", result.PublicKeyFields.Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PublicKeyToModel(tt.entity)
			tt.check(t, result)
		})
	}
}
