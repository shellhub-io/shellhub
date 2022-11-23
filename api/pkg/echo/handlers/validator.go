package handlers

import (
	errors "github.com/shellhub-io/shellhub/api/routes/errors"
	"github.com/shellhub-io/shellhub/pkg/validator"
)

type Validator struct {
	validator *validator.Validator
}

func NewValidator() *Validator {
	return &Validator{validator: validator.New()}
}

func (v *Validator) Validate(structure interface{}) error {
	if ok, err := v.validator.Struct(structure); !ok {
		fields, err := validator.GetInvalidFieldsFromErr(err)
		if err != nil {
			return err
		}

		return errors.NewErrInvalidEntity(fields)
	}

	return nil
}
