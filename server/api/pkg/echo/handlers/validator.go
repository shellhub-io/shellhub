package handlers

import (
	"github.com/shellhub-io/shellhub/pkg/validator"
	errors "github.com/shellhub-io/shellhub/server/api/routes/errors"
)

// Validator adapts ShellHub's validator to echo's, so a request is checked against the same
// rules the services apply to the same models.
type Validator struct {
	validator *validator.Validator
}

// NewValidator creates a new validator for the echo framework from the ShellHub validator.
func NewValidator() *Validator {
	return &Validator{validator: validator.New()}
}

// Validate is called by the echo framework to validate the request body.
// If the request body is invalid, it returns an error with the invalid fields.
func (v *Validator) Validate(structure any) error {
	if ok, err := v.validator.Struct(structure); !ok || err != nil {
		return errors.NewErrInvalidEntity(nil)
	}

	return nil
}
