package entity

import (
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestAPIKeyFromModel(t *testing.T) {
	now := clock.Now()

	tests := []struct {
		name     string
		model    *models.APIKey
		expected *APIKey
	}{
		{
			name: "full fields",
			model: &models.APIKey{
				ID:        "c629572a-b643-4301-90fe-4572b00d007e",
				Digest:    "digest-abc123",
				Name:      "my-api-key",
				TenantID:  "namespace-id-1",
				Role:      authorizer.RoleAdministrator,
				CreatedBy: "user-id-1",
				CreatedAt: now,
				UpdatedAt: now.Add(time.Hour),
				ExpiresIn: 3600,
			},
			expected: &APIKey{
				ID:          "c629572a-b643-4301-90fe-4572b00d007e",
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
				ID:        "9f1f3b2e-1d9a-4a6f-8f2c-2b3f4a5d6e70",
				Digest:    "digest-def456",
				Name:      "read-only-key",
				TenantID:  "namespace-id-2",
				Role:      authorizer.RoleObserver,
				CreatedBy: "user-id-2",
				CreatedAt: now,
				UpdatedAt: now,
				ExpiresIn: 0,
			},
			expected: &APIKey{
				ID:          "9f1f3b2e-1d9a-4a6f-8f2c-2b3f4a5d6e70",
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
			assert.Equal(t, tt.expected, APIKeyFromModel(tt.model))
		})
	}
}

func TestAPIKeyToModel(t *testing.T) {
	now := clock.Now()

	tests := []struct {
		name     string
		entity   *APIKey
		expected *models.APIKey
	}{
		{
			name: "full fields",
			entity: &APIKey{
				ID:          "c629572a-b643-4301-90fe-4572b00d007e",
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
				ID:        "c629572a-b643-4301-90fe-4572b00d007e",
				Digest:    "digest-abc123",
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
				ID:          "9f1f3b2e-1d9a-4a6f-8f2c-2b3f4a5d6e70",
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
				ID:        "9f1f3b2e-1d9a-4a6f-8f2c-2b3f4a5d6e70",
				Digest:    "digest-def456",
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
			assert.Equal(t, tt.expected, APIKeyToModel(tt.entity))
		})
	}
}
