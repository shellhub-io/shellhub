package pairingcode

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	for _, length := range []int{DeviceCodeLength, InviteCodeLength} {
		seen := make(map[string]struct{})

		for range 2000 {
			code, err := New(length)
			require.NoError(t, err)

			assert.Len(t, code, length)
			assert.Equal(t, strings.ToUpper(code), code, "code must be uppercase")
			assert.NotContains(t, code, "-", "canonical code has no separator")

			for _, r := range code {
				assert.Truef(t, strings.ContainsRune(Alphabet, r), "unexpected char %q", r)
			}

			// Ambiguous characters must never appear.
			for _, bad := range []string{"0", "O", "1", "I", "L", "U"} {
				assert.NotContains(t, code, bad)
			}

			seen[code] = struct{}{}
		}

		// Collisions in 2000 draws are effectively impossible.
		assert.Len(t, seen, 2000, "expected all generated codes to be unique")
	}
}

func TestNormalize(t *testing.T) {
	cases := []struct {
		in       string
		expected string
	}{
		{"WXYZ2K7Q", "WXYZ2K7Q"},
		{"wxyz2k7q", "WXYZ2K7Q"},
		{"wxyz-2k7q", "WXYZ2K7Q"},
		{"  WXYZ 2K7Q ", "WXYZ2K7Q"},
		{"", ""},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.expected, Normalize(tc.in))
	}
}

func TestIsValid(t *testing.T) {
	assert.True(t, IsValid("WXYZ2K7Q", DeviceCodeLength))

	// Generated codes are always valid at their length.
	code, err := New(InviteCodeLength)
	require.NoError(t, err)
	assert.True(t, IsValid(code, InviteCodeLength))

	// Wrong length (including the old 32-hex format).
	assert.False(t, IsValid("", DeviceCodeLength))
	assert.False(t, IsValid("WXYZ2K7", DeviceCodeLength))
	assert.False(t, IsValid("WXYZ2K7QQ", DeviceCodeLength))
	assert.False(t, IsValid("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", DeviceCodeLength))

	// Ambiguous / out-of-alphabet characters.
	assert.False(t, IsValid("WXYZ2K7O", DeviceCodeLength), "O not in alphabet")
	assert.False(t, IsValid("WXYZ2K71", DeviceCodeLength), "1 not in alphabet")
	assert.False(t, IsValid("wxyz2k7q", DeviceCodeLength), "lowercase is not canonical")
	assert.False(t, IsValid("WXYZ-2K7", DeviceCodeLength), "hyphen is not canonical")
}
