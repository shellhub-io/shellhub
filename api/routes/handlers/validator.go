package handlers

import (
	"regexp"

	"github.com/go-playground/validator/v10"
	"github.com/shellhub-io/shellhub/api/routes"
)

type Validator struct {
	validator *validator.Validate
}

func NewValidator() *Validator {
	v := validator.New()
	_ = v.RegisterValidation("regexp", func(fl validator.FieldLevel) bool {
		_, err := regexp.Compile(fl.Field().String())

		return err == nil
	})

	return &Validator{validator: v}
}

func (v *Validator) Validate(s interface{}) error {
	if err := v.validator.Struct(s); err != nil {
		fields := make(map[string]string)
		for _, err := range err.(validator.ValidationErrors) {
			fields[err.Field()] = err.Tag()
		}

		return routes.NewErrInvalidEntity(fields)
	}

	return nil
}
