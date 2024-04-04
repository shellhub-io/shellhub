package mongo_test

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	shstore "github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestAPIKeyCreate(t *testing.T) {
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
			ctx := context.Background()

			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, fixtures.Teardown())
			})

			err := store.APIKeyCreate(ctx, tc.APIKey)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestAPIKeyList(t *testing.T) {
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
			ctx := context.Background()

			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, fixtures.Teardown())
			})

			_, _, err := store.APIKeyList(ctx, tc.requestParams.UserID, tc.requestParams.Paginator, tc.requestParams.Sorter, "tenant")
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestDeleteAPIKey(t *testing.T) {
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
			expected:    shstore.ErrNoDocuments,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, fixtures.Teardown())
			})

			err := store.APIKeyDelete(ctx, tc.id, "tenant")
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestRenameAPIKey(t *testing.T) {
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
			expected: shstore.ErrNoDocuments,
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
			ctx := context.Background()

			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, fixtures.Teardown())
			})

			err := store.APIKeyEdit(ctx, tc.requestParams)
			assert.Equal(t, tc.expected, err)
		})
	}
}
