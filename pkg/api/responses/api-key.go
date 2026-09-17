package responses

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
)

// CreateAPIKey is what the create-API-key route returns. Key is the credential and this is the
// only time it is ever sent: nothing can read it back afterwards. ID is the key's surrogate
// identifier, which every other route returns too and which an access policy and an SSH identity
// point at.
type CreateAPIKey struct {
	ID        string          `json:"id"`
	Key       string          `json:"key"`
	Name      string          `json:"name"`
	CreatedBy string          `json:"created_by"`
	TenantID  string          `json:"tenant_id"`
	Role      authorizer.Role `json:"role" validate:"required,oneof=administrator operator observer"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	ExpiresIn int64           `json:"expires_in"`
}

// CreateAPIKeyFromModel projects the stored key onto the response, which is where the model's
// internal fields are dropped rather than serialized by accident. The plaintext is a parameter
// because it is not a field of the key: it exists only in the call that generated it, and is
// never persisted.
func CreateAPIKeyFromModel(m *models.APIKey, plaintext string) *CreateAPIKey {
	return &CreateAPIKey{
		ID:        m.ID,
		Key:       plaintext,
		Name:      m.Name,
		CreatedBy: m.CreatedBy,
		TenantID:  m.TenantID,
		Role:      m.Role,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		ExpiresIn: m.ExpiresIn,
	}
}
