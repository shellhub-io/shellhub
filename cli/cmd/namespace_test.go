package cmd

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
			cmd := NamespaceCommands(nil)
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
			cmd := NamespaceCommands(nil)
			cmd.SetArgs(append([]string{"delete"}, tc.args...))
			err := cmd.Execute()

			assert.Equal(t, tc.expectedErr, err != nil)
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
			cmd := NamespaceCommands(nil)
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
			cmd := NamespaceCommands(nil)
			cmd.SetArgs(append([]string{"member", "remove"}, tc.args...))
			err := cmd.Execute()

			assert.Equal(t, tc.expectedErr, err != nil)
		})
	}
}
