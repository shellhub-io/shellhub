package mongo_test

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/stretchr/testify/assert"
)

func TestDeleteCodes(t *testing.T) {
	cases := []struct {
		description string
		username    string
		fixtures    []string
		expected    error
	}{
		{
			description: "success when try to delete codes",
			username:    "username",
			fixtures:    []string{fixtures.FixtureUsers},
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, fixtures.Teardown())
			})

			err := s.DeleteCodes(ctx, tc.username)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestAddStatusMFA(t *testing.T) {
	cases := []struct {
		description string
		username    string
		status      bool
		fixtures    []string
		expected    error
	}{
		{
			description: "success when try to add status MFA",
			username:    "username",
			status:      true,
			fixtures:    []string{fixtures.FixtureUsers},
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, fixtures.Teardown())
			})

			err := s.AddStatusMFA(ctx, tc.username, tc.status)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestAddSecret(t *testing.T) {
	cases := []struct {
		description string
		username    string
		secret      string
		fixtures    []string
		expected    error
	}{
		{
			description: "success when try to add status MFA",
			username:    "username",
			secret:      "IOJDSFIAWMKXskdlmawOSDMCALWC",
			fixtures:    []string{fixtures.FixtureUsers},
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, fixtures.Teardown())
			})

			err := s.AddSecret(ctx, tc.username, tc.secret)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestDeleteSecret(t *testing.T) {
	cases := []struct {
		description string
		username    string
		fixtures    []string
		expected    error
	}{
		{
			description: "success to delete a status MFA",
			username:    "username",
			fixtures:    []string{fixtures.FixtureUsers},
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, fixtures.Teardown())
			})

			err := s.DeleteSecret(ctx, tc.username)
			assert.Equal(t, tc.expected, err)
		})
	}
}
