package models

import "github.com/go-playground/validator/v10"

type ID struct {
	ID string
}

func getInvalidFields(err error) ([]string, error) {
	f := []string{}
	for _, err := range err.(validator.ValidationErrors) {
		f = append(f, err.Field())
	}

	return f, err
}

func Validate(data interface{}) ([]string, error) {
	if err := validator.New().Struct(data); err != nil {
		return getInvalidFields(err)
	}

	return nil, nil
}
