package models

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// FirewallFilter contains the filter rule of a Public Key.
//
// A FirewallFilter can contain either Hostname, string, or Tags, slice of strings never both.
type FirewallFilter struct {
	Hostname string   `json:"hostname,omitempty" bson:"hostname,omitempty" validate:"required_without=Tags,excluded_with=Tags,regexp"`
	Tags     []string `json:"tags,omitempty" bson:"tags,omitempty" validate:"required_without=Hostname,excluded_with=Hostname,max=3,dive,min=3,max=255,alphanum,ascii,excludes=/@&:"`
}

type FirewallRuleFields struct {
	Priority int            `json:"priority"`
	Action   string         `json:"action" validate:"required,oneof=allow deny"`
	Active   bool           `json:"active"`
	SourceIP string         `json:"source_ip" bson:"source_ip" validate:"required,regexp"`
	Username string         `json:"username" validate:"required,regexp"`
	Filter   FirewallFilter `json:"filter" bson:"filter" validate:"required"`
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
