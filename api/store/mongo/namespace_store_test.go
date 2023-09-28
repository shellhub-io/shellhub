package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/mongotest"
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
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	type Expected struct {
		ns    []models.Namespace
		count int
		err   error
	}

	cases := []struct {
		description string
		tenant      string
		setup       func() error
		expected    Expected
	}{
		{
			description: "succeeds",
			tenant:      "00000000-0000-4000-0000-000000000000",
			setup: func() error {
				return mongotest.UseFixture(fixtures.Namespace)
			},
			expected: Expected{
				ns: []models.Namespace{
					{
						CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						Name:      "namespace",
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
				},
				count: 1,
				err:   nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			ns, count, err := mongostore.NamespaceList(ctx, paginator.Query{Page: -1, PerPage: -1}, nil, false)
			assert.Equal(t, tc.expected, Expected{ns: ns, count: count, err: err})

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestNamespaceGet(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	type Expected struct {
		ns  *models.Namespace
		err error
	}

	cases := []struct {
		description string
		tenant      string
		setup       func() error
		expected    Expected
	}{
		{
			description: "fails when tenant is not found",
			tenant:      "nonexistent",
			setup: func() error {
				return mongotest.UseFixture(fixtures.Namespace, fixtures.Device)
			},
			expected: Expected{
				ns:  nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when tenant is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			setup: func() error {
				return mongotest.UseFixture(fixtures.Namespace, fixtures.Device)
			},
			expected: Expected{
				ns: &models.Namespace{
					CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					Name:      "namespace",
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
					DevicesCount: 1,
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			ns, err := mongostore.NamespaceGet(ctx, tc.tenant)
			assert.Equal(t, tc.expected, Expected{ns: ns, err: err})

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestNamespaceGetByName(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	type Expected struct {
		ns  *models.Namespace
		err error
	}

	cases := []struct {
		description string
		name        string
		setup       func() error
		expected    Expected
	}{
		{
			description: "fails when namespace is not found",
			name:        "nonexistent",
			setup: func() error {
				return mongotest.UseFixture(fixtures.Namespace)
			},
			expected: Expected{
				ns:  nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when namespace is found",
			name:        "namespace",
			setup: func() error {
				return mongotest.UseFixture(fixtures.Namespace)
			},
			expected: Expected{
				ns: &models.Namespace{
					CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					Name:      "namespace",
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

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			ns, err := mongostore.NamespaceGetByName(ctx, tc.name)
			assert.Equal(t, tc.expected, Expected{ns: ns, err: err})

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestNamespaceGetFirst(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	type Expected struct {
		ns  *models.Namespace
		err error
	}

	cases := []struct {
		description string
		member      string
		setup       func() error
		expected    Expected
	}{
		{
			description: "fails when member is not found",
			member:      "000000000000000000000000",
			setup: func() error {
				return mongotest.UseFixture(fixtures.Namespace)
			},
			expected: Expected{
				ns:  nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when member is found",
			member:      "507f1f77bcf86cd799439011",
			setup: func() error {
				return mongotest.UseFixture(fixtures.Namespace)
			},
			expected: Expected{
				ns: &models.Namespace{
					CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					Name:      "namespace",
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

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			ns, err := mongostore.NamespaceGetFirst(ctx, tc.member)
			assert.Equal(t, tc.expected, Expected{ns: ns, err: err})

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestNamespaceCreate(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	type Expected struct {
		ns  *models.Namespace
		err error
	}

	cases := []struct {
		description string
		ns          *models.Namespace
		setup       func() error
		expected    Expected
	}{
		{
			description: "succeeds when data is valid",
			ns: &models.Namespace{
				Name:     "namespace",
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
			setup: func() error {
				return nil
			},
			expected: Expected{
				ns: &models.Namespace{
					Name:     "namespace",
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

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			ns, err := mongostore.NamespaceCreate(ctx, tc.ns)
			assert.Equal(t, tc.expected, Expected{ns: ns, err: err})
		})
	}
}

func TestNamespaceRename(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	type Expected struct {
		ns  *models.Namespace
		err error
	}

	cases := []struct {
		description string
		tenant      string
		name        string
		setup       func() error
		expected    Expected
	}{
		{
			description: "fails when tenant is not found",
			tenant:      "nonexistent",
			name:        "edited-namespace",
			setup: func() error {
				return mongotest.UseFixture(fixtures.Namespace)
			},
			expected: Expected{
				ns:  nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when tenant is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			name:        "edited-namespace",
			setup: func() error {
				return mongotest.UseFixture(fixtures.Namespace)
			},
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

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			ns, err := mongostore.NamespaceRename(ctx, tc.tenant, tc.name)
			assert.Equal(t, tc.expected, Expected{ns: ns, err: err})

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestNamespaceUpdate(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description string
		tenant      string
		ns          *models.Namespace
		setup       func() error
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
			setup: func() error {
				return mongotest.UseFixture(fixtures.Namespace)
			},
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
			setup: func() error {
				return mongotest.UseFixture(fixtures.Namespace)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.NamespaceUpdate(ctx, tc.tenant, tc.ns)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestNamespaceDelete(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description string
		tenant      string
		setup       func() error
		expected    error
	}{
		{
			description: "fails when namespace is not found",
			tenant:      "nonexistent",
			setup: func() error {
				return mongotest.UseFixture(fixtures.Namespace)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when namespace is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			setup: func() error {
				return mongotest.UseFixture(fixtures.Namespace)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.NamespaceDelete(ctx, tc.tenant)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestNamespaceAddMember(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	type Expected struct {
		ns  *models.Namespace
		err error
	}

	cases := []struct {
		description string
		tenant      string
		member      string
		role        string
		setup       func() error
		expected    Expected
	}{
		{
			description: "fails when tenant is not found",
			tenant:      "nonexistent",
			member:      "6509de884238881ac1b2b289",
			role:        guard.RoleObserver,
			setup: func() error {
				return mongotest.UseFixture(fixtures.Namespace, fixtures.Member)
			},
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
			setup: func() error {
				return mongotest.UseFixture(fixtures.Namespace, fixtures.Member)
			},
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
			setup: func() error {
				return mongotest.UseFixture(fixtures.Namespace, fixtures.Member)
			},
			expected: Expected{
				ns: &models.Namespace{
					CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					Name:      "namespace",
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

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			ns, err := mongostore.NamespaceAddMember(ctx, tc.tenant, tc.member, tc.role)
			assert.Equal(t, tc.expected, Expected{ns: ns, err: err})

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestNamespaceEditMember(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description string
		tenant      string
		member      string
		role        string
		setup       func() error
		expected    error
	}{
		{
			description: "fails when user is not found",
			tenant:      "nonexistent",
			member:      "000000000000000000000000",
			role:        guard.RoleObserver,
			setup: func() error {
				return mongotest.UseFixture(fixtures.Namespace, fixtures.Member)
			},
			expected: ErrUserNotFound,
		},
		{
			description: "succeeds when tenant and user is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			member:      "6509e169ae6144b2f56bf288",
			role:        guard.RoleOperator,
			setup: func() error {
				return mongotest.UseFixture(fixtures.Namespace)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.NamespaceEditMember(ctx, tc.tenant, tc.member, tc.role)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestNamespaceRemoveMember(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	type Expected struct {
		ns  *models.Namespace
		err error
	}

	cases := []struct {
		description string
		tenant      string
		member      string
		setup       func() error
		expected    Expected
	}{
		{
			description: "fails when tenant is not found",
			tenant:      "nonexistent",
			member:      "6509de884238881ac1b2b289",
			setup: func() error {
				return mongotest.UseFixture(fixtures.Namespace)
			},
			expected: Expected{
				ns:  nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "fails when member is not found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			member:      "000000000000000000000000",
			setup: func() error {
				return mongotest.UseFixture(fixtures.Namespace)
			},
			expected: Expected{
				ns:  nil,
				err: ErrUserNotFound,
			},
		},
		{
			description: "succeeds when tenant and user is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			member:      "6509e169ae6144b2f56bf288",
			setup: func() error {
				return mongotest.UseFixture(fixtures.Namespace)
			},
			expected: Expected{
				ns: &models.Namespace{
					CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					Name:      "namespace",
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

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			ns, err := mongostore.NamespaceRemoveMember(ctx, tc.tenant, tc.member)
			assert.Equal(t, tc.expected, Expected{ns: ns, err: err})

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestNamespaceSetSessionRecord(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description string
		tenant      string
		sessionRec  bool
		setup       func() error
		expected    error
	}{
		{
			description: "fails when tenant is not found",
			tenant:      "nonexistent",
			sessionRec:  true,
			setup: func() error {
				return mongotest.UseFixture(fixtures.Namespace)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when tenant is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			sessionRec:  true,
			setup: func() error {
				return mongotest.UseFixture(fixtures.Namespace)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.NamespaceSetSessionRecord(ctx, tc.sessionRec, tc.tenant)
			assert.Equal(t, tc.expected, err)

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}

func TestNamespaceGetSessionRecord(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	type Expected struct {
		set bool
		err error
	}

	cases := []struct {
		description string
		tenant      string
		setup       func() error
		expected    Expected
	}{
		{
			description: "fails when tenant is not found",
			tenant:      "nonexistent",
			setup: func() error {
				return mongotest.UseFixture(fixtures.Namespace)
			},
			expected: Expected{
				set: false,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when tenant is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			setup: func() error {
				return mongotest.UseFixture(fixtures.Namespace)
			},
			expected: Expected{
				set: true,
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			set, err := mongostore.NamespaceGetSessionRecord(ctx, tc.tenant)
			assert.Equal(t, tc.expected, Expected{set: set, err: err})

			err = mongotest.DropDatabase()
			assert.NoError(t, err)
		})
	}
}
