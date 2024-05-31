package mongo_test

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

func TestAPIKeyCreate(t *testing.T) {
	type Expected struct {
		insertedID string
		err        error
	}

	cases := []struct {
		description string
		apiKey      *models.APIKey
		expected    Expected
	}{
		{
			description: "succeeds",
			apiKey: &models.APIKey{
				ID:        "f23a2e56cd3fcfba002c72675c870e1e7813292adc40bbf14cea479a2e07976a",
				Name:      "dev",
				CreatedBy: "507f1f77bcf86cd799439011",
				TenantID:  "00000000-0000-4000-0000-000000000000",
				Role:      "admin",
				CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				ExpiresIn: 0,
			},
			expected: Expected{
				insertedID: "f23a2e56cd3fcfba002c72675c870e1e7813292adc40bbf14cea479a2e07976a",
				err:        nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			insertedID, err := s.APIKeyCreate(ctx, tc.apiKey)
			require.Equal(t, tc.expected, Expected{insertedID, err})
		})
	}
}

func TestAPIKeyGet(t *testing.T) {
	type Expected struct {
		apiKey *models.APIKey
		err    error
	}

	cases := []struct {
		description string
		id          string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when name and tenant id does not exists",
			id:          "nonexistent",
			fixtures:    []string{fixtureAPIKeys},
			expected: Expected{
				apiKey: nil,
				err:    store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds",
			id:          "f23a2e56cd3fcfba002c72675c870e1e7813292adc40bbf14cea479a2e07976a",
			fixtures:    []string{fixtureAPIKeys},
			expected: Expected{
				apiKey: &models.APIKey{
					ID:        "f23a2e56cd3fcfba002c72675c870e1e7813292adc40bbf14cea479a2e07976a",
					Name:      "dev",
					CreatedBy: "507f1f77bcf86cd799439011",
					TenantID:  "00000000-0000-4000-0000-000000000000",
					Role:      "admin",
					CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					ExpiresIn: 0,
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			require.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() { require.NoError(t, srv.Reset()) })

			apiKey, err := s.APIKeyGet(ctx, tc.id)
			require.Equal(t, tc.expected, Expected{apiKey, err})
		})
	}
}

func TestAPIKeyGetByName(t *testing.T) {
	type Expected struct {
		apiKey *models.APIKey
		err    error
	}

	cases := []struct {
		description string
		name        string
		tenantID    string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when name and tenant id does not exists",
			tenantID:    "nonexistent",
			name:        "nonexistent",
			fixtures:    []string{fixtureAPIKeys},
			expected: Expected{
				apiKey: nil,
				err:    store.ErrNoDocuments,
			},
		},
		{
			description: "fails when name is valid but tenant id not",
			tenantID:    "nonexistent",
			name:        "dev",
			fixtures:    []string{fixtureAPIKeys},
			expected: Expected{
				apiKey: nil,
				err:    store.ErrNoDocuments,
			},
		},
		{
			description: "fails when tenant id is valid but name not",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			name:        "nonexistent",
			fixtures:    []string{fixtureAPIKeys},
			expected: Expected{
				apiKey: nil,
				err:    store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			name:        "dev",
			fixtures:    []string{fixtureAPIKeys},
			expected: Expected{
				apiKey: &models.APIKey{
					ID:        "f23a2e56cd3fcfba002c72675c870e1e7813292adc40bbf14cea479a2e07976a",
					Name:      "dev",
					CreatedBy: "507f1f77bcf86cd799439011",
					TenantID:  "00000000-0000-4000-0000-000000000000",
					Role:      "admin",
					CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					ExpiresIn: 0,
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			require.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() { require.NoError(t, srv.Reset()) })

			apiKey, err := s.APIKeyGetByName(ctx, tc.tenantID, tc.name)
			require.Equal(t, tc.expected, Expected{apiKey, err})
		})
	}
}

func TestAPIKeyConflicts(t *testing.T) {
	type Expected struct {
		conflicts []string
		ok        bool
		err       error
	}

	cases := []struct {
		description string
		tenantID    string
		target      *models.APIKeyConflicts
		fixtures    []string
		expected    Expected
	}{
		{
			description: "no conflicts when target is empty",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			target:      &models.APIKeyConflicts{},
			fixtures:    []string{fixtureAPIKeys},
			expected:    Expected{[]string{}, false, nil},
		},
		{
			description: "no conflicts with non existing name",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			target:      &models.APIKeyConflicts{Name: "nonexistent"},
			fixtures:    []string{fixtureAPIKeys},
			expected:    Expected{[]string{}, false, nil},
		},
		{
			description: "no conflict detected with existing attribute but different tenant id",
			tenantID:    "nonexistent",
			target:      &models.APIKeyConflicts{Name: "dev"},
			fixtures:    []string{fixtureAPIKeys},
			expected:    Expected{[]string{}, false, nil},
		},
		{
			description: "conflict detected with existing name",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			target:      &models.APIKeyConflicts{Name: "dev"},
			fixtures:    []string{fixtureAPIKeys},
			expected:    Expected{[]string{"name"}, true, nil},
		},
		{
			description: "conflict detected with existing id",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			target:      &models.APIKeyConflicts{ID: "f23a2e56cd3fcfba002c72675c870e1e7813292adc40bbf14cea479a2e07976a"},
			fixtures:    []string{fixtureAPIKeys},
			expected:    Expected{[]string{"id"}, true, nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			require.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() { require.NoError(t, srv.Reset()) })

			conflicts, ok, err := s.APIKeyConflicts(ctx, tc.tenantID, tc.target)
			require.Equal(t, tc.expected, Expected{conflicts, ok, err})
		})
	}
}

func TestAPIKeyList(t *testing.T) {
	type Expected struct {
		apiKeys []models.APIKey
		count   int
		err     error
	}

	cases := []struct {
		description string
		tenantID    string
		paginator   query.Paginator
		sorter      query.Sorter
		fixtures    []string
		expected    Expected
	}{
		{
			description: "succeeds when there are no api keys",
			tenantID:    "nonexistent",
			paginator:   query.Paginator{Page: 1, PerPage: 10},
			sorter:      query.Sorter{By: "expires_in", Order: query.OrderAsc},
			fixtures:    []string{fixtureAPIKeys},
			expected: Expected{
				apiKeys: []models.APIKey{},
				count:   0,
				err:     nil,
			},
		},
		{
			description: "succeeds when there are api keys",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			paginator:   query.Paginator{Page: 1, PerPage: 10},
			sorter:      query.Sorter{By: "expires_in", Order: query.OrderAsc},
			fixtures:    []string{fixtureAPIKeys},
			expected: Expected{
				apiKeys: []models.APIKey{
					{
						ID:        "f23a2e56cd3fcfba002c72675c870e1e7813292adc40bbf14cea479a2e07976a",
						Name:      "dev",
						CreatedBy: "507f1f77bcf86cd799439011",
						TenantID:  "00000000-0000-4000-0000-000000000000",
						Role:      "admin",
						CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						ExpiresIn: 0,
					},
					{
						ID:        "a1b2c73ea41f70870c035283336d72228118213ed03ec78043ffee48d827af11",
						Name:      "prod",
						CreatedBy: "507f1f77bcf86cd799439011",
						TenantID:  "00000000-0000-4000-0000-000000000000",
						Role:      "operator",
						CreatedAt: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						ExpiresIn: 10,
					},
				},
				count: 2,
				err:   nil,
			},
		},
		{
			description: "succeeds when there are api keys and pagination",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			paginator:   query.Paginator{Page: 1, PerPage: 1},
			sorter:      query.Sorter{By: "expires_in", Order: query.OrderAsc},
			fixtures:    []string{fixtureAPIKeys},
			expected: Expected{
				apiKeys: []models.APIKey{
					{
						ID:        "f23a2e56cd3fcfba002c72675c870e1e7813292adc40bbf14cea479a2e07976a",
						Name:      "dev",
						CreatedBy: "507f1f77bcf86cd799439011",
						TenantID:  "00000000-0000-4000-0000-000000000000",
						Role:      "admin",
						CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						ExpiresIn: 0,
					},
				},
				count: 2,
				err:   nil,
			},
		},
		{
			description: "succeeds when there are api keys and sorter",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			paginator:   query.Paginator{Page: 1, PerPage: 10},
			sorter:      query.Sorter{By: "expires_in", Order: query.OrderDesc},
			fixtures:    []string{fixtureAPIKeys},
			expected: Expected{
				apiKeys: []models.APIKey{
					{
						ID:        "a1b2c73ea41f70870c035283336d72228118213ed03ec78043ffee48d827af11",
						Name:      "prod",
						CreatedBy: "507f1f77bcf86cd799439011",
						TenantID:  "00000000-0000-4000-0000-000000000000",
						Role:      "operator",
						CreatedAt: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						ExpiresIn: 10,
					},
					{
						ID:        "f23a2e56cd3fcfba002c72675c870e1e7813292adc40bbf14cea479a2e07976a",
						Name:      "dev",
						CreatedBy: "507f1f77bcf86cd799439011",
						TenantID:  "00000000-0000-4000-0000-000000000000",
						Role:      "admin",
						CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						ExpiresIn: 0,
					},
				},
				count: 2,
				err:   nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			require.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() { require.NoError(t, srv.Reset()) })

			apiKeys, count, err := s.APIKeyList(ctx, tc.tenantID, tc.paginator, tc.sorter)
			require.Equal(t, tc.expected, Expected{apiKeys, count, err})
		})
	}
}

func TestAPIKeyUpdate(t *testing.T) {
	type Expected struct {
		name string
		err  error
	}

	cases := []struct {
		description string
		tenantID    string
		name        string
		changes     *models.APIKeyChanges
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when name and tenant id does not exists",
			tenantID:    "nonexistent",
			name:        "nonexistent",
			changes:     &models.APIKeyChanges{},
			fixtures:    []string{fixtureAPIKeys},
			expected: Expected{
				name: "",
				err:  store.ErrNoDocuments,
			},
		},
		{
			description: "fails when name is valid but tenant id not",
			tenantID:    "nonexistent",
			name:        "dev",
			changes:     &models.APIKeyChanges{},
			fixtures:    []string{fixtureAPIKeys},
			expected: Expected{
				name: "",
				err:  store.ErrNoDocuments,
			},
		},
		{
			description: "fails when tenant id is valid but name not",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			name:        "nonexistent",
			changes:     &models.APIKeyChanges{},
			fixtures:    []string{fixtureAPIKeys},
			expected: Expected{
				name: "",
				err:  store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when changes is empty",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			name:        "dev",
			changes:     &models.APIKeyChanges{},
			fixtures:    []string{fixtureAPIKeys},
			expected: Expected{
				name: "dev",
				err:  nil,
			},
		},
		{
			description: "succeeds when changes is not empty",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			name:        "dev",
			changes:     &models.APIKeyChanges{Name: "new"},
			fixtures:    []string{fixtureAPIKeys},
			expected: Expected{
				name: "new",
				err:  nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			require.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() { require.NoError(t, srv.Reset()) })

			err := s.APIKeyUpdate(ctx, tc.tenantID, tc.name, tc.changes)
			if tc.expected.err != nil {
				require.Equal(t, tc.expected.err, err)

				return
			}

			filter := bson.M{"tenant_id": tc.tenantID}
			if tc.expected.name != "" {
				filter = bson.M{"name": tc.expected.name}
			}

			apiKey := new(models.APIKey)
			require.NoError(t, db.Collection("api_keys").FindOne(ctx, filter).Decode(apiKey))
			require.Equal(t, tc.expected.name, apiKey.Name)
			require.WithinDuration(t, time.Now(), apiKey.UpdatedAt, 10*time.Second)
		})
	}
}

func TestDeleteAPIKey(t *testing.T) {
	cases := []struct {
		description string
		tenantID    string
		name        string
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when name and tenant id does not exists",
			tenantID:    "nonexistent",
			name:        "nonexistent",
			fixtures:    []string{fixtureAPIKeys},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "fails when name is valid but tenant id not",
			tenantID:    "nonexistent",
			name:        "dev",
			fixtures:    []string{fixtureAPIKeys},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "fails when tenant id is valid but name not",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			name:        "nonexistent",
			fixtures:    []string{fixtureAPIKeys},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			name:        "dev",
			fixtures:    []string{fixtureAPIKeys},
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			require.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() { require.NoError(t, srv.Reset()) })

			err := s.APIKeyDelete(ctx, tc.tenantID, tc.name)
			require.Equal(t, tc.expected, err)
		})
	}
}
