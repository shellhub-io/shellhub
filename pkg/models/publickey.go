package models

import (
	"regexp"
	"slices"
	"time"

	"github.com/go-playground/validator/v10"
)

// PublicKeyFilter contains the filter rule of a Public Key.
//
// A PublicKeyFilter can contain either Hostname, string, or Tags, slice of strings never both.
type PublicKeyFilter struct {
	Hostname string `json:"hostname,omitempty" validate:"required_without=Tags,excluded_with=Tags,regexp"`
	Taggable `json:",inline"`
}

// Matches reports whether the given device satisfies the filter. A filter is
// either a hostname regexp matched against the device name, or a tag set matched
// by intersection against the device's tag ids; an empty filter matches every
// device. It is the shared device-selector matcher used by both the public-key
// ACL and Access Policies.
//
// The device must already carry its tag ids (Taggable.TagIDs) for the tag
// branch; callers resolving a device from an agent-sent payload must populate
// them first, since the agent does not send tag ids.
func (f PublicKeyFilter) Matches(device *Device) (bool, error) {
	switch {
	case f.Hostname != "":
		return regexp.MatchString(f.Hostname, device.Name)
	case len(f.TagIDs) > 0:
		for _, tagID := range f.TagIDs {
			if slices.Contains(device.TagIDs, tagID) {
				return true, nil
			}
		}

		return false, nil
	default:
		return true, nil
	}
}

type PublicKeyFields struct {
	Name     string          `json:"name"`
	Username string          `json:"username" validate:"regexp"`
	Filter   PublicKeyFilter `json:"filter" validate:"required"`
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
	Data        []byte    `json:"data"`
	Fingerprint string    `json:"fingerprint"`
	CreatedAt   time.Time `json:"created_at"`
	TenantID    string    `json:"tenant_id"`
	PublicKeyFields
}

type PublicKeyUpdate struct {
	PublicKeyFields
}

type PublicKeyAuthRequest struct {
	Fingerprint string `json:"fingerprint"`
	Data        string `json:"data"`
}

type PublicKeyAuthResponse struct {
	Signature string `json:"signature"`
}
