package validator

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

// GetInvalidFieldsFromErr gets the invalids fields from an error returned by Struct function.
// If the error is not from a field validation, it returns an error. Otherwise, it returns the invalid fields.
func GetInvalidFieldsFromErr(err error) ([]string, error) {
	err = errors.Unwrap(err)

	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		return nil, fmt.Errorf("the error is not from a field validation: %w", err)
	}

	fields := make([]string, len(errs))

	for _, err := range errs {
		fields = append(fields, err.Field())
	}

	return fields, nil
}

// GetInvalidValuesFromErr gets the invalids values from an error returned by Struct function.
func GetInvalidValuesFromErr(err error) ([]string, error) {
	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		return nil, fmt.Errorf("the error is not from a field validation: %w", err)
	}

	values := make([]string, len(errs))
	for index, err := range errs {
		values[index] = err.Value().(string)
	}

	return values, nil
}

// GetInvalidFieldValueFromErr gets the invalid field value from an error returned by Struct function.
func GetInvalidFieldValueFromErr(err error) (map[string]string, error) {
	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		return nil, fmt.Errorf("the error is not from a field validation: %w", err)
	}

	fields := make(map[string]string, len(errs))
	for _, err := range errs {
		fields[err.Field()] = err.Value().(string)
	}

	return fields, nil
}

// GetFirstFieldError gets the first invalid field error from an error returned by Struct function.
func GetFirstFieldError(err error) error {
	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		return fmt.Errorf("the error is not from a field validation: %w", err)
	}

	for _, err := range errs {
		for _, rule := range Rules {
			if err.Tag() == rule.Tag {
				return rule.Error
			}
		}
	}

	return fmt.Errorf("the field %s is invalid for rule %s", errs[0].Field(), errs[0].ActualTag())
}
