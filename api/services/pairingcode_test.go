package services

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPairingCode(t *testing.T) {
	seen := make(map[string]struct{})

	for range 2000 {
		code, err := newPairingCode()
		require.NoError(t, err)

		assert.Len(t, code, pairingCodeLength)
		assert.Equal(t, strings.ToUpper(code), code, "code must be uppercase")
		assert.NotContains(t, code, "-", "canonical code has no separator")

		for _, r := range code {
			assert.Truef(t, strings.ContainsRune(pairingCodeAlphabet, r), "unexpected char %q", r)
		}

		// Ambiguous characters must never appear.
		for _, bad := range []string{"0", "O", "1", "I", "L", "U"} {
			assert.NotContains(t, code, bad)
		}

		seen[code] = struct{}{}
	}

	// Collisions in 2000 draws from ~2^39 are effectively impossible.
	assert.Len(t, seen, 2000, "expected all generated codes to be unique")
}

func TestNormalizePairingCode(t *testing.T) {
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
		assert.Equal(t, tc.expected, normalizePairingCode(tc.in))
	}
}

func TestIsValidPairingCode(t *testing.T) {
	assert.True(t, isValidPairingCode("WXYZ2K7Q"))

	// Generated codes are always valid.
	code, err := newPairingCode()
	require.NoError(t, err)
	assert.True(t, isValidPairingCode(code))

	// Wrong length (including the old 32-hex format).
	assert.False(t, isValidPairingCode(""))
	assert.False(t, isValidPairingCode("WXYZ2K7"))
	assert.False(t, isValidPairingCode("WXYZ2K7QQ"))
	assert.False(t, isValidPairingCode("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"))

	// Ambiguous / out-of-alphabet characters.
	assert.False(t, isValidPairingCode("WXYZ2K7O"), "O not in alphabet")
	assert.False(t, isValidPairingCode("WXYZ2K71"), "1 not in alphabet")
	assert.False(t, isValidPairingCode("wxyz2k7q"), "lowercase is not canonical")
	assert.False(t, isValidPairingCode("WXYZ-2K7"), "hyphen is not canonical")
}
