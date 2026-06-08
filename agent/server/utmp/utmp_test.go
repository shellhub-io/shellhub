package utmp

import (
	"testing"
)

// TestUtmpConstants verifies the POSIX accounting constants remain correct.
// These values are specified by the utmpx(5) spec and must not change.
func TestUtmpConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{
			name:     "UserProcess type",
			got:      UserProcess,
			expected: 0x7,
		},
		{
			name:     "DeadProcess type",
			got:      DeadProcess,
			expected: 0x8,
		},
		{
			name:     "UtmpxFile path",
			got:      UtmpxFile,
			expected: "/var/run/utmp",
		},
		{
			name:     "WtmpxFile path",
			got:      WtmpxFile,
			expected: "/var/log/wtmp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.got != tt.expected {
				t.Errorf("got %v, want %v", tt.got, tt.expected)
			}
		})
	}
}
