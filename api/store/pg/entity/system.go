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
}

type SystemAuthenticationLocal struct {
	Enabled bool `bun:"enabled"`
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
		},
	}
}
