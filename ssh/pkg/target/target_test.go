package target

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTarget(t *testing.T) {
	type Expected struct {
		target *Target
		err    error
	}

	cases := []struct {
		description string
		sshid       string
		expected    Expected
	}{
		{
			description: "fails when could not split the target",
			sshid:       "username",
			expected: Expected{
				target: nil,
				err:    ErrSplitTarget,
			},
		},
		{
			description: "succeeds when target is valid",
			sshid:       "username@namespace.00-00-00-00-00-00@localhost",
			expected: Expected{
				target: &Target{
					Username: "username",
					Data:     "namespace.00-00-00-00-00-00@localhost",
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			target, err := NewTarget(tc.sshid)
			assert.Equal(t, tc.expected, Expected{target, err})
		})
	}
}

func TestIsSSHID(t *testing.T) {
	cases := []struct {
		description string
		target      *Target
		expected    bool
	}{
		{
			description: "returns false when Data does not contain a dot",
			target: &Target{
				Username: "username",
				Data:     "username@localhost",
			},
			expected: false,
		},
		{
			description: "returns true when Data contains a dot",
			target: &Target{
				Username: "username",
				Data:     "namespace.00-00-00-00-00-00@localhost",
			},
			expected: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			isSSHID := tc.target.IsSSHID()
			assert.Equal(t, tc.expected, isSSHID)
		})
	}
}

func TestSplitSSHID(t *testing.T) {
	type Expected struct {
		namespace string
		hostname  string
		err       error
	}

	cases := []struct {
		description string
		target      *Target
		expected    Expected
	}{
		{
			description: "failst when when Data does not contain a dot",
			target: &Target{
				Username: "username",
				Data:     "username@localhost",
			},
			expected: Expected{
				namespace: "",
				hostname:  "",
				err:       ErrNotSSHID,
			},
		},
		{
			description: "succeeds when Data contains a dot",
			target: &Target{
				Username: "username",
				Data:     "namespace.00-00-00-00-00-00@localhost",
			},
			expected: Expected{
				namespace: "namespace",
				hostname:  "00-00-00-00-00-00@localhost",
				err:       nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			namespace, hostname, err := tc.target.SplitSSHID()
			assert.Equal(t, tc.expected, Expected{namespace, hostname, err})
		})
	}
}
