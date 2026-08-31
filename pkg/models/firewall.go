package models

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// FirewallFilter contains the filter rule of a Public Key.
//
// A FirewallFilter can contain either Hostname, string, or Tags, slice of strings never both.
type FirewallFilter struct {
	Hostname string   `json:"hostname,omitempty" validate:"required_without=Tags,excluded_with=Tags,regexp"`
	Tags     []string `json:"tags,omitempty" validate:"required_without=Hostname,excluded_with=Hostname,max=3,unique,dive,min=3,max=255,alphanum,ascii,excludes=/@&:"`
}

// FirewallRuleFields is the editable part of a firewall rule. Rules are evaluated in Priority
// order and the first match decides, so Action and Priority together are what a reader has to hold
// in mind. An inactive rule is skipped rather than treated as a deny.
type FirewallRuleFields struct {
	Priority int            `json:"priority"`
	Action   string         `json:"action" validate:"required,oneof=allow deny"`
	Active   bool           `json:"active"`
	SourceIP string         `json:"source_ip" validate:"required,regexp"`
	Username string         `json:"username" validate:"required,regexp"`
	Filter   FirewallFilter `json:"filter" validate:"required"`
}

// Validate checks the fields, including that SourceIP, Username and the filter's Hostname compile
// as regular expressions — they are patterns, and an uncompilable one would match nothing while
// looking like a rule that works.
func (f *FirewallRuleFields) Validate() error {
	v := validator.New()

	_ = v.RegisterValidation("regexp", func(fl validator.FieldLevel) bool {
		_, err := regexp.Compile(fl.Field().String())

		return err == nil
	})

	return v.Struct(f)
}

// FirewallRule is a stored rule, scoped to one namespace.
type FirewallRule struct {
	ID       string `json:"id,omitempty"`
	TenantID string `json:"tenant_id"`
	FirewallRuleFields
}

// FirewallRuleUpdate is a full replacement of a rule's fields: everything editable is sent, so an
// omitted field is a cleared field, not an unchanged one.
type FirewallRuleUpdate struct {
	FirewallRuleFields
}
