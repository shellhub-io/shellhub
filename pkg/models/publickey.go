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
// Hostname is a regexp matched against the whole device name; see MatchPattern.
type PublicKeyFilter struct {
	Hostname string `json:"hostname,omitempty" validate:"required_without=Tags,excluded_with=Tags,regexp"`
	Taggable `json:",inline"`
}

// Matches reports whether the given device satisfies the filter. A filter is
// either a hostname pattern matched against the whole device name (see
// MatchPattern), or a tag set matched by intersection against the device's tag
// ids; an empty filter matches every device. It is the shared device-selector
// matcher used by both the public-key ACL and Access Policies.
//
// The device must already carry its tag ids (Taggable.TagIDs) for the tag
// branch; callers resolving a device from an agent-sent payload must populate
// them first, since the agent does not send tag ids.
func (f PublicKeyFilter) Matches(device *Device) (bool, error) {
	switch {
	case f.Hostname != "":
		return MatchPattern(f.Hostname, device.Name)
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

// PublicKeyFields is the editable part of a public key: who it logs in as and which devices it
// reaches. The key material itself is not here, so an update can change the rule without touching
// the key.
type PublicKeyFields struct {
	Name     string          `json:"name"`
	Username string          `json:"username" validate:"regexp"`
	Filter   PublicKeyFilter `json:"filter" validate:"required"`
}

// Validate checks the fields, including that Username and the filter's Hostname compile as regular
// expressions — they are patterns, not literals, and an uncompilable one would silently match
// nothing. It does not require them to be anchored: MatchPattern anchors them at match time.
func (p *PublicKeyFields) Validate() error {
	v := validator.New()

	_ = v.RegisterValidation("regexp", func(fl validator.FieldLevel) bool {
		_, err := regexp.Compile(fl.Field().String())

		return err == nil
	})

	return v.Struct(p)
}

// PublicKey is an SSH key registered in a namespace, and the ACL entry attached to it. Data is the
// key in wire format; Fingerprint is what the SSH gateway looks it up by during authentication.
type PublicKey struct {
	Data        []byte    `json:"data"`
	Fingerprint string    `json:"fingerprint"`
	CreatedAt   time.Time `json:"created_at"`
	TenantID    string    `json:"tenant_id"`
	PublicKeyFields
}

// PublicKeyUpdate is what an edit may change: the rule, never the key material. Replacing a key
// means deleting and re-adding it, so its fingerprint stays the identity.
type PublicKeyUpdate struct {
	PublicKeyFields
}

// PublicKeyAuthRequest is what the SSH gateway sends to have a key challenge signed on behalf of a
// session.
type PublicKeyAuthRequest struct {
	Fingerprint string `json:"fingerprint"`
	Data        string `json:"data"`
}

// PublicKeyAuthResponse carries the signature produced for a PublicKeyAuthRequest.
type PublicKeyAuthResponse struct {
	Signature string `json:"signature"`
}
