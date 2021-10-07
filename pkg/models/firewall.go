package models

import (
	"regexp"

	validator "gopkg.in/go-playground/validator.v9"
)

type FirewallRuleFields struct {
	Priority int    `json:"priority"`
	Action   string `json:"action" validate:"required,oneof=allow deny"`
	Active   bool   `json:"active"`
	SourceIP string `json:"source_ip" bson:"source_ip" validate:"required,regexp"`
	Username string `json:"username" validate:"required,regexp"`
	Hostname string `json:"hostname" validate:"required,regexp"`
}

func (f *FirewallRuleFields) Validate() error {
	v := validator.New()

	_ = v.RegisterValidation("regexp", func(fl validator.FieldLevel) bool {
		_, err := regexp.Compile(fl.Field().String())

		return err == nil
	})

	return v.Struct(f)
}

type FirewallRule struct {
	ID                 string `json:"id,omitempty" bson:"_id,omitempty"`
	TenantID           string `json:"tenant_id" bson:"tenant_id"`
	FirewallRuleFields `bson:",inline"`
}

type FirewallRuleUpdate struct {
	FirewallRuleFields `bson:",inline"`
}
