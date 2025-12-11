package mongo_test

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/clock"
	clockmocks "github.com/shellhub-io/shellhub/pkg/clock/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

func TestNamespaceList(t *testing.T) {
	type Expected struct {
		ns    []models.Namespace
		count int
		err   error
	}

	cases := []struct {
		description string
		opts        []store.QueryOption
		fixtures    []string
		expected    Expected
	}{
		{
			description: "succeeds when namespaces list is not empty",
			opts: []store.QueryOption{
				s.Options().Match(&query.Filters{}),
				s.Options().Paginate(&query.Paginator{Page: -1, PerPage: -1}),
			},
			fixtures: []string{fixtureNamespaces},
			expected: Expected{
				ns: []models.Namespace{
					{
						CreatedAt:            time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						Name:                 "namespace-1",
						Owner:                "507f1f77bcf86cd799439011",
						TenantID:             "00000000-0000-4000-0000-000000000000",
						DevicesAcceptedCount: 15,
						DevicesPendingCount:  3,
						DevicesRejectedCount: 2,
						DevicesRemovedCount:  1,
						Members: []models.Member{
							{
								ID:      "507f1f77bcf86cd799439011",
								AddedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
								Role:    authorizer.RoleOwner,
							},
							{
								ID:      "6509e169ae6144b2f56bf288",
								AddedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
								Role:    authorizer.RoleObserver,
							},
						},
						MaxDevices: -1,
						Settings:   &models.NamespaceSettings{SessionRecord: true},
					},
					{
						CreatedAt:            time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						Name:                 "namespace-2",
						Owner:                "6509e169ae6144b2f56bf288",
						TenantID:             "00000000-0000-4001-0000-000000000000",
						DevicesAcceptedCount: 8,
						DevicesPendingCount:  1,
						DevicesRejectedCount: 0,
						DevicesRemovedCount:  2,
						Members: []models.Member{
							{
								ID:      "6509e169ae6144b2f56bf288",
								AddedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
								Role:    authorizer.RoleOwner,
							},
							{
								ID:      "907f1f77bcf86cd799439022",
								AddedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
								Role:    authorizer.RoleOperator,
							},
						},
						MaxDevices: 10,
						Settings:   &models.NamespaceSettings{SessionRecord: false},
					},
					{
						CreatedAt:            time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						Name:                 "namespace-3",
						Owner:                "657b0e3bff780d625f74e49a",
						TenantID:             "00000000-0000-4002-0000-000000000000",
						DevicesAcceptedCount: 342,
						DevicesPendingCount:  0,
						DevicesRejectedCount: 2,
						DevicesRemovedCount:  4,
						Members: []models.Member{
							{
								ID:      "657b0e3bff780d625f74e49a",
								AddedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
								Role:    authorizer.RoleOwner,
							},
						},
						MaxDevices: 3,
						Settings:   &models.NamespaceSettings{SessionRecord: true},
					},
					{
						CreatedAt:            time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						Name:                 "namespace-4",
						Owner:                "6577267d8752d05270a4c07d",
						TenantID:             "00000000-0000-4003-0000-000000000000",
						DevicesAcceptedCount: 25,
						DevicesPendingCount:  5,
						DevicesRejectedCount: 3,
						DevicesRemovedCount:  0,
						Members: []models.Member{
							{
								ID:      "6577267d8752d05270a4c07d",
								AddedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
								Role:    authorizer.RoleOwner,
							},
						},
						MaxDevices: -1,
						Settings:   &models.NamespaceSettings{SessionRecord: true},
					},
				},
				count: 4,
				err:   nil,
			},
		},
	}

	// Due to the non-deterministic order of applying fixtures when dealing with multiple datasets,
	// we ensure that both the expected and result arrays are correctly sorted.
	sort := func(ns []models.Namespace) {
		sort.Slice(ns, func(i, j int) bool {
			return ns[i].TenantID < ns[j].TenantID
		})
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			ns, count, err := s.NamespaceList(ctx, tc.opts...)
			sort(tc.expected.ns)
			sort(ns)
			assert.Equal(t, tc.expected, Expected{ns: ns, count: count, err: err})
		})
	}
}

func TestNamespaceResolve(t *testing.T) {
	type Expected struct {
		namespace *models.Namespace
		err       error
	}

	cases := []struct {
		description string
		resolver    store.NamespaceResolver
		value       string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when namespace not found by tenant ID",
			resolver:    store.NamespaceTenantIDResolver,
			value:       "nonexistent-tenant-id",
			fixtures:    []string{fixtureNamespaces, fixtureUsers},
			expected: Expected{
				namespace: nil,
				err:       store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds resolving namespace by tenant ID",
			resolver:    store.NamespaceTenantIDResolver,
			value:       "00000000-0000-4000-0000-000000000000",
			fixtures:    []string{fixtureNamespaces, fixtureUsers},
			expected: Expected{
				namespace: &models.Namespace{
					CreatedAt:            time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					Name:                 "namespace-1",
					Owner:                "507f1f77bcf86cd799439011",
					TenantID:             "00000000-0000-4000-0000-000000000000",
					DevicesAcceptedCount: 15,
					DevicesPendingCount:  3,
					DevicesRejectedCount: 2,
					DevicesRemovedCount:  1,
					Members: []models.Member{
						{
							ID:      "507f1f77bcf86cd799439011",
							AddedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
							Role:    authorizer.RoleOwner,
							Email:   "john.doe@test.com",
						},
						{
							ID:      "6509e169ae6144b2f56bf288",
							AddedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
							Role:    authorizer.RoleObserver,
							Email:   "maria.garcia@test.com",
						},
					},
					MaxDevices: -1,
					Settings:   &models.NamespaceSettings{SessionRecord: true},
				},
				err: nil,
			},
		},
		{
			description: "fails when namespace not found by name",
			resolver:    store.NamespaceNameResolver,
			value:       "nonexistent-namespace",
			fixtures:    []string{fixtureNamespaces, fixtureUsers},
			expected: Expected{
				namespace: nil,
				err:       store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds resolving namespace by name",
			resolver:    store.NamespaceNameResolver,
			value:       "namespace-1",
			fixtures:    []string{fixtureNamespaces, fixtureUsers},
			expected: Expected{
				namespace: &models.Namespace{
					CreatedAt:            time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					Name:                 "namespace-1",
					Owner:                "507f1f77bcf86cd799439011",
					TenantID:             "00000000-0000-4000-0000-000000000000",
					DevicesAcceptedCount: 15,
					DevicesPendingCount:  3,
					DevicesRejectedCount: 2,
					DevicesRemovedCount:  1,
					Members: []models.Member{
						{
							ID:      "507f1f77bcf86cd799439011",
							AddedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
							Role:    authorizer.RoleOwner,
							Email:   "john.doe@test.com",
						},
						{
							ID:      "6509e169ae6144b2f56bf288",
							AddedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
							Role:    authorizer.RoleObserver,
							Email:   "maria.garcia@test.com",
						},
					},
					MaxDevices: -1,
					Settings:   &models.NamespaceSettings{SessionRecord: true},
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			namespace, err := s.NamespaceResolve(ctx, tc.resolver, tc.value)
			assert.Equal(t, tc.expected, Expected{namespace: namespace, err: err})
		})
	}
}

func TestNamespaceGetPreferred(t *testing.T) {
	type Expected struct {
		ns  *models.Namespace
		err error
	}

	cases := []struct {
		description string
		memberID    string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when member is not found",
			memberID:    "000000000000000000000000",
			fixtures:    []string{fixtureNamespaces},
			expected: Expected{
				ns:  nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when member is found and tenantID is empty",
			memberID:    "507f1f77bcf86cd799439011",
			fixtures:    []string{fixtureNamespaces},
			expected: Expected{
				ns: &models.Namespace{
					CreatedAt:            time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					Name:                 "namespace-1",
					Owner:                "507f1f77bcf86cd799439011",
					TenantID:             "00000000-0000-4000-0000-000000000000",
					DevicesAcceptedCount: 15,
					DevicesPendingCount:  3,
					DevicesRejectedCount: 2,
					DevicesRemovedCount:  1,
					Members: []models.Member{
						{
							ID:      "507f1f77bcf86cd799439011",
							AddedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
							Role:    authorizer.RoleOwner,
						},
						{
							ID:      "6509e169ae6144b2f56bf288",
							AddedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
							Role:    authorizer.RoleObserver,
						},
					},
					MaxDevices: -1,
					Settings:   &models.NamespaceSettings{SessionRecord: true},
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			ns, err := s.NamespaceGetPreferred(ctx, tc.memberID)
			assert.Equal(t, tc.expected, Expected{ns: ns, err: err})
		})
	}
}

func TestNamespaceCreate(t *testing.T) {
	now := time.Now()

	clockMock := new(clockmocks.Clock)
	clockMock.On("Now").Return(now)
	clock.DefaultBackend = clockMock

	type Expected struct {
		tenantID string
		err      error
	}

	cases := []struct {
		description string
		ns          *models.Namespace
		fixtures    []string
		expected    Expected
	}{
		{
			description: "succeeds when data is valid",
			ns: &models.Namespace{
				Name:     "namespace-1",
				Owner:    "507f1f77bcf86cd799439011",
				TenantID: "00000000-0000-4000-0000-000000000000",
				Members: []models.Member{
					{
						ID:   "507f1f77bcf86cd799439011",
						Role: authorizer.RoleOwner,
					},
				},
				MaxDevices: -1,
				Settings:   &models.NamespaceSettings{SessionRecord: true},
			},
			fixtures: []string{},
			expected: Expected{
				tenantID: "00000000-0000-4000-0000-000000000000",
				err:      nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			tenantID, err := s.NamespaceCreate(ctx, tc.ns)
			assert.Equal(t, tc.expected, Expected{tenantID: tenantID, err: err})
		})
	}
}

func TestNamespaceConflicts(t *testing.T) {
	type Expected struct {
		conflicts []string
		ok        bool
		err       error
	}

	cases := []struct {
		description string
		target      *models.NamespaceConflicts
		fixtures    []string
		expected    Expected
	}{
		{
			description: "no conflicts when target is empty",
			target:      &models.NamespaceConflicts{},
			fixtures:    []string{fixtureNamespaces},
			expected:    Expected{[]string{}, false, nil},
		},
		{
			description: "no conflicts with non existing name",
			target:      &models.NamespaceConflicts{Name: "nonexistent-namespace"},
			fixtures:    []string{fixtureNamespaces},
			expected:    Expected{[]string{}, false, nil},
		},
		{
			description: "conflict detected with existing name",
			target:      &models.NamespaceConflicts{Name: "namespace-1"},
			fixtures:    []string{fixtureNamespaces},
			expected:    Expected{[]string{"name"}, true, nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			require.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() { require.NoError(t, srv.Reset()) })

			conflicts, ok, err := s.NamespaceConflicts(ctx, tc.target)
			require.Equal(t, tc.expected, Expected{conflicts, ok, err})
		})
	}
}

func TestStore_NamespaceUpdate(t *testing.T) {
	cases := []struct {
		description string
		namespace   *models.Namespace
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when namespace is not found",
			namespace: &models.Namespace{
				TenantID: "nonexistent",
				Name:     "edited-namespace",
			},
			fixtures: []string{fixtureNamespaces},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when namespace is found",
			namespace: &models.Namespace{
				TenantID: "00000000-0000-4000-0000-000000000000",
				Name:     "edited-namespace",
			},
			fixtures: []string{fixtureNamespaces},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			err := s.NamespaceUpdate(ctx, tc.namespace)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestStore_NamespaceDelete(t *testing.T) {
	cases := []struct {
		description string
		namespace   *models.Namespace
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when namespace is not found",
			namespace: &models.Namespace{
				TenantID: "nonexistent",
			},
			fixtures: []string{fixtureNamespaces},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when namespace is found",
			namespace: &models.Namespace{
				TenantID: "00000000-0000-4000-0000-000000000000",
			},
			fixtures: []string{fixtureNamespaces},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			err := s.NamespaceDelete(ctx, tc.namespace)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestStore_NamespaceDeleteMany(t *testing.T) {
	cases := []struct {
		description   string
		tenantIDs     []string
		fixtures      []string
		expectedCount int64
		expectedError error
	}{
		{
			description:   "fails when no namespaces are found",
			tenantIDs:     []string{"nonexistent1", "nonexistent2"},
			fixtures:      []string{fixtureNamespaces},
			expectedCount: 0,
			expectedError: nil,
		},
		{
			description:   "succeeds deleting single namespace",
			tenantIDs:     []string{"00000000-0000-4000-0000-000000000000"},
			fixtures:      []string{fixtureNamespaces},
			expectedCount: 1,
			expectedError: nil,
		},
		{
			description:   "succeeds deleting multiple namespaces",
			tenantIDs:     []string{"00000000-0000-4000-0000-000000000000", "00000000-0000-4001-0000-000000000000", "00000000-0000-4002-0000-000000000000"},
			fixtures:      []string{fixtureNamespaces},
			expectedCount: 3,
			expectedError: nil,
		},
		{
			description:   "succeeds with mix of valid and invalid tenant IDs",
			tenantIDs:     []string{"00000000-0000-4000-0000-000000000000", "nonexistent"},
			fixtures:      []string{fixtureNamespaces},
			expectedCount: 1,
			expectedError: nil,
		},
		{
			description:   "handles empty tenant IDs list",
			tenantIDs:     []string{},
			fixtures:      []string{fixtureNamespaces},
			expectedCount: 0,
			expectedError: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			deletedCount, err := s.NamespaceDeleteMany(ctx, tc.tenantIDs)

			if tc.expectedError != nil {
				assert.Equal(t, tc.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.expectedCount, deletedCount)
		})
	}
}

func TestNamespaceIncrementDeviceCount(t *testing.T) {
	type Expected struct {
		acceptedCount int64
		pendingCount  int64
		rejectedCount int64
		err           error
	}

	cases := []struct {
		description string
		tenant      string
		status      models.DeviceStatus
		count       int64
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when tenant is not found",
			tenant:      "nonexistent",
			status:      models.DeviceStatusAccepted,
			count:       5,
			fixtures:    []string{fixtureNamespaces},
			expected: Expected{
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when incrementing devices count",
			tenant:      "00000000-0000-4000-0000-000000000000",
			status:      models.DeviceStatusAccepted,
			count:       5,
			fixtures:    []string{fixtureNamespaces},
			expected: Expected{
				acceptedCount: 20, // 15 + 5
				pendingCount:  3,
				rejectedCount: 2,
				err:           nil,
			},
		},
		{
			description: "succeeds when decrementing devices count",
			tenant:      "00000000-0000-4000-0000-000000000000",
			status:      models.DeviceStatusPending,
			count:       -2,
			fixtures:    []string{fixtureNamespaces},
			expected: Expected{
				acceptedCount: 15,
				pendingCount:  1, // 3 - 2
				rejectedCount: 2,
				err:           nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			require.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() { require.NoError(t, srv.Reset()) })

			err := s.NamespaceIncrementDeviceCount(ctx, tc.tenant, tc.status, tc.count)
			require.Equal(t, tc.expected.err, err)
			if err != nil {
				return
			}

			namespace := new(models.Namespace)
			require.NoError(t, db.Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": tc.tenant}).Decode(namespace))

			require.Equal(t, tc.expected.acceptedCount, namespace.DevicesAcceptedCount)
			require.Equal(t, tc.expected.pendingCount, namespace.DevicesPendingCount)
			require.Equal(t, tc.expected.rejectedCount, namespace.DevicesRejectedCount)
		})
	}
}
