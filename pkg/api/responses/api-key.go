package responses

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type CreateAPIKey struct {
	ID        string          `json:"id" bson:"_id"`
	Name      string          `json:"name" bson:"name"`
	UserID    string          `json:"user_id" bson:"user_id"`
	TenantID  string          `json:"tenant_id" bson:"tenant_id"`
	Role      authorizer.Role `json:"role" bson:"role" validate:"required,oneof=administrator operator observer"`
	CreatedAt time.Time       `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time       `json:"updated_at" bson:"updated_at"`
	ExpiresIn int64           `json:"expires_in" bson:"expires_in"`
}

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
