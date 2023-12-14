package mongo

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/shellhub-io/shellhub/api/pkg/guard"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestNamespaceList(t *testing.T) {
	type Expected struct {
		ns    []models.Namespace
		count int
		err   error
	}

	cases := []struct {
		description string
		page        paginator.Query
		filters     []models.Filter
		export      bool
		fixtures    []string
		expected    Expected
	}{
		{
			description: "succeeds when namespaces list is not empty",
			page:        paginator.Query{Page: -1, PerPage: -1},
			filters:     []models.Filter{},
			export:      false,
			fixtures:    []string{fixtures.FixtureNamespaces},
			expected: Expected{
				ns: []models.Namespace{
					{
						CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						Name:      "namespace-1",
						Owner:     "507f1f77bcf86cd799439011",
						TenantID:  "00000000-0000-4000-0000-000000000000",
						Members: []models.Member{
							{
								ID:   "507f1f77bcf86cd799439011",
								Role: guard.RoleOwner,
							},
							{
								ID:   "6509e169ae6144b2f56bf288",
								Role: guard.RoleObserver,
							},
						},
						MaxDevices: -1,
						Settings:   &models.NamespaceSettings{SessionRecord: true},
					},
					{
						CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						Name:      "namespace-2",
						Owner:     "6509e169ae6144b2f56bf288",
						TenantID:  "00000000-0000-4001-0000-000000000000",
						Members: []models.Member{
							{
								ID:   "6509e169ae6144b2f56bf288",
								Role: guard.RoleOwner,
							},
							{
								ID:   "907f1f77bcf86cd799439022",
								Role: guard.RoleOperator,
							},
						},
						MaxDevices: 10,
						Settings:   &models.NamespaceSettings{SessionRecord: false},
					},
					{
						CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						Name:      "namespace-3",
						Owner:     "657b0e3bff780d625f74e49a",
						TenantID:  "00000000-0000-4002-0000-000000000000",
						Members: []models.Member{
							{
								ID:   "657b0e3bff780d625f74e49a",
								Role: guard.RoleOwner,
							},
						},
						MaxDevices: 3,
						Settings:   &models.NamespaceSettings{SessionRecord: true},
					},
					{
						CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						Name:      "namespace-4",
						Owner:     "6577267d8752d05270a4c07d",
						TenantID:  "00000000-0000-4003-0000-000000000000",
						Members: []models.Member{
							{
								ID:   "6577267d8752d05270a4c07d",
								Role: guard.RoleOwner,
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

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	// Due to the non-deterministic order of applying fixtures when dealing with multiple datasets,
	// we ensure that both the expected and result arrays are correctly sorted.
	sort := func(ns []models.Namespace) {
		sort.Slice(ns, func(i, j int) bool {
			return ns[i].TenantID < ns[j].TenantID
		})
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			ns, count, err := mongostore.NamespaceList(context.TODO(), tc.page, tc.filters, tc.export)
			sort(tc.expected.ns)
			sort(ns)
			assert.Equal(t, tc.expected, Expected{ns: ns, count: count, err: err})
		})
	}
}

func TestNamespaceGet(t *testing.T) {
	type Expected struct {
		ns  *models.Namespace
		err error
	}

	cases := []struct {
		description string
		tenant      string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when tenant is not found",
			tenant:      "nonexistent",
			fixtures:    []string{fixtures.FixtureNamespaces, fixtures.FixtureDevices},
			expected: Expected{
				ns:  nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when tenant is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			fixtures:    []string{fixtures.FixtureNamespaces, fixtures.FixtureDevices},
			expected: Expected{
				ns: &models.Namespace{
					CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					Name:      "namespace-1",
					Owner:     "507f1f77bcf86cd799439011",
					TenantID:  "00000000-0000-4000-0000-000000000000",
					Members: []models.Member{
						{
							ID:   "507f1f77bcf86cd799439011",
							Role: guard.RoleOwner,
						},
						{
							ID:   "6509e169ae6144b2f56bf288",
							Role: guard.RoleObserver,
						},
					},
					MaxDevices:   -1,
					Settings:     &models.NamespaceSettings{SessionRecord: true},
					DevicesCount: 3,
				},
				err: nil,
			},
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			ns, err := mongostore.NamespaceGet(context.TODO(), tc.tenant)
			assert.Equal(t, tc.expected, Expected{ns: ns, err: err})
		})
	}
}

func TestNamespaceGetByName(t *testing.T) {
	type Expected struct {
		ns  *models.Namespace
		err error
	}

	cases := []struct {
		description string
		name        string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when namespace is not found",
			name:        "nonexistent",
			fixtures:    []string{fixtures.FixtureNamespaces},
			expected: Expected{
				ns:  nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when namespace is found",
			name:        "namespace-1",
			fixtures:    []string{fixtures.FixtureNamespaces},
			expected: Expected{
				ns: &models.Namespace{
					CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					Name:      "namespace-1",
					Owner:     "507f1f77bcf86cd799439011",
					TenantID:  "00000000-0000-4000-0000-000000000000",
					Members: []models.Member{
						{
							ID:   "507f1f77bcf86cd799439011",
							Role: guard.RoleOwner,
						},
						{
							ID:   "6509e169ae6144b2f56bf288",
							Role: guard.RoleObserver,
						},
					},
					MaxDevices: -1,
					Settings:   &models.NamespaceSettings{SessionRecord: true},
				},
				err: nil,
			},
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			ns, err := mongostore.NamespaceGetByName(context.TODO(), tc.name)
			assert.Equal(t, tc.expected, Expected{ns: ns, err: err})
		})
	}
}

func TestNamespaceGetFirst(t *testing.T) {
	type Expected struct {
		ns  *models.Namespace
		err error
	}

	cases := []struct {
		description string
		member      string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when member is not found",
			member:      "000000000000000000000000",
			fixtures:    []string{fixtures.FixtureNamespaces},
			expected: Expected{
				ns:  nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when member is found",
			member:      "507f1f77bcf86cd799439011",
			fixtures:    []string{fixtures.FixtureNamespaces},
			expected: Expected{
				ns: &models.Namespace{
					CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					Name:      "namespace-1",
					Owner:     "507f1f77bcf86cd799439011",
					TenantID:  "00000000-0000-4000-0000-000000000000",
					Members: []models.Member{
						{
							ID:   "507f1f77bcf86cd799439011",
							Role: guard.RoleOwner,
						},
						{
							ID:   "6509e169ae6144b2f56bf288",
							Role: guard.RoleObserver,
						},
					},
					MaxDevices: -1,
					Settings:   &models.NamespaceSettings{SessionRecord: true},
				},
				err: nil,
			},
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			ns, err := mongostore.NamespaceGetFirst(context.TODO(), tc.member)
			assert.Equal(t, tc.expected, Expected{ns: ns, err: err})
		})
	}
}

func TestNamespaceCreate(t *testing.T) {
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
						Role: guard.RoleOwner,
					},
				},
				MaxDevices: -1,
				Settings:   &models.NamespaceSettings{SessionRecord: true},
			},
			fixtures: []string{},
			expected: Expected{
				ns: &models.Namespace{
					Name:     "namespace-1",
					Owner:    "507f1f77bcf86cd799439011",
					TenantID: "00000000-0000-4000-0000-000000000000",
					Members: []models.Member{
						{
							ID:   "507f1f77bcf86cd799439011",
							Role: guard.RoleOwner,
						},
					},
					MaxDevices: -1,
					Settings:   &models.NamespaceSettings{SessionRecord: true},
				},
				err: nil,
			},
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			ns, err := mongostore.NamespaceCreate(context.TODO(), tc.ns)
			assert.Equal(t, tc.expected, Expected{ns: ns, err: err})
		})
	}
}

func TestNamespaceRename(t *testing.T) {
	type Expected struct {
		ns  *models.Namespace
		err error
	}

	cases := []struct {
		description string
		tenant      string
		name        string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when tenant is not found",
			tenant:      "nonexistent",
			name:        "edited-namespace",
			fixtures:    []string{fixtures.FixtureNamespaces},
			expected: Expected{
				ns:  nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when tenant is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			name:        "edited-namespace",
			fixtures:    []string{fixtures.FixtureNamespaces},
			expected: Expected{
				ns: &models.Namespace{
					CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					Name:      "edited-namespace",
					Owner:     "507f1f77bcf86cd799439011",
					TenantID:  "00000000-0000-4000-0000-000000000000",
					Members: []models.Member{
						{
							ID:   "507f1f77bcf86cd799439011",
							Role: guard.RoleOwner,
						},
						{
							ID:   "6509e169ae6144b2f56bf288",
							Role: guard.RoleObserver,
						},
					},
					MaxDevices:   -1,
					Settings:     &models.NamespaceSettings{SessionRecord: true},
					DevicesCount: 0,
				},
				err: nil,
			},
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			ns, err := mongostore.NamespaceRename(context.TODO(), tc.tenant, tc.name)
			assert.Equal(t, tc.expected, Expected{ns: ns, err: err})
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
			fixtures: []string{fixtures.FixtureNamespaces},
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
			fixtures: []string{fixtures.FixtureNamespaces},
			expected: nil,
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			err := mongostore.NamespaceUpdate(context.TODO(), tc.tenant, tc.ns)
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
			fixtures:    []string{fixtures.FixtureNamespaces},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when namespace is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			fixtures:    []string{fixtures.FixtureNamespaces},
			expected:    nil,
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			err := mongostore.NamespaceDelete(context.TODO(), tc.tenant)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestNamespaceAddMember(t *testing.T) {
	type Expected struct {
		ns  *models.Namespace
		err error
	}

	cases := []struct {
		description string
		tenant      string
		member      string
		role        string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when tenant is not found",
			tenant:      "nonexistent",
			member:      "6509de884238881ac1b2b289",
			role:        guard.RoleObserver,
			fixtures:    []string{fixtures.FixtureNamespaces},
			expected: Expected{
				ns:  nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "fails when member has already been added",
			tenant:      "00000000-0000-4000-0000-000000000000",
			member:      "6509e169ae6144b2f56bf288",
			role:        guard.RoleObserver,
			fixtures:    []string{fixtures.FixtureNamespaces},
			expected: Expected{
				ns:  nil,
				err: ErrNamespaceDuplicatedMember,
			},
		},
		{
			description: "succeeds when tenant is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			member:      "6509de884238881ac1b2b289",
			role:        guard.RoleObserver,
			fixtures:    []string{fixtures.FixtureNamespaces},
			expected: Expected{
				ns: &models.Namespace{
					CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					Name:      "namespace-1",
					Owner:     "507f1f77bcf86cd799439011",
					TenantID:  "00000000-0000-4000-0000-000000000000",
					Members: []models.Member{
						{
							ID:   "507f1f77bcf86cd799439011",
							Role: guard.RoleOwner,
						},
						{
							ID:   "6509e169ae6144b2f56bf288",
							Role: guard.RoleObserver,
						},
						{
							ID:   "6509de884238881ac1b2b289",
							Role: guard.RoleObserver,
						},
					},
					MaxDevices:   -1,
					Settings:     &models.NamespaceSettings{SessionRecord: true},
					DevicesCount: 0,
				},
				err: nil,
			},
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			ns, err := mongostore.NamespaceAddMember(context.TODO(), tc.tenant, tc.member, tc.role)
			assert.Equal(t, tc.expected, Expected{ns: ns, err: err})
		})
	}
}

func TestNamespaceEditMember(t *testing.T) {
	cases := []struct {
		description string
		tenant      string
		member      string
		role        string
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when user is not found",
			tenant:      "nonexistent",
			member:      "000000000000000000000000",
			role:        guard.RoleObserver,
			fixtures:    []string{fixtures.FixtureNamespaces},
			expected:    ErrUserNotFound,
		},
		{
			description: "succeeds when tenant and user is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			member:      "6509e169ae6144b2f56bf288",
			role:        guard.RoleOperator,
			fixtures:    []string{fixtures.FixtureNamespaces},
			expected:    nil,
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			err := mongostore.NamespaceEditMember(context.TODO(), tc.tenant, tc.member, tc.role)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestNamespaceRemoveMember(t *testing.T) {
	type Expected struct {
		ns  *models.Namespace
		err error
	}

	cases := []struct {
		description string
		tenant      string
		member      string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when tenant is not found",
			tenant:      "nonexistent",
			member:      "6509de884238881ac1b2b289",
			fixtures:    []string{fixtures.FixtureNamespaces},
			expected: Expected{
				ns:  nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "fails when member is not found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			member:      "nonexistent",
			fixtures:    []string{fixtures.FixtureNamespaces},
			expected: Expected{
				ns:  nil,
				err: ErrUserNotFound,
			},
		},
		{
			description: "succeeds when tenant and user is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			member:      "6509e169ae6144b2f56bf288",
			fixtures:    []string{fixtures.FixtureNamespaces},
			expected: Expected{
				ns: &models.Namespace{
					CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					Name:      "namespace-1",
					Owner:     "507f1f77bcf86cd799439011",
					TenantID:  "00000000-0000-4000-0000-000000000000",
					Members: []models.Member{
						{
							ID:   "507f1f77bcf86cd799439011",
							Role: guard.RoleOwner,
						},
					},
					MaxDevices:   -1,
					Settings:     &models.NamespaceSettings{SessionRecord: true},
					DevicesCount: 0,
				},
				err: nil,
			},
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			ns, err := mongostore.NamespaceRemoveMember(context.TODO(), tc.tenant, tc.member)
			assert.Equal(t, tc.expected, Expected{ns: ns, err: err})
		})
	}
}

func TestNamespaceSetSessionRecord(t *testing.T) {
	cases := []struct {
		description string
		tenant      string
		sessionRec  bool
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when tenant is not found",
			tenant:      "nonexistent",
			sessionRec:  true,
			fixtures:    []string{fixtures.FixtureNamespaces},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when tenant is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			sessionRec:  true,
			fixtures:    []string{fixtures.FixtureNamespaces},
			expected:    nil,
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			err := mongostore.NamespaceSetSessionRecord(context.TODO(), tc.sessionRec, tc.tenant)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestNamespaceGetSessionRecord(t *testing.T) {
	type Expected struct {
		set bool
		err error
	}

	cases := []struct {
		description string
		tenant      string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when tenant is not found",
			tenant:      "nonexistent",
			fixtures:    []string{fixtures.FixtureNamespaces},
			expected: Expected{
				set: false,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when tenant is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			fixtures:    []string{fixtures.FixtureNamespaces},
			expected: Expected{
				set: true,
				err: nil,
			},
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			set, err := mongostore.NamespaceGetSessionRecord(context.TODO(), tc.tenant)
			assert.Equal(t, tc.expected, Expected{set: set, err: err})
		})
	}
}
