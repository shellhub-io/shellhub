package mongo

import (
	"context"
	"testing"

	"github.com/shellhub-io/mongotest"
	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestAPIKeyCreate(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	cases := []struct {
		description string
		APIKey      *models.APIKey
		fixtures    []string
		expected    error
	}{
		{
			description: "success when try create a APIKey",
			APIKey: &models.APIKey{
				UserID: "id",
				Name:   "APIKeyName",
			},
			fixtures: []string{fixtures.FixtureUsers},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			err := mongostore.APIKeyCreate(ctx, tc.APIKey)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestAPIKeyList(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	cases := []struct {
		description   string
		requestParams *requests.APIKeyList
		fixtures      []string
		expected      error
	}{
		{
			description: "failure when  ID is invalid",
			requestParams: &requests.APIKeyList{
				TenantParam: requests.TenantParam{Tenant: "00000000-0000-4000-0000-000000000000"},
				Paginator:   query.Paginator{Page: 1, PerPage: 10},
				Sorter:      query.Sorter{By: "expires_in", Order: query.OrderAsc},
			},
			fixtures: []string{fixtures.FixtureUsers},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			_, _, err := mongostore.APIKeyList(ctx, tc.requestParams.UserID, tc.requestParams.Paginator, tc.requestParams.Sorter, "tenant")
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestDeleteAPIKey(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	cases := []struct {
		description string
		id          string
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when try delete with a invalid id",
			fixtures:    []string{fixtures.FixtureUsers},
			id:          "507f1f77bcf86cd7994390bb",
			expected:    store.ErrNoDocuments,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			err := mongostore.APIKeyDelete(ctx, tc.id, "tenant")
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestRenameAPIKey(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	cases := []struct {
		description   string
		requestParams *requests.APIKeyChanges
		fixtures      []string
		expected      error
	}{
		{
			description: "fails when try rename with invalid dates",
			requestParams: &requests.APIKeyChanges{
				ID:   "507f1f77bcf86cd7994390bb",
				Name: "invalid",
			},
			fixtures: []string{fixtures.FixtureUsers},
			expected: store.ErrNoDocuments,
		},
		{
			description: "success",
			requestParams: &requests.APIKeyChanges{
				ID: "507f1f77bcf86cd7994390bb",
			},
			fixtures: []string{fixtures.FixtureUsers},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			err := mongostore.APIKeyEdit(ctx, tc.requestParams)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}
