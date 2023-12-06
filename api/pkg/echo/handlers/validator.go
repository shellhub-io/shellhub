package handlers

import (
	errors "github.com/shellhub-io/shellhub/api/routes/errors"
	"github.com/shellhub-io/shellhub/pkg/validator"
)

type Validator struct {
	validator *validator.Validator
}

// NewValidator creates a new validator for the echo framework from the ShellHub validator.
func NewValidator() *Validator {
	return &Validator{validator: validator.New()}
}

// Validate is called by the echo framework to validate the request body.
// If the request body is invalid, it returns an error with the invalid fields.
func (v *Validator) Validate(structure interface{}) error {
	// Use the ShellHub package validator to validate the request body.
	if ok, err := v.validator.Struct(structure); !ok || err != nil {
		return errors.NewErrInvalidEntity(nil)
	}

	return nil
}
