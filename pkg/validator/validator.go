package validator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"unicode"

	"github.com/go-playground/validator/v10"
)

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
)

// Rules is a slice that contains all validation rules.
var Rules = []Rule{
	{
		Tag: RegexpTag,
		Handler: func(field validator.FieldLevel) bool {
			_, err := regexp.Compile(field.Field().String())

			return err == nil
		},
		Error: fmt.Errorf("the regexp is invalid"),
	},
	{
		Tag: NameTag,
		Handler: func(field validator.FieldLevel) bool {
			return regexp.MustCompile(`^(.){1,64}$`).MatchString(field.Field().String())
		},
		Error: fmt.Errorf("the name must be between 1 and 64 characters"),
	},
	{
		Tag: UserNameTag,
		Handler: func(field validator.FieldLevel) bool {
			return regexp.MustCompile(`^([a-z0-9-_.@]){3,32}$`).MatchString(field.Field().String())
		},
		Error: fmt.Errorf("the username must be between 3 and 32 characters, and can only contain letters, numbers, and the following characters: -_.@"),
	},
	{
		Tag: UserPasswordTag,
		Handler: func(field validator.FieldLevel) bool {
			return regexp.MustCompile(`^(.){5,32}$`).MatchString(field.Field().String())
		},
		Error: fmt.Errorf("the password cannot be empty and must be between 5 and 32 characters"),
	},
	{
		Tag: DeviceNameTag,
		Handler: func(field validator.FieldLevel) bool {
			return regexp.MustCompile(`^([a-zA-Z0-9_-]){1,64}$`).MatchString(field.Field().String())
		},
		Error: fmt.Errorf("the device name can only contain `_`, `-` and alpha numeric characters"),
	},
	// api-key_name reports whether a given string is a valid name for an api key or not. A valid
	// value must be more than 3 characters, less than 20 and does not contains any whitespace.
	{
		Tag: "api-key_name",
		Handler: func(field validator.FieldLevel) bool {
			name := field.Field().String()

			if len(name) < 3 || len(name) > 20 {
				return false
			}

			for _, c := range field.Field().String() {
				if unicode.IsSpace(c) {
					return false
				}
			}

			return true
		},
		Error: fmt.Errorf("name must contain at least 3 characters, at most 20 characters, and no whitespaces"),
	},
	// api-key_expires-at reports whether a given int is in [ 30 60 90 365 -1 ].
	{
		Tag: "api-key_expires-at",
		Handler: func(field validator.FieldLevel) bool {
			if !field.Field().CanInt() {
				return false
			}

			expiresAt := field.Field().Int()

			return expiresAt == -1 || expiresAt == 30 || expiresAt == 60 || expiresAt == 90 || expiresAt == 365
		},
		Error: fmt.Errorf("expires_at must be in [ -1 30 60 90 365 ]"),
	},
	// member_role reports whether a given string is a guard.Role or not
	{
		Tag: "member_role",
		Handler: func(field validator.FieldLevel) bool {
			// TODO: put guard in shellhub/pkg and use it here
			switch field.Field().String() {
			case "owner":
				fallthrough
			case "administrator":
				fallthrough
			case "operator":
				fallthrough
			case "observer":
				return true
			default:
				return false
			}
		},
		Error: fmt.Errorf("role must be \"owner\", \"administrator\", \"operator\" or \"observer\""),
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
func (v *Validator) StructWithFields(structure any) (bool, map[string]interface{}, error) {
	if err := v.Validate.Struct(structure); err != nil {
		fields := make(map[string]interface{}, 0)

		errs := err.(validator.ValidationErrors)

		for _, e := range errs {
			fields[e.Field()] = e.Tag()
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
