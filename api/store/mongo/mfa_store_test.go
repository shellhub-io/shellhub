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
	fixtures.Configure(&db)

	cases := []struct {
		description string
		username    string
		setup       func() error
		expected    error
	}{
		{
			description: "success when try to delete codes",
			username:    "username",
			setup: func() error {
				return mongotest.UseFixture(fixtures.User)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.DeleteCodes(ctx, tc.username)
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
	fixtures.Configure(&db)

	cases := []struct {
		description string
		username    string
		status      bool
		setup       func() error
		expected    error
	}{
		{
			description: "success when try to add status MFA",
			username:    "username",
			status:      true,
			setup: func() error {
				return mongotest.UseFixture(fixtures.User)
			},
			expected: nil,
		},
		// {
		// 	description: "fails when public key is not found due to tenant",
		// 	fingerprint: "fingerprint",
		// 	tenant:      "nonexistent",
		// 	tag:         "tag0",
		// 	setup: func() error {
		// 		return mongotest.UseFixture(fixtures.PublicKey)
		// 	},
		// 	expected: store.ErrNoDocuments,
		// },
		// {
		// 	description: "succeeds when public key is found",
		// 	fingerprint: "fingerprint",
		// 	tenant:      "00000000-0000-4000-0000-000000000000",
		// 	tag:         "tag0",
		// 	setup: func() error {
		// 		return mongotest.UseFixture(fixtures.PublicKey)
		// 	},
		// 	expected: nil,
		// },
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.AddStatusMFA(ctx, tc.username, tc.status)
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
	fixtures.Configure(&db)

	cases := []struct {
		description string
		username    string
		secret      string
		setup       func() error
		expected    error
	}{
		{
			description: "success when try to add status MFA",
			username:    "username",
			secret:      "IOJDSFIAWMKXskdlmawOSDMCALWC",
			setup: func() error {
				return mongotest.UseFixture(fixtures.User)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.AddSecret(ctx, tc.username, tc.secret)
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
	fixtures.Configure(&db)

	cases := []struct {
		description string
		username    string
		setup       func() error
		expected    error
	}{
		{
			description: "success to delete a status MFA",
			username:    "username",
			setup: func() error {
				return mongotest.UseFixture(fixtures.User)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.DeleteSecret(ctx, tc.username)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}
