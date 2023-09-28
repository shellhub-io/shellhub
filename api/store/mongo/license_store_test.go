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

func TestLicenseLoad(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	type Expected struct {
		license *models.License
		err     error
	}

	cases := []struct {
		description string
		setup       func() error
		expected    Expected
	}{
		{
			description: "fails when license is not found",
			setup: func() error {
				return nil
			},
			expected: Expected{
				license: nil,
				err:     store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when license is found",
			setup: func() error {
				return mongotest.UseFixture(fixtures.License)
			},
			expected: Expected{
				license: &models.License{
					CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					RawData:   []byte("test"),
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			license, err := mongostore.LicenseLoad(ctx)
			assert.Equal(t, tc.expected, Expected{license: license, err: err})

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestLicenseSave(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	cases := []struct {
		description string
		license     *models.License
		setup       func() error
		expected    error
	}{
		{
			description: "succeeds when data is valid",
			license: &models.License{
				RawData:   []byte("test"),
				CreatedAt: time.Now(),
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

			err = mongostore.LicenseSave(ctx, tc.license)
			assert.Equal(t, tc.expected, err)
		})
	}
}
