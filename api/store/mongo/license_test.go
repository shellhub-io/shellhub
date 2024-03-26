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

func TestLicenseLoad(t *testing.T) {
	type Expected struct {
		license *models.License
		err     error
	}

	cases := []struct {
		description string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when license is not found",
			fixtures:    []string{},
			expected: Expected{
				license: nil,
				err:     store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when license is found",
			fixtures:    []string{fixtures.FixtureLicenses},
			expected: Expected{
				license: &models.License{
					CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				},
				err: nil,
			},
		},
	}

	ctx := context.TODO()
	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	collection := mongostore.db.Collection("licenses")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown()

			testData := bson.M{
				"created_at": time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				"raw_data":   "test",
			}

			if len(tc.fixtures) > 0 {
				if err := dbtest.InsertMockData(ctx, collection, []interface{}{testData}); err != nil {
					t.Fatalf("failed to insert documents: %v", err)
				}
			}

			license, err := mongostore.LicenseLoad(ctx)
			assert.Equal(t, tc.expected.license, license)
			assert.Equal(t, tc.expected.err, err)
		})
	}
}

func TestLicenseSave(t *testing.T) {
	cases := []struct {
		description string
		license     *models.License
		fixtures    []string
		expected    error
	}{
		{
			description: "succeeds when data is valid",
			license: &models.License{
				RawData:   []byte("test"),
				CreatedAt: time.Now(),
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
			defer fixtures.Teardown()

			err := mongostore.LicenseSave(ctx, tc.license)
			assert.Equal(t, tc.expected, err)
		})
	}
}
