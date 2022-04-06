// Package validator has functions to help to validate structures and fields for ShellHub.
package validator

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/shellhub-io/shellhub/pkg/models"
)

// getValidateTag gets the tag string from a structure field.
func getValidateTag(s interface{}, name string) (string, bool) {
	// Gets the structure's type.
	p := reflect.TypeOf(s)
	// Gets the structure's field name.
	f, ok := p.FieldByName(name)
	if !ok {
		return "", false
	}

	// Returns the "validate" fields's tag.
	return f.Tag.Get("validate"), true
}

// getInvalidFields gets the fields reported as invalids.
func getInvalidFields(err error) ([]string, error) {
	f := []string{}
	for _, err := range err.(validator.ValidationErrors) {
		f = append(f, strings.ToLower(err.Field()))
	}

	return f, ErrInvalidFields
}

func ValidateStructFields(data interface{}) (map[string]interface{}, error) {
	if err := validator.New().Struct(data); err != nil {
		d := make(map[string]interface{})
		for _, e := range err.(validator.ValidationErrors) {
			d[e.Field()] = e.Value()
		}

		return d, ErrInvalidFields
	}

	return nil, nil
}

func ValidateStruct(data interface{}) ([]string, error) {
	if err := validator.New().Struct(data); err != nil {
		return getInvalidFields(err)
	}

	return nil, nil
}

func ValidateVar(data interface{}, tag string) ([]string, error) {
	if err := validator.New().Var(data, tag); err != nil {
		return getInvalidFields(err)
	}

	return nil, nil
}

// ValidateField validates if a structure's field is valid.
func ValidateField(structure interface{}, field, value string) bool {
	// Getting tag string from a structure's field.
	t, ok := getValidateTag(structure, field)
	if !ok {
		return false
	}

	// Validating the input data against the tag got.
	if _, err := ValidateVar(value, t); err != nil {
		return false
	}

	return true
}

// ValidateFieldTag validate the data for the field Tag from structure models.Device.
func ValidateFieldTag(tag string) bool {
	const Tag = "required,min=3,max=255,alphanum,ascii,excludes=/@&:"
	if _, err := ValidateVar(tag, Tag); err != nil {
		return false
	}

	return true
}

// ValidateFieldUsername validate the data for the field Username from structure models.UserData.
func ValidateFieldUsername(username string) bool {
	// Field's name that have a tag value.
	const Field = "Username"
	// Structure that contains the field above.
	s := models.UserData{}

	return ValidateField(s, Field, username)
}

// ValidateFieldEmail validate the data for the field Email from structure models.UserData.
func ValidateFieldEmail(email string) bool {
	// Field's name that have a tag value.
	const Field = "Email"
	// Structure that contains the field above.
	s := models.UserData{}

	return ValidateField(s, Field, email)
}

// ValidateFieldPassword validate the data for the field Password from structure models.UserPassword.
func ValidateFieldPassword(password string) bool {
	// Field's name that have a tag value.
	const Field = "Password"
	// Structure that contains the field above.
	s := models.UserPassword{}

	return ValidateField(s, Field, password)
}
