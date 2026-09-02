package validator

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"reflect"
	"regexp"

	"github.com/go-playground/validator/v10"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
)

// The errors validation returns, distinguishing a whole struct that failed from a single value.
// Both wrap the underlying library's error, so the failing field is still reachable.
var (
	ErrStructureInvalid = errors.New("invalid structure")
	ErrVarInvalid       = errors.New("invalid var")
)

// Rule is a struct that contains a validation rule.
type Rule struct {
	Tag     string
	Handler func(field validator.FieldLevel) bool
	Error   error
}

// Tag is the rule used to validate a variable or a structure's field.
type Tag string

const (
	// RegexpTag indicates that the regexp must be valide.
	RegexpTag = "regexp"
	// NameTag contains the rule to validate the user's name.
	NameTag = "name"
	// UserNameTag contains the rule to validate the user's username.
	UserNameTag = "username"
	// UserPasswordTag contains the rule to validate the user's password.
	UserPasswordTag = "password"
	// DeviceNameTag contains the rule to validate the device's name.
	DeviceNameTag = "device_name"
	// DeviceIdentityTag contains the rule to validate the device's identity.
	DeviceIdentityTag = "device_identity"
	// PrivateKeyPEMTag contains the rule to validate a private key.
	PrivateKeyPEMTag = "privateKeyPEM"
	// CertPEMTag contains the rule to validate a certificate.
	CertPEMTag = "certPEM"
	// APIKeyNameTag contains the rule to validate an API key's name.
	APIKeyNameTag = "api-key_name"
	// APIKeyExpiresAtTag contains the rule to validate an API key's expiration value.
	APIKeyExpiresAtTag = "api-key_expires-at" //nolint:gosec // G101: not a credential, this is a validator tag name
	// InstanceAPIKeyExpiresAtTag contains the rule to validate an instance API key's expiration
	// value. It is the API key rule without -1: an instance key must always expire.
	InstanceAPIKeyExpiresAtTag = "instance-api-key_expires-at"
	// MemberRoleTag contains the rule to validate a namespace member's role.
	MemberRoleTag = "member_role"
)

// Rules is a slice that contains all validation rules.
var Rules = []Rule{
	{
		Tag: RegexpTag,
		Handler: func(field validator.FieldLevel) bool {
			_, err := regexp.Compile(field.Field().String())

			return err == nil
		},
		Error: errors.New("the regexp is invalid"),
	},
	{
		Tag: NameTag,
		Handler: func(field validator.FieldLevel) bool {
			return regexp.MustCompile(`^(.){1,64}$`).MatchString(field.Field().String())
		},
		Error: errors.New("the name must be between 1 and 64 characters"),
	},
	{
		Tag: UserNameTag,
		Handler: func(field validator.FieldLevel) bool {
			return regexp.MustCompile(`^([a-z0-9-_.@]){3,32}$`).MatchString(field.Field().String())
		},
		Error: errors.New("the username must be between 3 and 32 characters, and can only contain letters, numbers, and the following characters: -_.@"),
	},
	{
		Tag: UserPasswordTag,
		Handler: func(field validator.FieldLevel) bool {
			return regexp.MustCompile(`^(.){5,32}$`).MatchString(field.Field().String())
		},
		Error: errors.New("the password cannot be empty and must be between 5 and 32 characters"),
	},
	{
		Tag: DeviceNameTag,
		Handler: func(field validator.FieldLevel) bool {
			return regexp.MustCompile(`^([a-zA-Z0-9_-]){1,64}$`).MatchString(field.Field().String())
		},
		Error: errors.New("the device name can only contain `_`, `-` and alpha numeric characters"),
	},
	{
		Tag: DeviceIdentityTag,
		Handler: func(field validator.FieldLevel) bool {
			return regexp.MustCompile(`^([a-zA-Z0-9:_.-]){1,64}$`).MatchString(field.Field().String())
		},
		Error: errors.New("the device identity can only contain `:`, `.`, `_`, `-` and alpha numeric characters"),
	},
	{
		Tag: APIKeyNameTag,
		Handler: func(field validator.FieldLevel) bool {
			return regexp.MustCompile(`^[a-zA-Z0-9_-]{3,20}$`).MatchString(field.Field().String())
		},
		Error: errors.New("name must be between 3 and 20 characters and can only contain letters, numbers, `-` and `_`"),
	},
	{
		Tag: APIKeyExpiresAtTag,
		Handler: func(field validator.FieldLevel) bool {
			if !field.Field().CanInt() {
				return false
			}

			expiresAt := field.Field().Int()

			return expiresAt == -1 || expiresAt == 30 || expiresAt == 60 || expiresAt == 90 || expiresAt == 365
		},
		Error: errors.New("expires_at must be in [ -1 30 60 90 365 ]"),
	},
	{
		Tag: InstanceAPIKeyExpiresAtTag,
		Handler: func(field validator.FieldLevel) bool {
			if !field.Field().CanInt() {
				return false
			}

			expiresAt := field.Field().Int()

			return expiresAt == 30 || expiresAt == 60 || expiresAt == 90 || expiresAt == 365
		},
		Error: errors.New("expires_at must be in [ 30 60 90 365 ]"),
	},
	{
		Tag: MemberRoleTag,
		Handler: func(field validator.FieldLevel) bool {
			return authorizer.RoleFromString(field.Field().String()) != authorizer.RoleInvalid
		},
		Error: errors.New("role must be \"owner\", \"administrator\", \"operator\" or \"observer\""),
	},
	{
		Tag: PrivateKeyPEMTag,
		Handler: func(field validator.FieldLevel) bool {
			block, _ := pem.Decode([]byte(field.Field().String()))
			if block == nil {
				return false
			}

			key, err := x509.ParsePKCS8PrivateKey(block.Bytes)

			return err == nil && key != nil
		},
		Error: errors.New("the private key is invalid"),
	},
	{
		Tag: CertPEMTag,
		Handler: func(field validator.FieldLevel) bool {
			block, _ := pem.Decode([]byte(field.Field().String()))
			if block == nil {
				return false
			}

			cert, err := x509.ParseCertificate(block.Bytes)

			return err == nil && cert != nil
		},
		Error: errors.New("the cert is invalid"),
	},
}

// Validator is the ShellHub validator.
// It uses the go-playground/validator package internally and add custom validation rules for ShellHub types.
type Validator struct {
	Validate *validator.Validate
}

// New creates a new ShellHub validator.
//
// The ShellHub validator contains custom validation rules for ShellHub types.
func New() *Validator {
	validate := validator.New()

	for _, rule := range Rules {
		validate.RegisterValidation(rule.Tag, rule.Handler) //nolint:errcheck
	}

	return &Validator{
		Validate: validate,
	}
}

// Var validates a variable using a ShellHub validation's tags.
func (v *Validator) Var(value any, tag Tag) (bool, error) {
	if err := v.Validate.Var(value, string(tag)); err != nil {
		return false, ErrVarInvalid
	}

	return true, nil
}

// Struct validates a structure using ShellHub validation's tags.
func (v *Validator) Struct(structure any) (bool, error) {
	if err := v.Validate.Struct(structure); err != nil {
		return false, ErrStructureInvalid
	}

	return true, nil
}

// StructWithFields validades a structure using ShellHub validation's tags, returnig the invalid fields and its tags.
func (v *Validator) StructWithFields(structure any) (bool, map[string]any, error) {
	if err := v.Validate.Struct(structure); err != nil {
		fields := make(map[string]any, 0)

		var errs validator.ValidationErrors
		if errors.As(err, &errs) {
			for _, e := range errs {
				fields[e.Field()] = e.Tag()
			}
		}

		return false, fields, ErrStructureInvalid
	}

	return true, nil, nil
}

// GetTagFromStructure returns the validation's tag from structure.
func GetTagFromStructure(structure any, field string) (Tag, bool) {
	kind := reflect.TypeOf(structure)
	name, ok := kind.FieldByName(field)
	if !ok {
		return "", false
	}

	return Tag(name.Tag.Get("validate")), true
}
