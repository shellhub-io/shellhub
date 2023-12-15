package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestPrivateKeyCreate(t *testing.T) {
	cases := []struct {
		description string
		priKey      *models.PrivateKey
		fixtures    []string
		expected    error
	}{
		{
			description: "succeeds when data is valid",
			priKey: &models.PrivateKey{
				Data:        []byte("test"),
				Fingerprint: "fingerprint",
				CreatedAt:   time.Now(),
			},
			fixtures: []string{},
			expected: nil,
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			err := mongostore.PrivateKeyCreate(context.TODO(), tc.priKey)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestPrivateKeyGet(t *testing.T) {
	type Expected struct {
		privKey *models.PrivateKey
		err     error
	}

	cases := []struct {
		description string
		fingerprint string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when private key is not found",
			fingerprint: "nonexistent",
			fixtures:    []string{fixtures.FixturePrivateKeys},
			expected: Expected{
				privKey: nil,
				err:     store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when private key is found",
			fingerprint: "fingerprint",
			fixtures:    []string{fixtures.FixturePrivateKeys},
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

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			privKey, err := mongostore.PrivateKeyGet(context.TODO(), tc.fingerprint)
			assert.Equal(t, tc.expected, Expected{privKey: privKey, err: err})
		})
	}
}
