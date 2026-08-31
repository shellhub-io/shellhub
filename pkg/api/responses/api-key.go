package responses

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
)

// CreateAPIKey is what the create-API-key route returns. ID is the key itself and this is the only
// time it is ever sent: nothing can read it back afterwards.
type CreateAPIKey struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	UserID    string          `json:"user_id"`
	TenantID  string          `json:"tenant_id"`
	Role      authorizer.Role `json:"role" validate:"required,oneof=administrator operator observer"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	ExpiresIn int64           `json:"expires_in"`
}

// CreateAPIKeyFromModel projects the stored key onto the response, which is where the model's
// internal fields are dropped rather than serialized by accident.
func CreateAPIKeyFromModel(m *models.APIKey) *CreateAPIKey {
	return &CreateAPIKey{
		ID:        m.ID,
		Name:      m.Name,
		UserID:    m.CreatedBy,
		TenantID:  m.TenantID,
		Role:      m.Role,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		ExpiresIn: m.ExpiresIn,
	}
}
