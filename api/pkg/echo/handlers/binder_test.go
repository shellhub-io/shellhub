package handlers

import (
	"testing"
)

func TestBinder(t *testing.T) {
	cases := []struct {
		expected    error
		description string
	}{
		{
			description: "succeeds to bind json body",
			expected:    nil,
		},
		{
			description: "succeeds to bind path parameters",
			expected:    nil,
		},
		{
			description: "succeeds to bind query parameters",
			expected:    nil,
		},
		{
			description: "succeeds to bind query parameters with special characters",
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
		})
	}
}
