package entity

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

type OAuthClient struct {
	bun.BaseModel `bun:"table:oauth_clients"`

	ID           string    `bun:"id,pk,type:uuid"`
	Name         string    `bun:"name"`
	ClientID     string    `bun:"client_id,type:uuid,unique"`
	ClientSecret string    `bun:"client_secret"`
	NamespaceID  string    `bun:"namespace_id,nullzero,type:uuid"`
	RedirectURIs []string  `bun:"redirect_uris,array"`
	CreatedAt    time.Time `bun:"created_at"`
	UpdatedAt    time.Time `bun:"updated_at"`
}

func OAuthClientFromModel(m *models.OAuthClient) *OAuthClient {
	return &OAuthClient{
		ID:           m.ID,
		Name:         m.Name,
		ClientID:     m.ClientID,
		ClientSecret: m.ClientSecret,
		NamespaceID:  m.TenantID,
		RedirectURIs: m.RedirectURIs,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func OAuthClientToModel(e *OAuthClient) *models.OAuthClient {
	return &models.OAuthClient{
		ID:           e.ID,
		Name:         e.Name,
		ClientID:     e.ClientID,
		ClientSecret: e.ClientSecret,
		TenantID:     e.NamespaceID,
		RedirectURIs: e.RedirectURIs,
		CreatedAt:    e.CreatedAt,
		UpdatedAt:    e.UpdatedAt,
	}
}
