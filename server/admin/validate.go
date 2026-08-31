package admin

import "github.com/shellhub-io/shellhub/pkg/validator"

func validateInput(input any) error {
	v := validator.New()
	ok, fields, err := v.StructWithFields(input)
	if !ok || err != nil {
		return mapValidationError(fields)
	}

	return nil
}

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
