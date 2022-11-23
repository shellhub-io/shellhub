package validator

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// regexpValidator is a function used to validate a regexp.
func regexpValidator(field validator.FieldLevel) bool {
	_, err := regexp.Compile(field.Field().String())

	return err == nil
}

// usernameValidator is a function used to validate ShellHub's username.
func usernameValidator(field validator.FieldLevel) bool {
	return regexp.MustCompile(`^([a-zA-Z0-9-_.@]){3,30}$`).MatchString(field.Field().String())
}
