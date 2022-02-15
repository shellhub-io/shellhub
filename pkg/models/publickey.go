package models

import (
	"regexp"
	"time"

	"github.com/go-playground/validator/v10"
)

// PublicKeyFilter contains the filter rule of a Public Key.
//
// A PublicKeyFilter can contain either Hostname, string, or Tags, slice of strings never both.
type PublicKeyFilter struct {
	Hostname string   `json:"hostname,omitempty" bson:"hostname,omitempty" validate:"required_without=Tags,excluded_with=Tags,regexp"`
	Tags     []string `json:"tags,omitempty" bson:"tags,omitempty" validate:"required_without=Hostname,excluded_with=Hostname,max=3,dive,min=3,max=255,alphanum,ascii,excludes=/@&:"`
}

type PublicKeyFields struct {
	Name     string          `json:"name"`
	Username string          `json:"username" bson:"username" validate:"regexp"`
	Filter   PublicKeyFilter `json:"filter" bson:"filter" validate:"required"`
}

func (p *PublicKeyFields) Validate() error {
	v := validator.New()

	_ = v.RegisterValidation("regexp", func(fl validator.FieldLevel) bool {
		_, err := regexp.Compile(fl.Field().String())

		return err == nil
	})

	return v.Struct(p)
}

type PublicKey struct {
	Data            []byte    `json:"data"`
	Fingerprint     string    `json:"fingerprint"`
	CreatedAt       time.Time `json:"created_at" bson:"created_at"`
	TenantID        string    `json:"tenant_id" bson:"tenant_id"`
	PublicKeyFields `bson:",inline"`
}

type PublicKeyUpdate struct {
	PublicKeyFields `bson:",inline"`
}

type PublicKeyAuthRequest struct {
	Fingerprint string `json:"fingerprint"`
	Data        string `json:"data"`
}

type PublicKeyAuthResponse struct {
	Signature string `json:"signature"`
}
