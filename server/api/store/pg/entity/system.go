package entity

import (
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

// System is the single row of systems, holding instance-wide state: whether setup has run, and
// which namespace the instance itself is bound to.
type System struct {
	bun.BaseModel `bun:"table:systems"`

	ID               string               `bun:"id,pk,type:uuid"`
	Setup            bool                 `bun:"setup"`
	InstanceTenantID string               `bun:"instance_tenant_id,nullzero,type:uuid"`
	Authentication   SystemAuthentication `bun:"embed:authentication_"`
}

// SystemAuthentication is embedded into the systems row, so the login methods are read with the
// rest of the instance state.
type SystemAuthentication struct {
	Local  SystemAuthenticationLocal  `bun:"embed:local_"`
	Device SystemAuthenticationDevice `bun:"embed:device_"`
}

// SystemAuthenticationLocal is whether password login is offered.
type SystemAuthenticationLocal struct {
	Enabled bool `bun:"enabled"`
}

// SystemAuthenticationDevice is whether device authentication must sign a server challenge.
type SystemAuthenticationDevice struct {
	RequireVerifiedKey bool `bun:"require_verified_key"`
}

// SystemFromModel projects instance state into its row form, tolerating a nil model. A model
// without device authentication settings, a nil one included, requires verified device keys, as a
// new instance does: a caller that writes the row without reading it first is creating it.
func SystemFromModel(model *models.System) *System {
	if model == nil {
		model = &models.System{}
	}

	entity := &System{
		Setup:            model.Setup,
		InstanceTenantID: model.InstanceTenantID,
	}

	entity.Authentication.Device.RequireVerifiedKey = model.Authentication.RequiresVerifiedDeviceKey()

	if model.Authentication != nil {
		if model.Authentication.Local != nil {
			entity.Authentication.Local.Enabled = model.Authentication.Local.Enabled
		}
	}

	return entity
}

// SystemToModel rebuilds instance state from its row.
func SystemToModel(entity *System) *models.System {
	if entity == nil {
		return &models.System{}
	}

	return &models.System{
		Setup:            entity.Setup,
		InstanceTenantID: entity.InstanceTenantID,
		Authentication: &models.SystemAuthentication{
			Local: &models.SystemAuthenticationLocal{
				Enabled: entity.Authentication.Local.Enabled,
			},
			Device: &models.SystemAuthenticationDevice{
				RequireVerifiedKey: entity.Authentication.Device.RequireVerifiedKey,
			},
		},
	}
}
