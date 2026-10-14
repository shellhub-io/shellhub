package yescrypt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
