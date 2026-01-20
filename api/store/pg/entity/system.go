package entity

import (
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

type System struct {
	bun.BaseModel `bun:"table:systems"`

	ID             string               `bun:"id,pk,type:uuid"`
	Setup          bool                 `bun:"setup"`
	Authentication SystemAuthentication `bun:"embed:authentication_"`
}

type SystemAuthentication struct {
	Local SystemAuthenticationLocal `bun:"embed:local_"`
	SAML  SystemAuthenticationSAML  `bun:"embed:saml_"`
}

type SystemAuthenticationLocal struct {
	Enabled bool `bun:"enabled"`
}

type SystemAuthenticationSAML struct {
	Enabled bool          `bun:"enabled"`
	Idp     SystemIdpSAML `bun:"embed:idp_"`
	Sp      SystemSpSAML  `bun:"embed:sp_"`
}

type SystemAuthenticationBinding struct {
	Post      string `bun:"binding_post"`
	Redirect  string `bun:"binding_redirect"`
	Preferred string `bun:"binding_preferred"`
}

type SystemIdpSAML struct {
	EntityID     string                      `bun:"entity_id"`
	Binding      SystemAuthenticationBinding `bun:"embed:"`
	Certificates []string                    `bun:"certificates,array"`
	Mappings     map[string]string           `bun:"mappings,type:jsonb"`
}

type SystemSpSAML struct {
	SignAuthRequests bool   `bun:"sign_auth_requests"`
	Certificate      string `bun:"certificate"`
	PrivateKey       string `bun:"private_key"`
}

func SystemFromModel(model *models.System) *System {
	if model == nil {
		return &System{}
	}

	entity := &System{
		Setup: model.Setup,
	}

	if model.Authentication != nil {
		if model.Authentication.Local != nil {
			entity.Authentication.Local.Enabled = model.Authentication.Local.Enabled
		}

		if model.Authentication.SAML != nil {
			entity.Authentication.SAML.Enabled = model.Authentication.SAML.Enabled

			if model.Authentication.SAML.Idp != nil {
				entity.Authentication.SAML.Idp.EntityID = model.Authentication.SAML.Idp.EntityID
				entity.Authentication.SAML.Idp.Certificates = model.Authentication.SAML.Idp.Certificates
				entity.Authentication.SAML.Idp.Mappings = model.Authentication.SAML.Idp.Mappings

				if model.Authentication.SAML.Idp.Binding != nil {
					entity.Authentication.SAML.Idp.Binding.Post = model.Authentication.SAML.Idp.Binding.Post
					entity.Authentication.SAML.Idp.Binding.Redirect = model.Authentication.SAML.Idp.Binding.Redirect
					entity.Authentication.SAML.Idp.Binding.Preferred = model.Authentication.SAML.Idp.Binding.Preferred
				}
			}

			if model.Authentication.SAML.Sp != nil {
				entity.Authentication.SAML.Sp.SignAuthRequests = model.Authentication.SAML.Sp.SignAuthRequests
				entity.Authentication.SAML.Sp.Certificate = model.Authentication.SAML.Sp.Certificate
				entity.Authentication.SAML.Sp.PrivateKey = model.Authentication.SAML.Sp.PrivateKey
			}
		}
	}

	return entity
}

func SystemToModel(entity *System) *models.System {
	if entity == nil {
		return &models.System{}
	}

	return &models.System{
		Setup: entity.Setup,
		Authentication: &models.SystemAuthentication{
			Local: &models.SystemAuthenticationLocal{
				Enabled: entity.Authentication.Local.Enabled,
			},
			SAML: &models.SystemAuthenticationSAML{
				Enabled: entity.Authentication.SAML.Enabled,
				Idp: &models.SystemIdpSAML{
					EntityID:     entity.Authentication.SAML.Idp.EntityID,
					Certificates: entity.Authentication.SAML.Idp.Certificates,
					Mappings:     entity.Authentication.SAML.Idp.Mappings,
					Binding: &models.SystemAuthenticationBinding{
						Post:      entity.Authentication.SAML.Idp.Binding.Post,
						Redirect:  entity.Authentication.SAML.Idp.Binding.Redirect,
						Preferred: entity.Authentication.SAML.Idp.Binding.Preferred,
					},
				},
				Sp: &models.SystemSpSAML{
					SignAuthRequests: entity.Authentication.SAML.Sp.SignAuthRequests,
					Certificate:      entity.Authentication.SAML.Sp.Certificate,
					PrivateKey:       entity.Authentication.SAML.Sp.PrivateKey,
				},
			},
		},
	}
}
