package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/pinzolo/mongotest"
	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestDeviceCreate(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	cases := []struct {
		description string
		hostname    string
		device      models.Device
		expected    error
	}{
		{
			description: "succeeds when all data is valid",
			hostname:    "hostname",
			device: models.Device{
				UID: "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
				Identity: &models.DeviceIdentity{
					MAC: "mac",
				},
				TenantID: "00000000-0000-4000-0000-000000000000",
				LastSeen: clock.Now(),
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := mongostore.DeviceCreate(ctx, tc.device, tc.hostname)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestDeviceGet(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	type Expected struct {
		dev *models.Device
		err error
	}
	cases := []struct {
		description string
		uid         models.UID
		setup       func() error
		expected    Expected
	}{
		{
			description: "fails when namespace is not found",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceAccepted)
			},
			expected: Expected{
				dev: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "fails when device is not found",
			uid:         models.UID("nonexistent"),
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceAccepted, fixtures.NamespaceUnlimited)
			},
			expected: Expected{
				dev: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "fails when device is not found due to tenant",
			uid:         models.UID("5600560h6ed5h960969e7f358g4568491247198ge8537e9g448609fff1b231f"),
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceWithInvalidTenant, fixtures.NamespaceUnlimited)
			},
			expected: Expected{
				dev: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when device is found",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceAccepted, fixtures.NamespaceUnlimited)
			},
			expected: Expected{
				dev: &models.Device{
					UID:              "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
					Name:             "hostname_accepted",
					Identity:         &models.DeviceIdentity{MAC: "mac"},
					Info:             nil,
					PublicKey:        "",
					TenantID:         "00000000-0000-4000-0000-000000000000",
					LastSeen:         time.Date(2023, 9, 12, 20, 24, 14, 0, time.UTC),
					Online:           true,
					Namespace:        "namespace_unlimited",
					Status:           "accepted",
					StatusUpdatedAt:  time.Date(2023, 9, 12, 20, 24, 16, 0, time.UTC),
					CreatedAt:        time.Date(2023, 9, 12, 20, 24, 16, 0, time.UTC),
					RemoteAddr:       "",
					Position:         nil,
					Tags:             []string{},
					PublicURL:        false,
					PublicURLAddress: "",
					Acceptable:       false,
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			dev, err := mongostore.DeviceGet(ctx, tc.uid)
			assert.Equal(t, tc.expected, Expected{dev: dev, err: err})
		})
	}
}

func TestDeviceGetByMac(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	type Expected struct {
		dev *models.Device
		err error
	}
	cases := []struct {
		description string
		mac         string
		tenant      string
		status      models.DeviceStatus
		setup       func() error
		expected    Expected
	}{
		{
			description: "fails when device is not found due to mac",
			mac:         "nonexistent",
			tenant:      "00000000-0000-4000-0000-000000000000",
			status:      models.DeviceStatus(""),
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceAccepted)
			},
			expected: Expected{
				dev: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "fails when device is not found due to tenant",
			mac:         "mac",
			tenant:      "nonexistent",
			status:      models.DeviceStatus(""),
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceAccepted)
			},
			expected: Expected{
				dev: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when device is found",
			mac:         "mac",
			tenant:      "00000000-0000-4000-0000-000000000000",
			status:      models.DeviceStatus(""),
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceAccepted)
			},
			expected: Expected{
				dev: &models.Device{
					UID:              "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
					Name:             "hostname_accepted",
					Identity:         &models.DeviceIdentity{MAC: "mac"},
					Info:             nil,
					PublicKey:        "",
					TenantID:         "00000000-0000-4000-0000-000000000000",
					LastSeen:         time.Date(2023, 9, 12, 20, 24, 14, 0, time.UTC),
					Online:           false,
					Status:           "accepted",
					StatusUpdatedAt:  time.Date(2023, 9, 12, 20, 24, 16, 0, time.UTC),
					CreatedAt:        time.Date(2023, 9, 12, 20, 24, 16, 0, time.UTC),
					RemoteAddr:       "",
					Position:         nil,
					Tags:             []string{},
					PublicURL:        false,
					PublicURLAddress: "",
					Acceptable:       false,
				},
				err: nil,
			},
		},
		{
			description: "succeeds when device with status is found",
			mac:         "mac",
			tenant:      "00000000-0000-4000-0000-000000000000",
			status:      models.DeviceStatus("accepted"),
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceAccepted)
			},
			expected: Expected{
				dev: &models.Device{
					UID:              "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
					Name:             "hostname_accepted",
					Identity:         &models.DeviceIdentity{MAC: "mac"},
					Info:             nil,
					PublicKey:        "",
					TenantID:         "00000000-0000-4000-0000-000000000000",
					LastSeen:         time.Date(2023, 9, 12, 20, 24, 14, 0, time.UTC),
					Online:           false,
					Status:           "accepted",
					StatusUpdatedAt:  time.Date(2023, 9, 12, 20, 24, 16, 0, time.UTC),
					CreatedAt:        time.Date(2023, 9, 12, 20, 24, 16, 0, time.UTC),
					RemoteAddr:       "",
					Position:         nil,
					Tags:             []string{},
					PublicURL:        false,
					PublicURLAddress: "",
					Acceptable:       false,
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			dev, err := mongostore.DeviceGetByMac(ctx, tc.mac, tc.tenant, tc.status)
			assert.Equal(t, tc.expected, Expected{dev: dev, err: err})
		})
	}
}

func TestDeviceGetByName(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	type Expected struct {
		dev *models.Device
		err error
	}
	cases := []struct {
		description string
		hostname    string
		tenant      string
		setup       func() error
		expected    Expected
	}{
		{
			description: "fails when device is not found due to name",
			hostname:    "nonexistent",
			tenant:      "00000000-0000-4000-0000-000000000000",
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceAccepted)
			},
			expected: Expected{
				dev: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "fails when device is not found due to tenant",
			hostname:    "hostname_accepted",
			tenant:      "nonexistent",
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceAccepted)
			},
			expected: Expected{
				dev: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when device is found",
			hostname:    "hostname_accepted",
			tenant:      "00000000-0000-4000-0000-000000000000",
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceAccepted)
			},
			expected: Expected{
				dev: &models.Device{
					UID:              "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
					Name:             "hostname_accepted",
					Identity:         &models.DeviceIdentity{MAC: "mac"},
					Info:             nil,
					PublicKey:        "",
					TenantID:         "00000000-0000-4000-0000-000000000000",
					LastSeen:         time.Date(2023, 9, 12, 20, 24, 14, 0, time.UTC),
					Online:           false,
					Status:           "accepted",
					StatusUpdatedAt:  time.Date(2023, 9, 12, 20, 24, 16, 0, time.UTC),
					CreatedAt:        time.Date(2023, 9, 12, 20, 24, 16, 0, time.UTC),
					RemoteAddr:       "",
					Position:         nil,
					Tags:             []string{},
					PublicURL:        false,
					PublicURLAddress: "",
					Acceptable:       false,
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			dev, err := mongostore.DeviceGetByName(ctx, tc.hostname, tc.tenant)
			assert.Equal(t, tc.expected, Expected{dev: dev, err: err})
		})
	}
}

func TestDeviceGetByUID(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	type Expected struct {
		dev *models.Device
		err error
	}
	cases := []struct {
		description string
		uid         models.UID
		tenant      string
		setup       func() error
		expected    Expected
	}{
		{
			description: "fails when device is not found due to UID",
			uid:         models.UID("nonexistent"),
			tenant:      "00000000-0000-4000-0000-000000000000",
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceAccepted)
			},
			expected: Expected{
				dev: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "fails when device is not found due to tenant",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			tenant:      "nonexistent",
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceAccepted)
			},
			expected: Expected{
				dev: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when device is found",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			tenant:      "00000000-0000-4000-0000-000000000000",
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceAccepted)
			},
			expected: Expected{
				dev: &models.Device{
					UID:              "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
					Name:             "hostname_accepted",
					Identity:         &models.DeviceIdentity{MAC: "mac"},
					Info:             nil,
					PublicKey:        "",
					TenantID:         "00000000-0000-4000-0000-000000000000",
					LastSeen:         time.Date(2023, 9, 12, 20, 24, 14, 0, time.UTC),
					Online:           false,
					Status:           "accepted",
					StatusUpdatedAt:  time.Date(2023, 9, 12, 20, 24, 16, 0, time.UTC),
					CreatedAt:        time.Date(2023, 9, 12, 20, 24, 16, 0, time.UTC),
					RemoteAddr:       "",
					Position:         nil,
					Tags:             []string{},
					PublicURL:        false,
					PublicURLAddress: "",
					Acceptable:       false,
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			dev, err := mongostore.DeviceGetByUID(ctx, tc.uid, tc.tenant)
			assert.Equal(t, tc.expected, Expected{dev: dev, err: err})
		})
	}
}

func TestDeviceLookup(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	type Expected struct {
		dev *models.Device
		err error
	}
	cases := []struct {
		description string
		namespace   string
		hostname    string
		setup       func() error
		expected    Expected
	}{
		{
			description: "fails when namespace does not exist",
			namespace:   "nonexistent",
			hostname:    "hostname_accepted",
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceAccepted, fixtures.NamespaceUnlimited)
			},
			expected: Expected{
				dev: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "fails when device does not exist due to name",
			namespace:   "namespace_unlimited",
			hostname:    "nonexistent",
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceAccepted, fixtures.NamespaceUnlimited)
			},
			expected: Expected{
				dev: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "fails when device does not exist due to tenant-id",
			namespace:   "namespace_unlimited",
			hostname:    "hostname_with_invalid_tenant_id",
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceWithInvalidTenant, fixtures.NamespaceUnlimited)
			},
			expected: Expected{
				dev: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "fails when device does not exist due to status other than accepted",
			namespace:   "namespace_unlimited",
			hostname:    "hostname_pending",
			setup: func() error {
				return mongotest.UseFixture(fixtures.DevicePending, fixtures.NamespaceUnlimited)
			},
			expected: Expected{
				dev: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when namespace exists and hostname status is accepted",
			namespace:   "namespace_unlimited",
			hostname:    "hostname_accepted",
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceAccepted, fixtures.NamespaceUnlimited)
			},
			expected: Expected{
				dev: &models.Device{
					UID:              "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
					Name:             "hostname_accepted",
					Identity:         &models.DeviceIdentity{MAC: "mac"},
					Info:             nil,
					PublicKey:        "",
					TenantID:         "00000000-0000-4000-0000-000000000000",
					LastSeen:         time.Date(2023, 9, 12, 20, 24, 14, 0, time.UTC),
					Online:           false,
					Status:           "accepted",
					StatusUpdatedAt:  time.Date(2023, 9, 12, 20, 24, 16, 0, time.UTC),
					CreatedAt:        time.Date(2023, 9, 12, 20, 24, 16, 0, time.UTC),
					RemoteAddr:       "",
					Position:         nil,
					Tags:             []string{},
					PublicURL:        false,
					PublicURLAddress: "",
					Acceptable:       false,
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			dev, err := mongostore.DeviceLookup(ctx, tc.namespace, tc.hostname)
			assert.Equal(t, tc.expected, Expected{dev: dev, err: err})
		})
	}
}

func TestDeviceList(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	type Expected struct {
		dev []models.Device
        len int
		err error
	}
	cases := []struct {
		description string
		setup       func() error
		expected    Expected
	}{
		{
			description: "succeeds when no devices are found",
			setup: func() error {
                return nil
			},
			expected: Expected{
				dev: []models.Device{
				},
                len: 0,
				err: nil,
			},
		},
		{
			description: "succeeds when a device is found",
			setup: func() error {
				return mongotest.UseFixture(fixtures.NamespaceUnlimited, fixtures.DeviceAccepted)
			},
			expected: Expected{
				dev: []models.Device{
					{
						UID:              "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
						Name:             "hostname_accepted",
						Identity:         &models.DeviceIdentity{MAC: "mac"},
						Info:             nil,
						PublicKey:        "",
						TenantID:         "00000000-0000-4000-0000-000000000000",
						LastSeen:         time.Date(2023, 9, 12, 20, 24, 14, 0, time.UTC),
						Online:           true,
                        Namespace:        "namespace_unlimited",
						Status:           "accepted",
						StatusUpdatedAt:  time.Date(2023, 9, 12, 20, 24, 16, 0, time.UTC),
						CreatedAt:        time.Date(2023, 9, 12, 20, 24, 16, 0, time.UTC),
						RemoteAddr:       "",
						Position:         nil,
						Tags:             []string{},
						PublicURL:        false,
						PublicURLAddress: "",
						Acceptable:       false,
					},
				},
                len: 1,
				err: nil,
			},
		},
		{
			description: "succeeds when two or more devices are found",
			setup: func() error {
				return mongotest.UseFixture(fixtures.NamespaceUnlimited, fixtures.DevicePending, fixtures.DeviceAccepted)
			},
			expected: Expected{
				dev: []models.Device{
					{
						UID:              "3400340f4db3f748747c5d136e2346279025976ec6315c7e226487ddf99019d",
						Name:             "hostname_pending",
						Identity:         &models.DeviceIdentity{MAC: "mac"},
						Info:             nil,
						PublicKey:        "",
						TenantID:         "00000000-0000-4000-0000-000000000000",
						LastSeen:         time.Date(2023, 9, 12, 20, 24, 14, 0, time.UTC),
						Online:           false,
                        Namespace:        "namespace_unlimited",
						Status:           "pending",
						StatusUpdatedAt:  time.Date(2023, 9, 12, 20, 24, 16, 0, time.UTC),
						CreatedAt:        time.Date(2023, 9, 12, 20, 24, 16, 0, time.UTC),
						RemoteAddr:       "",
						Position:         nil,
						Tags:             []string{},
						PublicURL:        false,
						PublicURLAddress: "",
						Acceptable:       false,
					},
					{
						UID:              "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
						Name:             "hostname_accepted",
						Identity:         &models.DeviceIdentity{MAC: "mac"},
						Info:             nil,
						PublicKey:        "",
						TenantID:         "00000000-0000-4000-0000-000000000000",
						LastSeen:         time.Date(2023, 9, 12, 20, 24, 14, 0, time.UTC),
						Online:           true,
                        Namespace:        "namespace_unlimited",
						Status:           "accepted",
						StatusUpdatedAt:  time.Date(2023, 9, 12, 20, 24, 16, 0, time.UTC),
						CreatedAt:        time.Date(2023, 9, 12, 20, 24, 16, 0, time.UTC),
						RemoteAddr:       "",
						Position:         nil,
						Tags:             []string{},
						PublicURL:        false,
						PublicURLAddress: "",
						Acceptable:       false,
					},
				},
                len: 2,
				err: nil,
			},
		},
	}

	for _, tc := range cases {
        t.Run(tc.description, func(t *testing.T) {
            err := tc.setup()
            assert.NoError(t, err)

            dev, count, err := mongostore.DeviceList(ctx, paginator.Query{Page: -1, PerPage: -1}, nil, "", "last_seen", "asc", store.DeviceListModeDefault)
            assert.Equal(t, tc.expected, Expected{dev: dev, len: count, err: err})
        })
	}
}

func TestDeviceListByUsage(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	type Expected struct {
		uid []models.UID
		len int
		err error
	}
	cases := []struct {
		description string
		tenant      string
		setup       func() error
		expected    Expected
	}{
		{
			description: "returns an empty list when tenant not exist",
			tenant:      "nonexistent",
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceSessions)
			},
			expected: Expected{
				uid: []models.UID{},
				len: 0,
				err: nil,
			},
		},
		{
			description: "succeeds when has 1 or more device sessions",
			tenant:      "00000000-0000-4000-0000-000000000000",
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceSessions)
			},
			expected: Expected{
				uid: []models.UID{"uid1", "uid2", "uid3"},
				len: 3,
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			uids, err := mongostore.DeviceListByUsage(ctx, tc.tenant)
			assert.Equal(t, tc.expected, Expected{uid: uids, len: len(uids), err: err})
		})
	}
}

func TestDeviceRename(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description string
		uid         models.UID
		hostname    string
		setup       func() error
		expected    error
	}{
		{
			description: "fails when the device is not found",
			uid:         models.UID("nonexistent"),
			hostname:    "new_hostname",
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceAccepted)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when the device is found",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			hostname:    "new_hostname",
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceAccepted)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.DeviceRename(ctx, tc.uid, tc.hostname)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestDeviceUpdateStatus(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description string
		uid         models.UID
		status      string
		setup       func() error
		expected    error
	}{
		{
			description: "fails when the device is not found",
			uid:         models.UID("nonexistent"),
			status:      "accepted",
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceAccepted)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when the device is found",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			status:      "accepted",
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceAccepted)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.DeviceUpdateStatus(ctx, tc.uid, models.DeviceStatus(tc.status))
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestDeviceUpdateOnline(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description string
		uid         models.UID
		online      bool
		setup       func() error
		expected    error
	}{
		{
			description: "fails when the device is not found",
			uid:         models.UID("nonexistent"),
			online:      true,
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceAccepted)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when the device is found",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			online:      true,
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceAccepted)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.DeviceUpdateOnline(ctx, tc.uid, tc.online)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestDeviceUpdateLastSeen(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description string
		uid         models.UID
		now         time.Time
		setup       func() error
		expected    error
	}{
		{
			description: "fails when the device is not found",
			uid:         models.UID("nonexistent"),
			now:         time.Now(),
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceAccepted)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when the device is found",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			now:         time.Now(),
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceAccepted)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.DeviceUpdateLastSeen(ctx, tc.uid, tc.now)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestDeviceSetOnline(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description string
		uid         models.UID
		online      bool
		setup       func() error
		expected    error
	}{
		{
			description: "succeeds when UID is valid and online is true",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			online:      true,
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceAccepted)
			},
			expected: nil,
		},
		{
			description: "succeeds when UID is valid and online is false",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			online:      false,
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceAccepted)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.DeviceSetOnline(ctx, tc.uid, tc.online)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestDeviceSetPosition(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description string
		uid         models.UID
		position    models.DevicePosition
		setup       func() error
		expected    error
	}{
		{
			description: "fails when the device is not found",
			uid:         models.UID("nonexistent"),
			position: models.DevicePosition{
				Longitude: 1,
				Latitude:  1,
			},
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceAccepted)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when the device is found",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			position: models.DevicePosition{
				Longitude: 1,
				Latitude:  1,
			},
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceAccepted)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.DeviceSetPosition(ctx, tc.uid, tc.position)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestDeviceChooser(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description string
		tenant      string
		chosen      []string
		setup       func() error
		expected    error
	}{
		{
			description: "",
			tenant:      "00000000-0000-4000-0000-000000000000",
			chosen:      []string{"nonexistent"},
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceAccepted)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "",
			tenant:      "00000000-0000-4000-0000-000000000000",
			chosen:      []string{""},
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceAccepted)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.DeviceChooser(ctx, tc.tenant, tc.chosen)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestDeviceDelete(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Configure(&db)

	cases := []struct {
		description string
		uid         models.UID
		setup       func() error
		expected    error
	}{
		{
			description: "fails when device is not found",
			uid:         models.UID("nonexistent"),
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceAccepted)
			},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when device is found",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			setup: func() error {
				return mongotest.UseFixture(fixtures.DeviceAccepted)
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			err := tc.setup()
			assert.NoError(t, err)

			err = mongostore.DeviceDelete(ctx, tc.uid)
			assert.Equal(t, tc.expected, err)
		})
	}
}
