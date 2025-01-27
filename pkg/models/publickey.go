package models

import (
	"regexp"
	"time"

	"github.com/go-playground/validator/v10"
)

type PublicKey struct {
	Data            []byte    `json:"data"`
	Fingerprint     string    `json:"fingerprint"`
	CreatedAt       time.Time `json:"created_at" bson:"created_at"`
	TenantID        string    `json:"tenant_id" bson:"tenant_id"`
	PublicKeyFields `bson:",inline"`
}

type PublicKeyFields struct {
	Name     string          `json:"name"`
	Username string          `json:"username" bson:"username" validate:"regexp"`
	Filter   PublicKeyFilter `json:"filter" bson:"filter" validate:"required"`
}

// PublicKeyFilter contains the filter rule of a Public Key.
//
// A PublicKeyFilter can contain either Hostname, string, or Tags, slice of strings never both.
type PublicKeyFilter struct {
	Hostname string `json:"hostname,omitempty" bson:"hostname,omitempty" validate:"required_without=Tags,excluded_with=Tags,regexp"`

	// TagsID contains the IDs of associated tags. It is only used internally for database storage and
	// relationships, and is not exposed in JSON responses.
	TagsID []string `json:"-" bson:"tags,omitempty"`
	// Tags represents the full tag objects associated with this device. This field is populated from
	// [PublicKeyFilter.TagsID] when retrieving from the database and is only used for JSON serialization. It is
	// not stored directly in the database as it is.
	Tags []Tag `json:"tags,omitempty" bson:"-"`
}

func (p *PublicKeyFields) Validate() error {
	v := validator.New()

	_ = v.RegisterValidation("regexp", func(fl validator.FieldLevel) bool {
		_, err := regexp.Compile(fl.Field().String())

		return err == nil
	})

	return v.Struct(p)
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
