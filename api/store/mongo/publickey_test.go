package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func TestPublicKeyGet(t *testing.T) {
	type Expected struct {
		pubKey *models.PublicKey
		err    error
	}

	cases := []struct {
		description string
		fingerprint string
		tenant      string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "succeeds when public key is not found due to fingerprint",
			fingerprint: "",
			tenant:      "00000000-0000-4000-0000-000000000000",
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected: Expected{
				pubKey: nil,
				err:    store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when public key is not found due to tenant",
			fingerprint: "fingerprint",
			tenant:      "",
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected: Expected{
				pubKey: nil,
				err:    store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when public key is found",
			fingerprint: "fingerprint",
			tenant:      "00000000-0000-4000-0000-000000000000",
			fixtures:    []string{fixtures.FixturePublicKeys},
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
							Tags:     []string{"tag-1"},
						},
					},
				},
				err: nil,
			},
		},
	}

	ctx := context.TODO()
	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	collection := mongostore.db.Collection("public_keys")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			var testData []interface{}
			if tc.fingerprint != "" && tc.tenant != "" {
				doc := bson.M{
					"fingerprint": tc.fingerprint,
					"tenant_id":   tc.tenant,
					"data":        []byte("test"),
					"created_at":  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					"name":        "public_key",
					"filter": bson.M{
						"hostname": ".*",
						"tags":     []string{"tag-1"},
					},
				}
				testData = append(testData, doc)

				if err := dbtest.InsertMockData(ctx, collection, testData); err != nil {
					t.Fatalf("failed to insert documents: %v", err)
				}
			}

			pubKey, err := mongostore.PublicKeyGet(context.TODO(), tc.fingerprint, tc.tenant)
			assert.Equal(t, tc.expected, Expected{pubKey: pubKey, err: err})
		})
	}
}

func TestPublicKeyList(t *testing.T) {
	type Expected struct {
		pubKey []models.PublicKey
		len    int
		err    error
	}

	cases := []struct {
		description string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "succeeds when public key list is empty",
			fixtures:    []string{},
			expected: Expected{
				pubKey: []models.PublicKey{},
				len:    0,
				err:    nil,
			},
		},
		{
			description: "succeeds when public key list len is greater than 1",
			fixtures:    []string{fixtures.FixturePublicKeys},
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
								Tags:     []string{"tag-1"},
							},
						},
					},
				},
				len: 1,
				err: nil,
			},
		},
	}

	ctx := context.TODO()
	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	collection := mongostore.db.Collection("public_keys")
	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			var testData []interface{}
			if tc.expected.len > 0 {
				doc := bson.M{
					"fingerprint": "fingerprint",
					"tenant_id":   "00000000-0000-4000-0000-000000000000",
					"data":        []byte("test"),
					"created_at":  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					"name":        "public_key",
					"filter": bson.M{
						"hostname": ".*",
						"tags":     []string{"tag-1"},
					},
				}

				testData = append(testData, doc)

				if err := dbtest.InsertMockData(ctx, collection, testData); err != nil {
					t.Fatalf("failed to insert documents: %v", err)
				}
			}

			pubKey, count, err := mongostore.PublicKeyList(ctx, query.Paginator{Page: -1, PerPage: -1})
			assert.Equal(t, tc.expected, Expected{pubKey: pubKey, len: count, err: err})
		})
	}
}

func TestPublicKeyCreate(t *testing.T) {
	cases := []struct {
		description string
		key         *models.PublicKey
		fixtures    []string
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
			fixtures: []string{},
			expected: nil,
		},
	}

	ctx := context.TODO()
	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			err := mongostore.PublicKeyCreate(ctx, tc.key)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestPublicKeyUpdate(t *testing.T) {
	type Expected struct {
		pubKey *models.PublicKey
		err    error
	}
	cases := []struct {
		description string
		fingerprint string
		tenant      string
		key         *models.PublicKeyUpdate
		fixtures    []string
		expected    Expected
	}{
		{
			description: "succeeds when public key is not found due to fingerprint",
			fingerprint: "",
			tenant:      "00000000-0000-4000-0000-000000000000",
			key: &models.PublicKeyUpdate{
				PublicKeyFields: models.PublicKeyFields{
					Name:   "edited_name",
					Filter: models.PublicKeyFilter{Hostname: ".*"},
				},
			},
			fixtures: []string{fixtures.FixturePublicKeys},
			expected: Expected{
				pubKey: nil,
				err:    store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when public key is not found due to tenant",
			fingerprint: "fingerprint",
			tenant:      "",
			key: &models.PublicKeyUpdate{
				PublicKeyFields: models.PublicKeyFields{
					Name:   "edited_name",
					Filter: models.PublicKeyFilter{Hostname: ".*"},
				},
			},
			fixtures: []string{fixtures.FixturePublicKeys},
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
						Tags:     []string{"edited-tag"},
					},
				},
			},
			fixtures: []string{fixtures.FixturePublicKeys},
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
							Tags:     []string{"edited-tag"},
						},
					},
				},
				err: nil,
			},
		},
	}

	ctx := context.TODO()
	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	collection := mongostore.db.Collection("public_keys")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			var testData []interface{}
			if tc.fingerprint != "" && tc.tenant != "" {
				doc := bson.M{
					"fingerprint": tc.fingerprint,
					"tenant_id":   tc.tenant,
					"data":        []byte("test"),
					"created_at":  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					"name":        "public_key",
					"filter": bson.M{
						"hostname": ".*",
						"tags":     []string{"tag-1"},
					},
				}
				testData = append(testData, doc)

				if err := dbtest.InsertMockData(ctx, collection, testData); err != nil {
					t.Fatalf("failed to insert documents: %v", err)
				}
			}

			pubKey, err := mongostore.PublicKeyUpdate(ctx, tc.fingerprint, tc.tenant, tc.key)
			assert.Equal(t, tc.expected, Expected{pubKey: pubKey, err: err})
		})
	}
}

func TestPublicKeyDelete(t *testing.T) {
	cases := []struct {
		description string
		fingerprint string
		tenant      string
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when public key is not found due to fingerprint",
			fingerprint: "",
			tenant:      "00000000-0000-4000-0000-000000000000",
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "fails when public key is not found due to tenant",
			fingerprint: "fingerprint",
			tenant:      "",
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when public key is found",
			fingerprint: "fingerprint",
			tenant:      "00000000-0000-4000-0000-000000000000",
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected:    nil,
		},
	}

	ctx := context.TODO()
	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	collection := mongostore.db.Collection("public_keys")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			var testData []interface{}
			if tc.fingerprint != "" && tc.tenant != "" {
				doc := bson.M{
					"fingerprint": tc.fingerprint,
					"tenant_id":   tc.tenant,
					"data":        []byte("test"),
					"created_at":  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					"name":        "public_key",
					"filter": bson.M{
						"hostname": ".*",
						"tags":     []string{"tag-1"},
					},
				}
				testData = append(testData, doc)

				if err := dbtest.InsertMockData(ctx, collection, testData); err != nil {
					t.Fatalf("failed to insert documents: %v", err)
				}
			}

			err := mongostore.PublicKeyDelete(ctx, tc.fingerprint, tc.tenant)
			assert.Equal(t, tc.expected, err)
		})
	}
}
