package handlers

import (
	"github.com/go-playground/validator/v10"
	errors "github.com/shellhub-io/shellhub/api/routes/errors"
	shellhub "github.com/shellhub-io/shellhub/pkg/validator"
)

type Validator struct {
	validator *validator.Validate
}

func NewValidator() *Validator {
	return &Validator{validator: shellhub.GetInstance()}
}

func (v *Validator) Validate(s interface{}) error {
	if err := v.validator.Struct(s); err != nil {
		fields := make(map[string]string)
		for _, err := range err.(validator.ValidationErrors) {
			fields[err.Field()] = err.Tag()
		}

		return errors.NewErrInvalidEntity(fields)
	}

	return nil
}
