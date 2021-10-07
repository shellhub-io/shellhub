package validator

import (
	"strings"

	validator "gopkg.in/go-playground/validator.v9"
)

type InvalidField struct {
	Name  string
	Kind  string
	Param string
	Extra string
}

func validate(err error) ([]InvalidField, error) {
	invalidFields := make([]InvalidField, 0, len(err.(validator.ValidationErrors)))
	for _, err := range err.(validator.ValidationErrors) {
		invalidFields = append(invalidFields, InvalidField{strings.ToLower(err.StructField()), "invalid", err.Tag(), err.Param()})
	}

	return invalidFields, ErrBadRequest
}

func ValidateStruct(data interface{}) ([]InvalidField, error) {
	if err := validator.New().Struct(data); err != nil {
		return validate(err)
	}

	return nil, nil
}

func ValidateVar(data interface{}, tag string) ([]InvalidField, error) {
	if err := validator.New().Var(data, tag); err != nil {
		return validate(err)
	}

	return nil, nil
}
