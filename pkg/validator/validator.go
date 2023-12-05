package validator

import (
	"fmt"
	"regexp"

	"github.com/go-playground/validator/v10"
)

// Rule is a struct that contains a validation rule.
type Rule struct {
	Tag     string
	Handler func(field validator.FieldLevel) bool
	Error   error
}

// Rules is a slice that contains all validation rules.
var Rules = []Rule{
	{
		Tag: "regexp",
		Handler: func(field validator.FieldLevel) bool {
			_, err := regexp.Compile(field.Field().String())

			return err == nil
		},
		Error: fmt.Errorf("the regexp is invalid"),
	},
	{
		Tag: "username",
		Handler: func(field validator.FieldLevel) bool {
			return regexp.MustCompile(`^([a-zA-Z0-9-_.@]){3,30}$`).MatchString(field.Field().String())
		},
		Error: fmt.Errorf("the username must be between 3 and 30 characters, and can only contain letters, numbers, and the following characters: -_.@"),
	},
	{
		Tag: "password",
		Handler: func(field validator.FieldLevel) bool {
			return regexp.MustCompile(`^(.){5,30}$`).MatchString(field.Field().String())
		},
		Error: fmt.Errorf("the password cannot be empty and must be between 5 and 30 characters"),
	},
	{
		Tag: "device_name",
		Handler: func(field validator.FieldLevel) bool {
			return regexp.MustCompile(`^([a-zA-Z0-9_.-] ){1,64}$`).MatchString(field.Field().String())
		},
		Error: fmt.Errorf("the device name can only contain `_`, `.` and alpha numeric characters"),
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
func (v *Validator) Var(value, tag string) (bool, error) {
	if err := v.Validate.Var(value, tag); err != nil {
		return false, fmt.Errorf("invalid variable: %w", fmt.Errorf("invalid validation on value %s using tag %s", value, tag))
	}

	return true, nil
}

// Struct validates a structure using ShellHub validation's tags.
func (v *Validator) Struct(structure interface{}) (bool, error) {
	if err := v.Validate.Struct(structure); err != nil {
		return false, fmt.Errorf("invalid structure: %w", err)
	}

	return true, nil
}
