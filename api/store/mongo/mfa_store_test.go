package mongo

import (
	"context"
	"testing"

	"github.com/shellhub-io/mongotest"
	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/stretchr/testify/assert"
)

func TestDeleteCodes(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

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
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			err := mongostore.DeleteCodes(ctx, tc.username)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestAddStatusMFA(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

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
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			err := mongostore.AddStatusMFA(ctx, tc.username, tc.status)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestAddSecret(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

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
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			err := mongostore.AddSecret(ctx, tc.username, tc.secret)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestDeleteSecret(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

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
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			err := mongostore.DeleteSecret(ctx, tc.username)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}
