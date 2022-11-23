package validator

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	Validate *validator.Validate
}

// New creates a new ShellHub validator.
//
// The ShellHub validator contains validations rules to name, username, email, password, etc.
func New() *Validator {
	validate := validator.New()
	validate.RegisterValidation(TagRegexp, regexpValidator)     //nolint:errcheck
	validate.RegisterValidation(TagUsername, usernameValidator) //nolint:errcheck

	return &Validator{
		Validate: validate,
	}
}

// Var validates a variable using a ShellHub validation's tags.
func (v *Validator) Var(value, tags string) (bool, error) {
	if err := v.Validate.Var(value, tags); err != nil {
		return false, fmt.Errorf("%s is invalid for %s tags: %w", value, tags, err)
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

// GetInvalidFieldsFromErr gets the invalids frields from a error returned by Struct function.
func GetInvalidFieldsFromErr(err error) ([]string, error) {
	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		return nil, ErrInvalidError
	}

	fields := make([]string, len(errs))
	for index, err := range errs {
		fields[index] = err.Field()
	}

	return fields, nil
}
