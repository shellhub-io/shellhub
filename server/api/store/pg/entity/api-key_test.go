package entity

import (
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestAPIKeyFromModel(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		model    *models.APIKey
		expected *APIKey
	}{
		{
			name: "full fields",
			model: &models.APIKey{
				ID:        "digest-abc123",
				Name:      "my-api-key",
				TenantID:  "namespace-id-1",
				Role:      authorizer.RoleAdministrator,
				CreatedBy: "user-id-1",
				CreatedAt: now,
				UpdatedAt: now.Add(time.Hour),
				ExpiresIn: 3600,
			},
			expected: &APIKey{
				KeyDigest:   "digest-abc123",
				Name:        "my-api-key",
				NamespaceID: "namespace-id-1",
				Role:        "administrator",
				UserID:      "user-id-1",
				CreatedAt:   now,
				UpdatedAt:   now.Add(time.Hour),
				ExpiresIn:   3600,
			},
		},
		{
			name: "observer role and zero ExpiresIn",
			model: &models.APIKey{
				ID:        "digest-def456",
				Name:      "read-only-key",
				TenantID:  "namespace-id-2",
				Role:      authorizer.RoleObserver,
				CreatedBy: "user-id-2",
				CreatedAt: now,
				UpdatedAt: now,
				ExpiresIn: 0,
			},
			expected: &APIKey{
				KeyDigest:   "digest-def456",
				Name:        "read-only-key",
				NamespaceID: "namespace-id-2",
				Role:        "observer",
				UserID:      "user-id-2",
				CreatedAt:   now,
				UpdatedAt:   now,
				ExpiresIn:   0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := APIKeyFromModel(tt.model)
			assert.Equal(t, tt.expected.KeyDigest, result.KeyDigest)
			assert.Equal(t, tt.expected.Name, result.Name)
			assert.Equal(t, tt.expected.NamespaceID, result.NamespaceID)
			assert.Equal(t, tt.expected.Role, result.Role)
			assert.Equal(t, tt.expected.UserID, result.UserID)
			assert.Equal(t, tt.expected.CreatedAt, result.CreatedAt)
			assert.Equal(t, tt.expected.UpdatedAt, result.UpdatedAt)
			assert.Equal(t, tt.expected.ExpiresIn, result.ExpiresIn)
		})
	}
}

func TestAPIKeyToModel(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		entity   *APIKey
		expected *models.APIKey
	}{
		{
			name: "full fields",
			entity: &APIKey{
				KeyDigest:   "digest-abc123",
				Name:        "my-api-key",
				NamespaceID: "namespace-id-1",
				Role:        "administrator",
				UserID:      "user-id-1",
				CreatedAt:   now,
				UpdatedAt:   now.Add(time.Hour),
				ExpiresIn:   3600,
			},
			expected: &models.APIKey{
				ID:        "digest-abc123",
				Name:      "my-api-key",
				TenantID:  "namespace-id-1",
				Role:      authorizer.RoleAdministrator,
				CreatedBy: "user-id-1",
				CreatedAt: now,
				UpdatedAt: now.Add(time.Hour),
				ExpiresIn: 3600,
			},
		},
		{
			name: "zero ExpiresIn",
			entity: &APIKey{
				KeyDigest:   "digest-def456",
				Name:        "no-expiry-key",
				NamespaceID: "namespace-id-2",
				Role:        "observer",
				UserID:      "user-id-2",
				CreatedAt:   now,
				UpdatedAt:   now,
				ExpiresIn:   0,
			},
			expected: &models.APIKey{
				ID:        "digest-def456",
				Name:      "no-expiry-key",
				TenantID:  "namespace-id-2",
				Role:      authorizer.RoleObserver,
				CreatedBy: "user-id-2",
				CreatedAt: now,
				UpdatedAt: now,
				ExpiresIn: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := APIKeyToModel(tt.entity)
			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.Name, result.Name)
			assert.Equal(t, tt.expected.TenantID, result.TenantID)
			assert.Equal(t, tt.expected.Role, result.Role)
			assert.Equal(t, tt.expected.CreatedBy, result.CreatedBy)
			assert.Equal(t, tt.expected.CreatedAt, result.CreatedAt)
			assert.Equal(t, tt.expected.UpdatedAt, result.UpdatedAt)
			assert.Equal(t, tt.expected.ExpiresIn, result.ExpiresIn)
		})
	}
}
