package validator

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/go-playground/validator/v10"
)

// GetInvalidFieldsFromErr gets the invalids fields from an error returned by structure validation function.
//
// The returned "map"'s key is the value of JSON tag and its value is the value of the tag "validate" from the
// structure for each invalid field from the error.
//
// The error should be from validator.ValidationErrors.
func GetInvalidFieldsFromErr(structure interface{}, err error) (map[string]string, error) {
	err = errors.Unwrap(err)

	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		return nil, fmt.Errorf("the error is not from a field validation: %w", err)
	}

	fields := map[string]string{}
	for _, err := range errs {
		field, _ := reflect.TypeOf(structure).Elem().FieldByName(err.Field())
		fields[field.Tag.Get("json")] = field.Tag.Get("validate")
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
