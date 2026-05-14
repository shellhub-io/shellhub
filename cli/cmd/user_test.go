package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserCreateCmd(t *testing.T) {
	cases := []struct {
		description string
		args        []string
		expectedErr bool
	}{
		{
			description: "fails when email is invalid",
			args:        []string{"john_doe", "secret", "invalidmail.com"},
			expectedErr: true,
		},
		{
			description: "fails when username is invalid",
			args:        []string{"", "secret", "john.doe@test.com"},
			expectedErr: true,
		},
		{
			description: "fails when password is invalid",
			args:        []string{"john_doe", "ab", "john.doe@test.com"},
			expectedErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			cmd := UserCommands(nil)
			cmd.SetArgs(append([]string{"create"}, tc.args...))
			err := cmd.Execute()

			assert.Equal(t, tc.expectedErr, err != nil)
		})
	}
}

func TestUserResetPasswordCmd(t *testing.T) {
	cases := []struct {
		description string
		args        []string
		expectedErr bool
	}{
		{
			description: "fails when username is invalid",
			args:        []string{"", "secret"},
			expectedErr: true,
		},
		{
			description: "fails when password is invalid",
			args:        []string{"john_doe", "ab"},
			expectedErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			cmd := UserCommands(nil)
			cmd.SetArgs(append([]string{"password"}, tc.args...))
			err := cmd.Execute()

			assert.Equal(t, tc.expectedErr, err != nil)
		})
	}
}

func TestUserDeleteCmd(t *testing.T) {
	cases := []struct {
		description string
		args        []string
		expectedErr bool
	}{
		{
			description: "fails when username is invalid",
			args:        []string{""},
			expectedErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			cmd := UserCommands(nil)
			cmd.SetArgs(append([]string{"delete"}, tc.args...))
			err := cmd.Execute()

			assert.Equal(t, tc.expectedErr, err != nil)
		})
	}
}
