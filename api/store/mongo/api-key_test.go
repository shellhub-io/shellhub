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

func TestAPIKeyResolve(t *testing.T) {
	type Expected struct {
		apiKey *models.APIKey
		err    error
	}

	cases := []struct {
		description string
		tenantID    string
		resolver    store.APIKeyResolver
		value       string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when ID does not exist",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			resolver:    store.APIKeyIDResolver,
			value:       "nonexistent-id",
			fixtures:    []string{fixtureAPIKeys},
			expected: Expected{
				apiKey: nil,
				err:    store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds resolving API key by ID",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			resolver:    store.APIKeyIDResolver,
			value:       "f23a2e56cd3fcfba002c72675c870e1e7813292adc40bbf14cea479a2e07976a",
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
		{
			description: "fails when name and tenant ID does not exist",
			tenantID:    "nonexistent-tenant",
			resolver:    store.APIKeyNameResolver,
			value:       "nonexistent",
			fixtures:    []string{fixtureAPIKeys},
			expected: Expected{
				apiKey: nil,
				err:    store.ErrNoDocuments,
			},
		},
		{
			description: "fails when name is valid but tenant ID is not",
			tenantID:    "nonexistent-tenant",
			resolver:    store.APIKeyNameResolver,
			value:       "dev",
			fixtures:    []string{fixtureAPIKeys},
			expected: Expected{
				apiKey: nil,
				err:    store.ErrNoDocuments,
			},
		},
		{
			description: "fails when tenant ID is valid but name is not",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			resolver:    store.APIKeyNameResolver,
			value:       "nonexistent",
			fixtures:    []string{fixtureAPIKeys},
			expected: Expected{
				apiKey: nil,
				err:    store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds resolving API key by name",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			resolver:    store.APIKeyNameResolver,
			value:       "dev",
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
		{
			description: "fails when API key exists but belongs to different tenant",
			tenantID:    "11111111-1111-4111-1111-111111111111",
			resolver:    store.APIKeyNameResolver,
			value:       "dev",
			fixtures:    []string{fixtureAPIKeys},
			expected: Expected{
				apiKey: nil,
				err:    store.ErrNoDocuments,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			require.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() { require.NoError(t, srv.Reset()) })

			apiKey, err := s.APIKeyResolve(ctx, tc.resolver, tc.value, s.Options().InNamespace(tc.tenantID))
			require.Equal(t, tc.expected, Expected{apiKey: apiKey, err: err})
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
		opts        []store.QueryOption
		fixtures    []string
		expected    Expected
	}{
		{
			description: "succeeds when there are no api keys",
			opts:        []store.QueryOption{s.Options().InNamespace("non-existent"), s.Options().Sort(&query.Sorter{By: "expires_in", Order: query.OrderAsc}), s.Options().Paginate(&query.Paginator{Page: 1, PerPage: 10})},
			fixtures:    []string{fixtureAPIKeys},
			expected: Expected{
				apiKeys: []models.APIKey{},
				count:   0,
				err:     nil,
			},
		},
		{
			description: "succeeds when there are api keys",
			opts:        []store.QueryOption{s.Options().InNamespace("00000000-0000-4000-0000-000000000000"), s.Options().Sort(&query.Sorter{By: "expires_in", Order: query.OrderAsc}), s.Options().Paginate(&query.Paginator{Page: 1, PerPage: 10})},
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
			opts:        []store.QueryOption{s.Options().InNamespace("00000000-0000-4000-0000-000000000000"), s.Options().Sort(&query.Sorter{By: "expires_in", Order: query.OrderAsc}), s.Options().Paginate(&query.Paginator{Page: 1, PerPage: 1})},
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
			opts:        []store.QueryOption{s.Options().InNamespace("00000000-0000-4000-0000-000000000000"), s.Options().Sort(&query.Sorter{By: "expires_in", Order: query.OrderDesc}), s.Options().Paginate(&query.Paginator{Page: 1, PerPage: 10})},
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

			apiKeys, count, err := s.APIKeyList(ctx, tc.opts...)
			require.Equal(t, tc.expected, Expected{apiKeys, count, err})
		})
	}
}

func TestAPIKeyUpdate(t *testing.T) {
	type Expected struct {
		updatedName string
		err         error
	}

	cases := []struct {
		description string
		apiKey      *models.APIKey
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when API key does not exist",
			apiKey: &models.APIKey{
				ID:       "nonexistent-id",
				TenantID: "00000000-0000-4000-0000-000000000000",
				Name:     "nonexistent",
			},
			fixtures: []string{fixtureAPIKeys},
			expected: Expected{
				updatedName: "",
				err:         store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when API key exists",
			apiKey: &models.APIKey{
				ID:        "f23a2e56cd3fcfba002c72675c870e1e7813292adc40bbf14cea479a2e07976a",
				Name:      "updated-dev",
				CreatedBy: "507f1f77bcf86cd799439011",
				TenantID:  "00000000-0000-4000-0000-000000000000",
				Role:      "admin",
				ExpiresIn: 0,
			},
			fixtures: []string{fixtureAPIKeys},
			expected: Expected{
				updatedName: "updated-dev",
				err:         nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			require.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() { require.NoError(t, srv.Reset()) })

			err := s.APIKeyUpdate(ctx, tc.apiKey)
			if tc.expected.err != nil {
				require.Equal(t, tc.expected.err, err)

				return
			}

			require.NoError(t, err)

			apiKey := new(models.APIKey)
			require.NoError(t, db.Collection("api_keys").FindOne(ctx, bson.M{"_id": tc.apiKey.ID}).Decode(apiKey))
			require.Equal(t, tc.expected.updatedName, apiKey.Name)
			require.WithinDuration(t, time.Now(), apiKey.UpdatedAt, 1*time.Second)
		})
	}
}

func TestAPIKeyDelete(t *testing.T) {
	cases := []struct {
		description string
		apiKey      *models.APIKey
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when API key does not exist",
			apiKey: &models.APIKey{
				ID:       "nonexistent-id",
				TenantID: "00000000-0000-4000-0000-000000000000",
				Name:     "nonexistent",
			},
			fixtures: []string{fixtureAPIKeys},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when API key exists",
			apiKey: &models.APIKey{
				ID:       "f23a2e56cd3fcfba002c72675c870e1e7813292adc40bbf14cea479a2e07976a",
				TenantID: "00000000-0000-4000-0000-000000000000",
				Name:     "dev",
			},
			fixtures: []string{fixtureAPIKeys},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			require.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() { require.NoError(t, srv.Reset()) })

			err := s.APIKeyDelete(ctx, tc.apiKey)
			require.Equal(t, tc.expected, err)

			if tc.expected == nil {
				count, err := db.Collection("api_keys").CountDocuments(ctx, bson.M{"_id": tc.apiKey.ID})
				require.NoError(t, err)
				require.Equal(t, int64(0), count)
			}
		})
	}
}
