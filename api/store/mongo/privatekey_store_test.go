package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/mongotest"
	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestPrivateKeyCreate(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	cases := []struct {
		description string
		priKey      *models.PrivateKey
		setup       func() error
		expected    error
	}{
		{
			description: "succeeds when data is valid",
			priKey: &models.PrivateKey{
				Data:        []byte("test"),
				Fingerprint: "fingerprint",
				CreatedAt:   time.Now(),
			},
			setup: func() error {
				return nil
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.PrivateKeyCreate(ctx, tc.priKey)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestPrivateKeyGet(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	type Expected struct {
		privKey *models.PrivateKey
		err     error
	}

	cases := []struct {
		description string
		fingerprint string
		setup       func() error
		expected    Expected
	}{
		{
			description: "fails when private key is not found",
			fingerprint: "nonexistent",
			setup: func() error {
				return mongotest.UseFixture(fixtures.PrivateKey)
			},
			expected: Expected{
				privKey: nil,
				err:     store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when private key is found",
			fingerprint: "fingerprint",
			setup: func() error {
				return mongotest.UseFixture(fixtures.PrivateKey)
			},
			expected: Expected{
				privKey: &models.PrivateKey{
					Data:        []byte("test"),
					Fingerprint: "fingerprint",
					CreatedAt:   time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			privKey, err := mongostore.PrivateKeyGet(ctx, tc.fingerprint)
			assert.Equal(t, tc.expected, Expected{privKey: privKey, err: err})

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}
