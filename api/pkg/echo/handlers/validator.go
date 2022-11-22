package handlers

import (
	"regexp"

	"github.com/go-playground/validator/v10"
	errors "github.com/shellhub-io/shellhub/api/routes/errors"
)

type Validator struct {
	validator *validator.Validate
}

func NewValidator() *Validator {
	validate := validator.New()
	_ = validate.RegisterValidation("regexp", func(fl validator.FieldLevel) bool {
		_, err := regexp.Compile(fl.Field().String())

		return err == nil
	})

	_ = validate.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		return regexp.MustCompile(`^([a-z0-9-_.@]){3,30}$`).MatchString(fl.Field().String())
	})

	return &Validator{validator: validate}
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
