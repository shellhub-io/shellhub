package validator

// IsEmail validates that the input is a valid email.
func IsEmail(input string) bool {
	type validate struct {
		Email string `validate:"required,email"`
	}

	v := New()
	if ok, err := v.Struct(validate{Email: input}); !ok || err != nil {
		return false
	}

	return true
}
