package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsEmail(t *testing.T) {
	cases := []struct {
		description string
		email       string
		expected    bool
	}{
		{
			description: "fails with an empty string",
			email:       "",
			expected:    false,
		},
		{
			description: "fails without the '@' symbol",
			email:       "testexample.com",
			expected:    false,
		},
		{
			description: "fails without a domain",
			email:       "test@",
			expected:    false,
		},
		{
			description: "fails with only a domain",
			email:       "@example.com",
			expected:    false,
		},
		{
			description: "fails with spaces",
			email:       "test @example.com",
			expected:    false,
		},
		{
			description: "fails with invalid characters",
			email:       "test@example!.com",
			expected:    false,
		},
		{
			description: "fails with multiple '@' symbols",
			email:       "te@st@example.com",
			expected:    false,
		},
		{
			description: "fails without a top-level domain",
			email:       "test@example",
			expected:    false,
		},
		{
			description: "fails with consecutive dots in domain",
			email:       "test@example..com",
			expected:    false,
		},
		{
			description: "fails with leading dot in domain",
			email:       "test@.example.com",
			expected:    false,
		},
		{
			description: "succeeds with a valid email format",
			email:       "test@example.com",
			expected:    true,
		},
		{
			description: "succeeds with a subdomain",
			email:       "test@sub.example.com",
			expected:    true,
		},
		{
			description: "succeeds with numbers in the local part",
			email:       "test123@example.com",
			expected:    true,
		},
		{
			description: "succeeds with a valid email with hyphens in domain",
			email:       "test@exa-mple.com",
			expected:    true,
		},
		{
			description: "succeeds with a valid .com.br email format",
			email:       "test@example.com.br",
			expected:    true,
		},
		{
			description: "succeeds with a subdomain in .com.br",
			email:       "test@sub.example.com.br",
			expected:    true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.Equal(t, tc.expected, IsEmail(tc.email))
		})
	}
}
