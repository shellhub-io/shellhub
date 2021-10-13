package validator

import (
	"strings"

	validator "gopkg.in/go-playground/validator.v9"
)

func validate(err error) ([]string, error) {
	invalidFields := []string{}
	for _, err := range err.(validator.ValidationErrors) {
		invalidFields = append(invalidFields, strings.ToLower(err.Field()))
	}

	return invalidFields, ErrBadRequest
}

func ValidateStruct(data interface{}) ([]string, error) {
	if err := validator.New().Struct(data); err != nil {
		return validate(err)
	}

	return nil, nil
}

func ValidateVar(data interface{}, tag string) ([]string, error) {
	if err := validator.New().Var(data, tag); err != nil {
		return validate(err)
	}

	return nil, nil
}
