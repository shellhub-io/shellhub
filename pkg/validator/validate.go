package validator

import (
	"errors"
	"strings"

	"gopkg.in/go-playground/validator.v9"
)

var (
	ErrorStruct     = errors.New("the struct provided couldn't be checked")
	ErrorBadRequest = errors.New("bad request")
)

type InvalidField struct {
	Name  string
	Kind  string
	Param string
	Extra string
}

func CheckValidation(data interface{}) ([]InvalidField, error) {
	var invalidFields []InvalidField

	if err := validator.New().Struct(data); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			invalidFields = append(invalidFields, InvalidField{strings.ToLower(err.StructField()), "invalid", err.Tag(), err.Param()})
		}

		return invalidFields, ErrorBadRequest
	}

	return invalidFields, nil
}
