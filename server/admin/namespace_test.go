package admin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNamespaceCreateCmd(t *testing.T) {
	cases := []struct {
		description string
		args        []string
		expectedErr bool
	}{
		{
			description: "fails when namespace is empty",
			args:        []string{"", "john_doe"},
			expectedErr: true,
		},
		{
			description: "fails when namespace has invalid characters",
			args:        []string{"invalid_namespace", "john_doe"},
			expectedErr: true,
		},
		{
			description: "fails when owner username is invalid",
			args:        []string{"namespace", ""},
			expectedErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			cmd := namespaceCommands(nil)
			cmd.SetArgs(append([]string{"create"}, tc.args...))
			err := cmd.Execute()

			assert.Equal(t, tc.expectedErr, err != nil)
		})
	}
}

func TestNamespaceDeleteCmd(t *testing.T) {
	cases := []struct {
		description string
		args        []string
		expectedErr bool
	}{
		{
			description: "fails when namespace is empty",
			args:        []string{""},
			expectedErr: true,
		},
		{
			description: "fails when namespace has invalid characters",
			args:        []string{"invalid_namespace"},
			expectedErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			cmd := namespaceCommands(nil)
			cmd.SetArgs(append([]string{"delete"}, tc.args...))
			err := cmd.Execute()

			assert.Equal(t, tc.expectedErr, err != nil)
		})
	}
}

func TestNamespaceLimitCmd(t *testing.T) {
	cases := []struct {
		description string
		args        []string
		expected    error
	}{
		{
			description: "fails when namespace has invalid characters",
			args:        []string{"invalid_namespace", "10"},
			expected:    ErrInvalidNamespace,
		},
		{
			description: "fails when max devices is not a number",
			args:        []string{"namespace", "ten"},
			expected:    ErrInvalidMaxDevices,
		},
		{
			description: "fails when max devices is zero",
			args:        []string{"namespace", "0"},
			expected:    ErrInvalidMaxDevices,
		},
		{
			description: "fails when max devices is below -1",
			args:        []string{"namespace", "-2"},
			expected:    ErrInvalidMaxDevices,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			cmd := namespaceCommands(nil)
			cmd.SetArgs(append([]string{"limit"}, tc.args...))
			err := cmd.Execute()

			assert.ErrorIs(t, err, tc.expected)
		})
	}
}

func TestMemberAddCmd(t *testing.T) {
	cases := []struct {
		description string
		args        []string
		expectedErr bool
	}{
		{
			description: "fails when username is invalid",
			args:        []string{"", "namespace", "observer"},
			expectedErr: true,
		},
		{
			description: "fails when role is invalid",
			args:        []string{"john_doe", "namespace", "invalidrole"},
			expectedErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			cmd := namespaceCommands(nil)
			cmd.SetArgs(append([]string{"member", "add"}, tc.args...))
			err := cmd.Execute()

			assert.Equal(t, tc.expectedErr, err != nil)
		})
	}
}

func TestMemberRemoveCmd(t *testing.T) {
	cases := []struct {
		description string
		args        []string
		expectedErr bool
	}{
		{
			description: "fails when username is invalid",
			args:        []string{"", "namespace"},
			expectedErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			cmd := namespaceCommands(nil)
			cmd.SetArgs(append([]string{"member", "remove"}, tc.args...))
			err := cmd.Execute()

			assert.Equal(t, tc.expectedErr, err != nil)
		})
	}
}
