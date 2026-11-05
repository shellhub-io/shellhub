package handlers

import (
	"reflect"

	"github.com/shellhub-io/shellhub/pkg/api/identity"
)

func identityField(caller identity.Identity, name string) any {
	switch name {
	case "id":
		return caller.ID
	case "tenant":
		return caller.TenantID
	case "device":
		return caller.DeviceUID
	case "role":
		return caller.Role
	case "admin":
		return caller.Admin
	case "admin_id":
		return caller.AdminID
	default:
		return nil
	}
}

func bindIdentity(caller identity.Identity, target reflect.Value) {
	for i := range target.NumField() {
		field, value := target.Type().Field(i), target.Field(i)

		if field.Anonymous && value.Kind() == reflect.Struct {
			bindIdentity(caller, value)

			continue
		}

		name, ok := field.Tag.Lookup("identity")
		if !ok {
			continue
		}

		bound := reflect.ValueOf(identityField(caller, name))
		if !bound.IsValid() || bound.IsZero() {
			continue
		}

		value.Set(bound.Convert(value.Type()))
	}
}
