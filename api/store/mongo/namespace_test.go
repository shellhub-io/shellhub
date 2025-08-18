package mongo_test

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mongo"
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
		page        query.Paginator
		filters     query.Filters
		fixtures    []string
		expected    Expected
	}{
		{
			description: "succeeds when namespaces list is not empty",
			page:        query.Paginator{Page: -1, PerPage: -1},
			filters:     query.Filters{},
			fixtures:    []string{fixtureNamespaces},
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
								Status:  models.MemberStatusAccepted,
							},
							{
								ID:      "6509e169ae6144b2f56bf288",
								AddedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
								Role:    authorizer.RoleObserver,
								Status:  models.MemberStatusPending,
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
								Status:  models.MemberStatusAccepted,
							},
							{
								ID:      "907f1f77bcf86cd799439022",
								AddedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
								Role:    authorizer.RoleOperator,
								Status:  models.MemberStatusAccepted,
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
								Status:  models.MemberStatusAccepted,
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
								Status:  models.MemberStatusAccepted,
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

			ns, count, err := s.NamespaceList(ctx, tc.page, tc.filters)
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
							Status:  models.MemberStatusAccepted,
							Email:   "john.doe@test.com",
						},
						{
							ID:      "6509e169ae6144b2f56bf288",
							AddedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
							Role:    authorizer.RoleObserver,
							Status:  models.MemberStatusPending,
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
							Status:  models.MemberStatusAccepted,
							Email:   "john.doe@test.com",
						},
						{
							ID:      "6509e169ae6144b2f56bf288",
							AddedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
							Role:    authorizer.RoleObserver,
							Status:  models.MemberStatusPending,
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
							Status:  models.MemberStatusAccepted,
						},
						{
							ID:      "6509e169ae6144b2f56bf288",
							AddedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
							Role:    authorizer.RoleObserver,
							Status:  models.MemberStatusPending,
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
		ns  *models.Namespace
		err error
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
				ns: &models.Namespace{
					CreatedAt: now,
					Name:      "namespace-1",
					Owner:     "507f1f77bcf86cd799439011",
					TenantID:  "00000000-0000-4000-0000-000000000000",
					Members: []models.Member{
						{
							ID:   "507f1f77bcf86cd799439011",
							Role: authorizer.RoleOwner,
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

			ns, err := s.NamespaceCreate(ctx, tc.ns)
			assert.Equal(t, tc.expected, Expected{ns: ns, err: err})
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

func TestNamespaceEdit(t *testing.T) {
	cases := []struct {
		description string
		tenant      string
		changes     *models.NamespaceChanges
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when tenant is not found",
			tenant:      "nonexistent",
			changes: &models.NamespaceChanges{
				Name: "edited-namespace",
			},
			fixtures: []string{fixtureNamespaces},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when tenant is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			changes: &models.NamespaceChanges{
				Name: "edited-namespace",
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

			err := s.NamespaceEdit(ctx, tc.tenant, tc.changes)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestNamespaceUpdate(t *testing.T) {
	cases := []struct {
		description string
		tenant      string
		ns          *models.Namespace
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when tenant is not found",
			tenant:      "nonexistent",
			ns: &models.Namespace{
				Name:       "edited-namespace",
				MaxDevices: 3,
				Settings:   &models.NamespaceSettings{SessionRecord: true},
			},
			fixtures: []string{fixtureNamespaces},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when tenant is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			ns: &models.Namespace{
				Name:       "edited-namespace",
				MaxDevices: 3,
				Settings:   &models.NamespaceSettings{SessionRecord: true},
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

			err := s.NamespaceUpdate(ctx, tc.tenant, tc.ns)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestNamespaceDelete(t *testing.T) {
	cases := []struct {
		description string
		tenant      string
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when namespace is not found",
			tenant:      "nonexistent",
			fixtures:    []string{fixtureNamespaces},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when namespace is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			fixtures:    []string{fixtureNamespaces},
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			err := s.NamespaceDelete(ctx, tc.tenant)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestNamespaceAddMember(t *testing.T) {
	type Expected struct {
		err error
	}

	cases := []struct {
		description string
		tenantID    string
		member      *models.Member
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when tenant is not found",
			tenantID:    "nonexistent",
			member: &models.Member{
				ID:     "6509de884238881ac1b2b289",
				Role:   authorizer.RoleObserver,
				Status: models.MemberStatusAccepted,
			},
			fixtures: []string{fixtureNamespaces},
			expected: Expected{err: store.ErrNoDocuments},
		},
		{
			description: "fails when member has already been added",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			member: &models.Member{
				ID:     "6509e169ae6144b2f56bf288",
				Role:   authorizer.RoleObserver,
				Status: models.MemberStatusAccepted,
			},
			fixtures: []string{fixtureNamespaces},
			expected: Expected{err: mongo.ErrNamespaceDuplicatedMember},
		},
		{
			description: "succeeds when tenant is found",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			member: &models.Member{
				ID:     "6509de884238881ac1b2b289",
				Role:   authorizer.RoleObserver,
				Status: models.MemberStatusAccepted,
			},
			fixtures: []string{fixtureNamespaces},
			expected: Expected{err: nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			if err := s.NamespaceAddMember(ctx, tc.tenantID, tc.member); tc.expected.err != nil {
				require.Equal(t, tc.expected.err, err)

				return
			}

			require.NoError(t, db.Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": tc.tenantID, "members.id": tc.member.ID}).Err())
		})
	}
}

func TestNamespaceUpdateMember(t *testing.T) {
	type Expected struct {
		err error
	}

	cases := []struct {
		description string
		tenantID    string
		memberID    string
		changes     *models.MemberChanges
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when user is not found",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			memberID:    "000000000000000000000000",
			changes: &models.MemberChanges{
				Role:   authorizer.RoleObserver,
				Status: models.MemberStatusPending,
			},
			fixtures: []string{fixtureNamespaces},
			expected: Expected{err: mongo.ErrUserNotFound},
		},
		{
			description: "succeeds when tenant and user is found",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			memberID:    "6509e169ae6144b2f56bf288",
			changes: &models.MemberChanges{
				Role:   authorizer.RoleAdministrator,
				Status: models.MemberStatusPending,
			},
			fixtures: []string{fixtureNamespaces},
			expected: Expected{err: nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			if err := s.NamespaceUpdateMember(ctx, tc.tenantID, tc.memberID, tc.changes); tc.expected.err != nil {
				require.Equal(t, tc.expected.err, err)

				return
			}

			namespace := new(models.Namespace)
			require.NoError(t, db.Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": tc.tenantID, "members.id": tc.memberID}).Decode(namespace))
			require.Equal(t, 2, len(namespace.Members))
			require.Equal(t, tc.memberID, namespace.Members[1].ID)
			require.Equal(t, tc.changes.Role, namespace.Members[1].Role)
			require.Equal(t, tc.changes.Status, namespace.Members[1].Status)
		})
	}
}

func TestNamespaceRemoveMember(t *testing.T) {
	type Expected struct {
		err error
	}

	cases := []struct {
		description string
		tenantID    string
		memberID    string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when tenant is not found",
			tenantID:    "nonexistent",
			memberID:    "6509de884238881ac1b2b289",
			fixtures:    []string{fixtureNamespaces},
			expected:    Expected{err: store.ErrNoDocuments},
		},
		{
			description: "fails when member is not found",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			memberID:    "nonexistent",
			fixtures:    []string{fixtureNamespaces},
			expected:    Expected{err: mongo.ErrUserNotFound},
		},
		{
			description: "succeeds when tenant and user is found",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			memberID:    "6509e169ae6144b2f56bf288",
			fixtures:    []string{fixtureNamespaces},
			expected:    Expected{err: nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			if err := s.NamespaceRemoveMember(ctx, tc.tenantID, tc.memberID); tc.expected.err != nil {
				require.Equal(t, tc.expected.err, err)

				return
			}

			namespace := new(models.Namespace)
			require.NoError(t, db.Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": tc.tenantID}).Decode(namespace))
			require.Equal(t, 1, len(namespace.Members))
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
