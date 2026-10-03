package hash

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			hash, err := Do(tc.password)
			require.NoError(t, err)
			assert.NotEmpty(t, hash)
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
			assert.Equal(t, tc.expected, CompareWith(tc.password, tc.hash))
		})
	}
}

type fixedHasher string

func (f fixedHasher) Do(string) (string, error) {
	return string(f), nil
}

func (f fixedHasher) CompareWith(_ string, hash string) bool {
	return hash == string(f)
}

func TestSet(t *testing.T) {
	Set(t, fixedHasher("outer"))

	t.Run("serves Do from the hasher until the test ends", func(t *testing.T) {
		Set(t, fixedHasher("inner"))

		hashed, err := Do("secret")
		require.NoError(t, err)
		assert.Equal(t, "inner", hashed)
	})

	hashed, err := Do("secret")
	require.NoError(t, err)
	assert.Equal(t, "outer", hashed)
}
