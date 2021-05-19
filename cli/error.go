package main

import (
	"errors"
	"fmt"
	"strings"

	"gopkg.in/go-playground/validator.v9"
)

var (
	ErrCreateNewUser          = errors.New("failed to create a new user")
	ErrCreateNewNamespace     = errors.New("failed to create a new namespace")
	ErrDuplicateUser          = errors.New("user already exists")
	ErrDuplicateNamespace     = errors.New("namespace already exists")
	ErrChangePassword         = errors.New("failed to reset the user password")
	ErrUserNotFound           = errors.New("user not found")
	ErrNamespaceNotFound      = errors.New("namespace not found")
	ErrFailedAddNamespaceUser = errors.New("failed to add the namespace for the user")
	ErrFailedDeleteUser       = errors.New("failed to delete the user")
	ErrFailedDeleteNamespace  = errors.New("failed to delete the namespace")
	ErrFailedRemoveMember     = errors.New("failed to remove member from the namespace")
	ErrFound                  = errors.New("errors has been founded")
)

func validateParameters(data Arguments) error {
	var errStrings []string

	if err := validator.New().Struct(data); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			switch err.Tag() {
			case "min":
				errStrings = append(errStrings, fmt.Errorf("minimum characters for %v is %v", strings.ToLower(err.StructField()), err.Param()).Error())
			case "max":
				errStrings = append(errStrings, fmt.Errorf("maximum characters for %v is %v", strings.ToLower(err.StructField()), err.Param()).Error())
			case "alphanum":
				errStrings = append(errStrings, fmt.Errorf("the %v is not a alphanumeric character", strings.ToLower(err.StructField())).Error())
			case "ascii":
				errStrings = append(errStrings, fmt.Errorf("the %v is not in a ascii standard", strings.ToLower(err.StructField())).Error())
			case "email":
				errStrings = append(errStrings, fmt.Errorf("the format for email is invalid").Error())
			}
		}
	}

	if len(errStrings) > 0 {
		for _, err := range errStrings {
			fmt.Println(err)
		}

		return ErrFound
	}

	return nil
}
