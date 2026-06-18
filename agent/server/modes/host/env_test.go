package host

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestAcceptClientEnv(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		// Accepted entries
		{
			name:     "accept LANG",
			input:    []string{"LANG=en_US.UTF-8"},
			expected: []string{"LANG=en_US.UTF-8"},
		},
		{
			name:     "accept LC_ALL",
			input:    []string{"LC_ALL=C"},
			expected: []string{"LC_ALL=C"},
		},
		{
			name:     "accept LC_CTYPE",
			input:    []string{"LC_CTYPE=UTF-8"},
			expected: []string{"LC_CTYPE=UTF-8"},
		},
		{
			name:     "accept LC_PAPER",
			input:    []string{"LC_PAPER=x"},
			expected: []string{"LC_PAPER=x"},
		},
		{
			name:     "accept LC_ with empty suffix (LC_=value)",
			input:    []string{"LC_="},
			expected: []string{"LC_="},
		},
		// Dropped entries
		{
			name:     "drop LD_PRELOAD",
			input:    []string{"LD_PRELOAD=/"},
			expected: []string{},
		},
		{
			name:     "drop LD_AUDIT",
			input:    []string{"LD_AUDIT=/"},
			expected: []string{},
		},
		{
			name:     "drop GODEBUG",
			input:    []string{"GODEBUG=inittrace=1"},
			expected: []string{},
		},
		{
			name:     "drop PATH",
			input:    []string{"PATH=/tmp/evil"},
			expected: []string{},
		},
		{
			name:     "drop BASH_ENV",
			input:    []string{"BASH_ENV=/tmp/x"},
			expected: []string{},
		},
		{
			name:     "drop TERM",
			input:    []string{"TERM=evil"},
			expected: []string{},
		},
		{
			name:     "drop HOME",
			input:    []string{"HOME=/tmp"},
			expected: []string{},
		},
		{
			name:     "drop USER",
			input:    []string{"USER=root"},
			expected: []string{},
		},
		{
			name:     "drop SSH_AUTH_SOCK",
			input:    []string{"SSH_AUTH_SOCK=/tmp/x"},
			expected: []string{},
		},
		{
			name:     "drop arbitrary variable MYVAR",
			input:    []string{"MYVAR=value"},
			expected: []string{},
		},
		{
			name:     "drop LANGUAGE (not LC_ prefix match)",
			input:    []string{"LANGUAGE=en"},
			expected: []string{},
		},
		{
			name:     "drop LCALL (not LC_ prefix)",
			input:    []string{"LCALL=x"},
			expected: []string{},
		},
		{
			name:     "drop LD_ prefix variable",
			input:    []string{"LD_=x"},
			expected: []string{},
		},
		// Edge cases
		{
			name:     "drop malformed entry without equals sign",
			input:    []string{"FOO"},
			expected: []string{},
		},
		{
			name:     "drop empty-name entry (=value)",
			input:    []string{"=value"},
			expected: []string{},
		},
		{
			name:     "drop entry with NUL byte",
			input:    []string{"LANG=en\x00evil"},
			expected: []string{},
		},
		{
			name:     "nil input returns nil",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty input returns empty",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "order preserved for multiple accepted entries",
			input:    []string{"LC_ALL=C", "MYVAR=bad", "LANG=en_US.UTF-8", "LD_PRELOAD=/", "LC_CTYPE=UTF-8"},
			expected: []string{"LC_ALL=C", "LANG=en_US.UTF-8", "LC_CTYPE=UTF-8"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := acceptClientEnv(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}
