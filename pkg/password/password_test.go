package password

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	cases := []struct {
		description string
		password    string
	}{
		{
			description: "succeeds when create a hash",
			password:    "secret",
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			hash, err := Hash(tc.password)
			assert.NoError(t, err)
			assert.NotEqual(t, hash, "")
		})
	}
}

func TestCompare(t *testing.T) {
	cases := []struct {
		description string
		password    string
		hash        string
		expected    bool
	}{
		{
			description: "should fail when the password is incorrect and hashed using SHA256",
			password:    "invalid",
			hash:        "2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b",
			expected:    false,
		},
		{
			description: "should succeed when the password is correct and hashed using SHA256",
			password:    "secret",
			hash:        "2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b",
			expected:    true,
		},
		{
			description: "should fail when the password is incorrect and hashed using bcrypt",
			password:    "invalid",
			hash:        "$2a$14$QPfofG/FHXFaRMiMjIgo8uHgJSj/zghR9abxEO6JmBu/rViSDNo.K",
			expected:    false,
		},
		{
			description: "should succeed when the password is correct and hashed using bcrypt",
			password:    "secret",
			hash:        "$2a$14$QPfofG/FHXFaRMiMjIgo8uHgJSj/zghR9abxEO6JmBu/rViSDNo.K",
			expected:    true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.Equal(t, tc.expected, Compare(tc.password, tc.hash))
		})
	}
}
