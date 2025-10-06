package yescrypt

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/openwall/yescrypt-go"
	"github.com/stretchr/testify/assert"
)

func FuzzVerify(f *testing.F) {
	const settings = "$y$j9T$AAt9R641xPvCI9nXw1HHW/"

	for i := 0; i < 100; i++ {
		v, err := rand.Int(rand.Reader, big.NewInt(64))
		assert.NoError(f, err)

		password := make([]byte, v.Int64())
		_, err = rand.Read(password)
		assert.NoError(f, err)

		hash, err := yescrypt.Hash(password, []byte(settings))
		assert.NoError(f, err)

		f.Add(string(password), string(hash))
	}

	f.Fuzz(func(_ *testing.T, a string, b string) {
		assert.True(f, Verify(a, b))
	})
}

func TestVeirfy(t *testing.T) {
	cases := []struct {
		description string
		password    string
		hash        string
		expected    bool
	}{
		{
			description: "invalid password",
			password:    "invalid",
			hash:        "$y$j9T$AAt9R641xPvCI9nXw1HHW/$nCv3bckjDEC9Q5ahIEpyXVNGZhySye/ZdjxNxTY5ttB",
			expected:    false,
		},
		{
			description: "invalid hash",
			password:    "password",
			hash:        "invalid",
			expected:    false,
		},
		{
			description: "valid",
			password:    "password",
			hash:        "$y$j9T$AAt9R641xPvCI9nXw1HHW/$nCv3bckjDEC9Q5ahIEpyXVNGZhySye/ZdjxNxTY5ttB",
			expected:    true,
		},
	}

	for _, test := range cases {
		t.Run(test.description, func(tt *testing.T) {
			result := Verify(test.password, test.hash)

			assert.Equal(tt, test.expected, result)
		})
	}
}
