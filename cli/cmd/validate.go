package cmd

import "github.com/shellhub-io/shellhub/pkg/validator"

// validateInput validates the provided input struct against its validation tags.
// It returns an error if any field fails validation.
func validateInput(input any) error {
	v := validator.New()
	ok, fields, err := v.StructWithFields(input)
	if !ok || err != nil {
		return mapValidationError(fields)
	}

	return nil
}

// mapValidationError maps a validation fields map returned by the validator to a
// human-readable sentinel error based on the first failing field.
func mapValidationError(fields map[string]any) error {
	for _, field := range []string{"Username", "Owner", "Password", "Email", "Namespace", "TenantID", "Type"} {
		if _, ok := fields[field]; !ok {
			continue
		}
		switch field {
		case "Username", "Owner":
			return ErrInvalidUsername
		case "Password":
			return ErrInvalidPassword
		case "Email":
			return ErrInvalidEmail
		case "Namespace":
			return ErrInvalidNamespace
		case "Type":
			return ErrInvalidType
		case "TenantID":
			return ErrInvalidTenantID
		}
	}

	return ErrInvalidInput
}
