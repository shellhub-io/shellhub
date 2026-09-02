package responses

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
)

// CreateInstanceAPIKey is the response to minting an instance API key. ID carries the plaintext
// key, which is returned here and nowhere else; the server keeps only its digest.
type CreateInstanceAPIKey struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// CreateInstanceAPIKeyFromModel projects a stored instance API key into its creation response.
// The plaintext key is a separate argument because the model carries only its digest: passing it
// explicitly is what stops the digest being returned to the caller by omission.
func CreateInstanceAPIKeyFromModel(m *models.InstanceAPIKey, plaintext string) *CreateInstanceAPIKey {
	return &CreateInstanceAPIKey{
		ID:        plaintext,
		Name:      m.Name,
		CreatedBy: m.CreatedBy,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		ExpiresAt: m.ExpiresAt,
	}
}
