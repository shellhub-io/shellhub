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

func TestLicenseLoad(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

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
			fixtures:    []string{fixtures.License},
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
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			license, err := mongostore.LicenseLoad(ctx)
			assert.Equal(t, tc.expected, Expected{license: license, err: err})
		})
	}
}

func TestLicenseSave(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

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

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			err := mongostore.LicenseSave(ctx, tc.license)
			assert.Equal(t, tc.expected, err)
		})
	}
}
