package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/mongotest"
	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestPublicKeyGet(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	type Expected struct {
		pubKey *models.PublicKey
		err    error
	}

	cases := []struct {
		description string
		fingerprint string
		tenant      string
		setup       func() error
		expected    Expected
	}{
		{
			description: "succeeds when public key is not found due to fingerprint",
			fingerprint: "nonexistent",
			tenant:      "00000000-0000-4000-0000-000000000000",
			setup: func() error {
				return mongotest.UseFixture(fixtures.PublicKey)
			},
			expected: Expected{
				pubKey: nil,
				err:    store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when public key is not found due to tenant",
			fingerprint: "fingerprint",
			tenant:      "nonexistent",
			setup: func() error {
				return mongotest.UseFixture(fixtures.PublicKey)
			},
			expected: Expected{
				pubKey: nil,
				err:    store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when public key is found",
			fingerprint: "fingerprint",
			tenant:      "00000000-0000-4000-0000-000000000000",
			setup: func() error {
				return mongotest.UseFixture(fixtures.PublicKey)
			},
			expected: Expected{
				pubKey: &models.PublicKey{
					Data:        []byte("test"),
					CreatedAt:   time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					Fingerprint: "fingerprint",
					TenantID:    "00000000-0000-4000-0000-000000000000",
					PublicKeyFields: models.PublicKeyFields{
						Name: "public_key",
						Filter: models.PublicKeyFilter{
							Hostname: ".*",
							Tags:     []string{"tag1"},
						},
					},
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			pubKey, err := mongostore.PublicKeyGet(ctx, tc.fingerprint, tc.tenant)
			assert.Equal(t, tc.expected, Expected{pubKey: pubKey, err: err})

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestPublicKeyList(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	type Expected struct {
		pubKey []models.PublicKey
		len    int
		err    error
	}

	cases := []struct {
		description string
		setup       func() error
		expected    Expected
	}{
		{
			description: "succeeds when public key list is empty",
			setup: func() error {
				return nil
			},
			expected: Expected{
				pubKey: []models.PublicKey{},
				len:    0,
				err:    nil,
			},
		},
		{
			description: "succeeds when public key list len is greater than 1",
			setup: func() error {
				return mongotest.UseFixture(fixtures.PublicKey)
			},
			expected: Expected{
				pubKey: []models.PublicKey{
					{
						Data:        []byte("test"),
						CreatedAt:   time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						Fingerprint: "fingerprint",
						TenantID:    "00000000-0000-4000-0000-000000000000",
						PublicKeyFields: models.PublicKeyFields{
							Name: "public_key",
							Filter: models.PublicKeyFilter{
								Hostname: ".*",
								Tags:     []string{"tag1"},
							},
						},
					},
				},
				len: 1,
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			pubKey, count, err := mongostore.PublicKeyList(ctx, paginator.Query{Page: -1, PerPage: -1})
			assert.Equal(t, tc.expected, Expected{pubKey: pubKey, len: count, err: err})

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestPublicKeyCreate(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	cases := []struct {
		description string
		key         *models.PublicKey
		setup       func() error
		expected    error
	}{
		{
			description: "succeeds when data is valid",
			key: &models.PublicKey{
				Data:            []byte("test"),
				Fingerprint:     "fingerprint",
				TenantID:        "00000000-0000-4000-0000-000000000000",
				PublicKeyFields: models.PublicKeyFields{Name: "public_key", Filter: models.PublicKeyFilter{Hostname: ".*"}},
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

			err = mongostore.PublicKeyCreate(ctx, tc.key)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestPublicKeyUpdate(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	type Expected struct {
		pubKey *models.PublicKey
		err    error
	}

	cases := []struct {
		description string
		fingerprint string
		tenant      string
		key         *models.PublicKeyUpdate
		setup       func() error
		expected    Expected
	}{
		{
			description: "succeeds when public key is not found due to fingerprint",
			fingerprint: "nonexistent",
			tenant:      "00000000-0000-4000-0000-000000000000",
			key: &models.PublicKeyUpdate{
				PublicKeyFields: models.PublicKeyFields{
					Name:   "edited_name",
					Filter: models.PublicKeyFilter{Hostname: ".*"},
				},
			},
			setup: func() error {
				return mongotest.UseFixture(fixtures.PublicKey)
			},
			expected: Expected{
				pubKey: nil,
				err:    store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when public key is not found due to tenant",
			fingerprint: "fingerprint",
			tenant:      "nonexistent",
			key: &models.PublicKeyUpdate{
				PublicKeyFields: models.PublicKeyFields{
					Name:   "edited_name",
					Filter: models.PublicKeyFilter{Hostname: ".*"},
				},
			},
			setup: func() error {
				return mongotest.UseFixture(fixtures.PublicKey)
			},
			expected: Expected{
				pubKey: nil,
				err:    store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when public key is found",
			fingerprint: "fingerprint",
			tenant:      "00000000-0000-4000-0000-000000000000",
			key: &models.PublicKeyUpdate{
				PublicKeyFields: models.PublicKeyFields{
					Name: "edited_key",
					Filter: models.PublicKeyFilter{
						Hostname: ".*",
						Tags:     []string{"edited-tag1"},
					},
				},
			},
			setup: func() error {
				return mongotest.UseFixture(fixtures.PublicKey)
			},
			expected: Expected{
				pubKey: &models.PublicKey{
					Data:        []byte("test"),
					CreatedAt:   time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					Fingerprint: "fingerprint",
					TenantID:    "00000000-0000-4000-0000-000000000000",
					PublicKeyFields: models.PublicKeyFields{
						Name: "edited_key",
						Filter: models.PublicKeyFilter{
							Hostname: ".*",
							Tags:     []string{"edited-tag1"},
						},
					},
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			pubKey, err := mongostore.PublicKeyUpdate(ctx, tc.fingerprint, tc.tenant, tc.key)
			assert.Equal(t, tc.expected, Expected{pubKey: pubKey, err: err})

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestPublicKeyDelete(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description string
		fingerprint string
		tenant      string
		setup       func() error
		expected    error
	}{
		{
			description: "fails when public key is not found due to fingerprint",
			fingerprint: "nonexistent",
			tenant:      "00000000-0000-4000-0000-000000000000",
			setup: func() error {
				return mongotest.UseFixture(fixtures.PublicKey)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "fails when public key is not found due to tenant",
			fingerprint: "fingerprint",
			tenant:      "nonexistent",
			setup: func() error {
				return mongotest.UseFixture(fixtures.PublicKey)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when public key is found",
			fingerprint: "fingerprint",
			tenant:      "00000000-0000-4000-0000-000000000000",
			setup: func() error {
				return mongotest.UseFixture(fixtures.PublicKey)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.PublicKeyDelete(ctx, tc.fingerprint, tc.tenant)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}
