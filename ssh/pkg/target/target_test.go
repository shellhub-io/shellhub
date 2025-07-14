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
				err:    ErrSplitTwoTarget,
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
		SSHID *SSHID
		err   error
	}

	cases := []struct {
		description string
		target      *Target
		expected    Expected
	}{
		{
			description: "fails when when Data does not contain a dot",
			target: &Target{
				Username: "username",
				Data:     "",
			},
			expected: Expected{
				SSHID: nil,
				err:   ErrNotSSHID,
			},
		},
		{
			description: "succeeds when Data contains a dot",
			target: &Target{
				Username: "username",
				Data:     "namespace.00-00-00-00-00-00",
			},
			expected: Expected{
				SSHID: &SSHID{
					Username:  "username",
					Namespace: "namespace",
					Device:    "00-00-00-00-00-00",
					Container: "",
				},
				err: nil,
			},
		},
		{
			description: "succeeds when Data contains two dots due to container",
			target: &Target{
				Username: "username",
				Data:     "namespace.00-00-00-00-00-00.container",
			},
			expected: Expected{
				SSHID: &SSHID{
					Username:  "username",
					Namespace: "namespace",
					Device:    "00-00-00-00-00-00",
					Container: "container",
				},
				err: nil,
			},
		},
		{
			description: "succeeds when Data contains more than two dots due to extra data",
			target: &Target{
				Username: "username",
				Data:     "namespace.00-00-00-00-00-00.container.extra",
			},
			expected: Expected{
				SSHID: &SSHID{
					Username:  "username",
					Namespace: "namespace",
					Device:    "00-00-00-00-00-00",
					Container: "container.extra",
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			sshid, err := tc.target.SplitSSHID()
			assert.Equal(t, tc.expected, Expected{sshid, err})
		})
	}
}
