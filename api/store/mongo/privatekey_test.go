package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
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

	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

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
			fingerprint: "",
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
	ctx := context.TODO()
	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	collection := mongostore.db.Collection("private_keys")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			var testData []interface{}
			if tc.fingerprint != "" {
				doc := bson.M{
					"fingerprint": tc.fingerprint,
					"data":        []byte("test"),
					"created_at":  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				}
				testData = append(testData, doc)

				if err := dbtest.InsertMockData(ctx, collection, testData); err != nil {
					t.Fatalf("failed to insert documents: %v", err)
				}
			}

			privKey, err := mongostore.PrivateKeyGet(context.TODO(), tc.fingerprint)
			assert.Equal(t, tc.expected, Expected{privKey: privKey, err: err})
		})
	}
}
