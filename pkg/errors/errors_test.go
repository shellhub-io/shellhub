package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestGetLastError(t *testing.T) {
	tests := []struct {
		description string
		err         error
		expected    error
	}{
		{
			description: "return nil if the error is nil",
			err:         nil,
			expected:    nil,
		},
		{
			description: "return the error wrapped by one layer",
			err:         Wrap(New("second error", "", 0), New("first error", "", 0)),
			expected:    New("first error", "", 0),
		},
		{
			description: "return the error wrapped by two layers",
			err:         Wrap(New("third error", "", 0), Wrap(New("second error", "", 0), New("first error", "", 0))),
			expected:    New("first error", "", 0),
		},
		{
			description: "return the error wrapped by two layers, but the last one is not from Error",
			err: Wrap(
				New("third error", "", 0), Wrap(
					New("second error", "", 0), errors.New("first error not from Error"),
				),
			),
			expected: errors.New("first error not from Error"),
		},
		{
			description: "return the error wrapped by two layers, but the last one is not from Error and has others errors wrapped",
			err: Wrap(
				New("third error", "", 0), Wrap(
					New("second error", "", 0), fmt.Errorf("first error not from Error: %w", errors.New("zero error")),
				),
			),
			expected: fmt.Errorf("first error not from Error: %w", errors.New("zero error")),
		},
	}

	for _, test := range tests {
		local := test
		t.Run(test.description, func(t *testing.T) {
			if last := GetLastError(local.err); (last != nil && local.expected != nil) && last.Error() != local.expected.Error() {
				t.Fatalf("expected %v, got %v", local.expected, last)
			}
		})
	}
}
