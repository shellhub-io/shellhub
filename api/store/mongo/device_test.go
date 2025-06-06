package mongo_test

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

func TestDeviceList(t *testing.T) {
	type Expected struct {
		dev []models.Device
		len int
		err error
	}
	cases := []struct {
		description string
		paginator   query.Paginator
		sorter      query.Sorter
		filters     query.Filters
		status      models.DeviceStatus
		fixtures    []string
		expected    Expected
	}{
		{
			description: "succeeds when no devices are found",
			sorter:      query.Sorter{By: "last_seen", Order: query.OrderAsc},
			paginator:   query.Paginator{Page: -1, PerPage: -1},
			filters:     query.Filters{},
			status:      models.DeviceStatus(""),
			fixtures:    []string{},
			expected: Expected{
				dev: []models.Device{},
				len: 0,
				err: nil,
			},
		},
		{
			description: "succeeds when devices are found",
			sorter:      query.Sorter{By: "last_seen", Order: query.OrderAsc},
			paginator:   query.Paginator{Page: -1, PerPage: -1},
			filters:     query.Filters{},
			status:      models.DeviceStatus(""),
			fixtures:    []string{fixtureNamespaces, fixtureDevices},
			expected: Expected{
				dev: []models.Device{
					{
						CreatedAt:       time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						LastSeen:        time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						UID:             "5300530e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809f",
						Name:            "device-1",
						Identity:        &models.DeviceIdentity{MAC: "mac-1"},
						Info:            nil,
						PublicKey:       "",
						TenantID:        "00000000-0000-4000-0000-000000000000",
						Online:          false,
						Namespace:       "namespace-1",
						Status:          "accepted",
						RemoteAddr:      "",
						Position:        nil,
						Tags:            []string{"tag-1"},
						Acceptable:      false,
					},
					{
						CreatedAt:       time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						LastSeen:        time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						UID:             "4300430e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809e",
						Name:            "device-2",
						Identity:        &models.DeviceIdentity{MAC: "mac-2"},
						Info:            nil,
						PublicKey:       "",
						TenantID:        "00000000-0000-4000-0000-000000000000",
						Online:          false,
						Namespace:       "namespace-1",
						Status:          "accepted",
						RemoteAddr:      "",
						Position:        nil,
						Tags:            []string{},
						Acceptable:      false,
					},
					{
						CreatedAt:       time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt: time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						LastSeen:        time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						UID:             "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
						Name:            "device-3",
						Identity:        &models.DeviceIdentity{MAC: "mac-3"},
						Info:            nil,
						PublicKey:       "",
						TenantID:        "00000000-0000-4000-0000-000000000000",
						Online:          false,
						Namespace:       "namespace-1",
						Status:          "accepted",
						RemoteAddr:      "",
						Position:        nil,
						Tags:            []string{"tag-1"},
						Acceptable:      false,
					},
					{
						CreatedAt:       time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt: time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						LastSeen:        time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						UID:             "3300330e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809d",
						Name:            "device-4",
						Identity:        &models.DeviceIdentity{MAC: "mac-4"},
						Info:            nil,
						PublicKey:       "",
						TenantID:        "00000000-0000-4000-0000-000000000000",
						Online:          false,
						Namespace:       "namespace-1",
						Status:          "pending",
						RemoteAddr:      "",
						Position:        nil,
						Tags:            []string{},
						Acceptable:      true,
					},
				},
				len: 4,
				err: nil,
			},
		},
		{
			description: "succeeds when devices are found with limited page and page size",
			sorter:      query.Sorter{By: "last_seen", Order: query.OrderAsc},
			paginator:   query.Paginator{Page: 2, PerPage: 2},
			filters:     query.Filters{},
			status:      models.DeviceStatus(""),
			fixtures:    []string{fixtureNamespaces, fixtureDevices},
			expected: Expected{
				dev: []models.Device{
					{
						CreatedAt:       time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt: time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						LastSeen:        time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						UID:             "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
						Name:            "device-3",
						Identity:        &models.DeviceIdentity{MAC: "mac-3"},
						Info:            nil,
						PublicKey:       "",
						TenantID:        "00000000-0000-4000-0000-000000000000",
						Online:          false,
						Namespace:       "namespace-1",
						Status:          "accepted",
						RemoteAddr:      "",
						Position:        nil,
						Tags:            []string{"tag-1"},
						Acceptable:      false,
					},
					{
						CreatedAt:       time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt: time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						LastSeen:        time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						UID:             "3300330e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809d",
						Name:            "device-4",
						Identity:        &models.DeviceIdentity{MAC: "mac-4"},
						Info:            nil,
						PublicKey:       "",
						TenantID:        "00000000-0000-4000-0000-000000000000",
						Online:          false,
						Namespace:       "namespace-1",
						Status:          "pending",
						RemoteAddr:      "",
						Position:        nil,
						Tags:            []string{},
						Acceptable:      true,
					},
				},
				len: 4,
				err: nil,
			},
		},
		{
			description: "succeeds when devices are found with sort created_at",
			sorter:      query.Sorter{By: "last_seen", Order: query.OrderAsc},
			paginator:   query.Paginator{Page: -1, PerPage: -1},
			filters:     query.Filters{},
			status:      models.DeviceStatus(""),
			fixtures:    []string{fixtureNamespaces, fixtureDevices},
			expected: Expected{
				dev: []models.Device{
					{
						CreatedAt:       time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						LastSeen:        time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						UID:             "5300530e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809f",
						Name:            "device-1",
						Identity:        &models.DeviceIdentity{MAC: "mac-1"},
						Info:            nil,
						PublicKey:       "",
						TenantID:        "00000000-0000-4000-0000-000000000000",
						Online:          false,
						Namespace:       "namespace-1",
						Status:          "accepted",
						RemoteAddr:      "",
						Position:        nil,
						Tags:            []string{"tag-1"},
						Acceptable:      false,
					},
					{
						CreatedAt:       time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						LastSeen:        time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						UID:             "4300430e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809e",
						Name:            "device-2",
						Identity:        &models.DeviceIdentity{MAC: "mac-2"},
						Info:            nil,
						PublicKey:       "",
						TenantID:        "00000000-0000-4000-0000-000000000000",
						Online:          false,
						Namespace:       "namespace-1",
						Status:          "accepted",
						RemoteAddr:      "",
						Position:        nil,
						Tags:            []string{},
						Acceptable:      false,
					},
					{
						CreatedAt:       time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt: time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						LastSeen:        time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						UID:             "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
						Name:            "device-3",
						Identity:        &models.DeviceIdentity{MAC: "mac-3"},
						Info:            nil,
						PublicKey:       "",
						TenantID:        "00000000-0000-4000-0000-000000000000",
						Online:          false,
						Namespace:       "namespace-1",
						Status:          "accepted",
						RemoteAddr:      "",
						Position:        nil,
						Tags:            []string{"tag-1"},
						Acceptable:      false,
					},
					{
						CreatedAt:       time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt: time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						LastSeen:        time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						UID:             "3300330e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809d",
						Name:            "device-4",
						Identity:        &models.DeviceIdentity{MAC: "mac-4"},
						Info:            nil,
						PublicKey:       "",
						TenantID:        "00000000-0000-4000-0000-000000000000",
						Online:          false,
						Namespace:       "namespace-1",
						Status:          "pending",
						RemoteAddr:      "",
						Position:        nil,
						Tags:            []string{},
						Acceptable:      true,
					},
				},
				len: 4,
				err: nil,
			},
		},
		{
			description: "succeeds when devices are found with order asc",
			sorter:      query.Sorter{By: "last_seen", Order: query.OrderAsc},
			paginator:   query.Paginator{Page: -1, PerPage: -1},
			filters:     query.Filters{},
			status:      models.DeviceStatus(""),
			fixtures:    []string{fixtureNamespaces, fixtureDevices},
			expected: Expected{
				dev: []models.Device{
					{
						CreatedAt:       time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						LastSeen:        time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						UID:             "5300530e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809f",
						Name:            "device-1",
						Identity:        &models.DeviceIdentity{MAC: "mac-1"},
						Info:            nil,
						PublicKey:       "",
						TenantID:        "00000000-0000-4000-0000-000000000000",
						Online:          false,
						Namespace:       "namespace-1",
						Status:          "accepted",
						RemoteAddr:      "",
						Position:        nil,
						Tags:            []string{"tag-1"},
						Acceptable:      false,
					},
					{
						CreatedAt:       time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						LastSeen:        time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						UID:             "4300430e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809e",
						Name:            "device-2",
						Identity:        &models.DeviceIdentity{MAC: "mac-2"},
						Info:            nil,
						PublicKey:       "",
						TenantID:        "00000000-0000-4000-0000-000000000000",
						Online:          false,
						Namespace:       "namespace-1",
						Status:          "accepted",
						RemoteAddr:      "",
						Position:        nil,
						Tags:            []string{},
						Acceptable:      false,
					},
					{
						CreatedAt:       time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt: time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						LastSeen:        time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						UID:             "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
						Name:            "device-3",
						Identity:        &models.DeviceIdentity{MAC: "mac-3"},
						Info:            nil,
						PublicKey:       "",
						TenantID:        "00000000-0000-4000-0000-000000000000",
						Online:          false,
						Namespace:       "namespace-1",
						Status:          "accepted",
						RemoteAddr:      "",
						Position:        nil,
						Tags:            []string{"tag-1"},
						Acceptable:      false,
					},
					{
						CreatedAt:       time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt: time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						LastSeen:        time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						UID:             "3300330e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809d",
						Name:            "device-4",
						Identity:        &models.DeviceIdentity{MAC: "mac-4"},
						Info:            nil,
						PublicKey:       "",
						TenantID:        "00000000-0000-4000-0000-000000000000",
						Online:          false,
						Namespace:       "namespace-1",
						Status:          "pending",
						RemoteAddr:      "",
						Position:        nil,
						Tags:            []string{},
						Acceptable:      true,
					},
				},
				len: 4,
				err: nil,
			},
		},
		{
			description: "succeeds when devices are found with order desc",
			sorter:      query.Sorter{By: "last_seen", Order: query.OrderDesc},
			paginator:   query.Paginator{Page: -1, PerPage: -1},
			filters:     query.Filters{},
			status:      models.DeviceStatus(""),
			fixtures:    []string{fixtureNamespaces, fixtureDevices},
			expected: Expected{
				dev: []models.Device{
					{
						CreatedAt:       time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt: time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						LastSeen:        time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						UID:             "3300330e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809d",
						Name:            "device-4",
						Identity:        &models.DeviceIdentity{MAC: "mac-4"},
						Info:            nil,
						PublicKey:       "",
						TenantID:        "00000000-0000-4000-0000-000000000000",
						Online:          false,
						Namespace:       "namespace-1",
						Status:          "pending",
						RemoteAddr:      "",
						Position:        nil,
						Tags:            []string{},
						Acceptable:      true,
					},
					{
						CreatedAt:       time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt: time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						LastSeen:        time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						UID:             "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
						Name:            "device-3",
						Identity:        &models.DeviceIdentity{MAC: "mac-3"},
						Info:            nil,
						PublicKey:       "",
						TenantID:        "00000000-0000-4000-0000-000000000000",
						Online:          false,
						Namespace:       "namespace-1",
						Status:          "accepted",
						RemoteAddr:      "",
						Position:        nil,
						Tags:            []string{"tag-1"},
						Acceptable:      false,
					},
					{
						CreatedAt:       time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						LastSeen:        time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						UID:             "4300430e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809e",
						Name:            "device-2",
						Identity:        &models.DeviceIdentity{MAC: "mac-2"},
						Info:            nil,
						PublicKey:       "",
						TenantID:        "00000000-0000-4000-0000-000000000000",
						Online:          false,
						Namespace:       "namespace-1",
						Status:          "accepted",
						RemoteAddr:      "",
						Position:        nil,
						Tags:            []string{},
						Acceptable:      false,
					},
					{
						CreatedAt:       time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						LastSeen:        time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						UID:             "5300530e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809f",
						Name:            "device-1",
						Identity:        &models.DeviceIdentity{MAC: "mac-1"},
						Info:            nil,
						PublicKey:       "",
						TenantID:        "00000000-0000-4000-0000-000000000000",
						Online:          false,
						Namespace:       "namespace-1",
						Status:          "accepted",
						RemoteAddr:      "",
						Position:        nil,
						Tags:            []string{"tag-1"},
						Acceptable:      false,
					},
				},
				len: 4,
				err: nil,
			},
		},
		{
			description: "succeeds when devices are found filtering status",
			sorter:      query.Sorter{By: "last_seen", Order: query.OrderAsc},
			paginator:   query.Paginator{Page: -1, PerPage: -1},
			filters:     query.Filters{},
			status:      models.DeviceStatusPending,
			fixtures:    []string{fixtureNamespaces, fixtureDevices},
			expected: Expected{
				dev: []models.Device{
					{
						CreatedAt:       time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt: time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						LastSeen:        time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						UID:             "3300330e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809d",
						Name:            "device-4",
						Identity:        &models.DeviceIdentity{MAC: "mac-4"},
						Info:            nil,
						PublicKey:       "",
						TenantID:        "00000000-0000-4000-0000-000000000000",
						Online:          false,
						Namespace:       "namespace-1",
						Status:          "pending",
						RemoteAddr:      "",
						Position:        nil,
						Tags:            []string{},
						Acceptable:      true,
					},
				},
				len: 1,
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

			dev, count, err := s.DeviceList(
				ctx,
				tc.status,
				tc.paginator,
				tc.filters,
				tc.sorter,
				store.DeviceAcceptableIfNotAccepted,
			)
			assert.Equal(t, tc.expected, Expected{dev: dev, len: count, err: err})
		})
	}
}

func TestDeviceListByUsage(t *testing.T) {
	type Expected struct {
		uid []models.UID
		len int
		err error
	}
	cases := []struct {
		description string
		tenant      string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "returns an empty list when tenant not exist",
			tenant:      "nonexistent",
			fixtures:    []string{fixtureSessions},
			expected: Expected{
				uid: []models.UID{},
				len: 0,
				err: nil,
			},
		},
		{
			description: "succeeds when has 1 or more device sessions",
			tenant:      "00000000-0000-4000-0000-000000000000",
			fixtures:    []string{fixtureSessions},
			expected: Expected{
				uid: []models.UID{"2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"},
				len: 1,
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

			uids, err := s.DeviceListByUsage(ctx, tc.tenant)
			assert.Equal(t, tc.expected, Expected{uid: uids, len: len(uids), err: err})
		})
	}
}

func TestDeviceResolve(t *testing.T) {
	type Expected struct {
		dev *models.Device
		err error
	}

	cases := []struct {
		description string
		tenantID    string
		resolver    store.DeviceResolver
		value       string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when device not found by UID",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			resolver:    store.DeviceUIDResolver,
			value:       "nonexistent",
			fixtures:    []string{fixtureDevices},
			expected: Expected{
				dev: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "fails when tenantID is incorrect",
			tenantID:    "invalid-tenant",
			resolver:    store.DeviceUIDResolver,
			value:       "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
			fixtures:    []string{fixtureDevices},
			expected: Expected{
				dev: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "fails when namespace does not exist",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			resolver:    store.DeviceUIDResolver,
			value:       "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
			fixtures:    []string{fixtureDevices},
			expected: Expected{
				dev: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds resolving device by UID",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			resolver:    store.DeviceUIDResolver,
			value:       "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
			fixtures:    []string{fixtureNamespaces, fixtureDevices},
			expected: Expected{
				dev: &models.Device{
					CreatedAt:       time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
					StatusUpdatedAt: time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
					LastSeen:        time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
					UID:             "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
					Name:            "device-3",
					Identity:        &models.DeviceIdentity{MAC: "mac-3"},
					Info:            nil,
					PublicKey:       "",
					TenantID:        "00000000-0000-4000-0000-000000000000",
					Online:          false,
					Status:          "accepted",
					RemoteAddr:      "",
					Position:        nil,
					Tags:            []string{"tag-1"},
					Acceptable:      false,
					Namespace:       "namespace-1",
				},
				err: nil,
			},
		},
		{
			description: "succeeds resolving device by hostname",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			resolver:    store.DeviceHostnameResolver,
			value:       "device-3",
			fixtures:    []string{fixtureNamespaces, fixtureDevices},
			expected: Expected{
				dev: &models.Device{
					CreatedAt:       time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
					StatusUpdatedAt: time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
					LastSeen:        time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
					UID:             "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
					Name:            "device-3",
					Identity:        &models.DeviceIdentity{MAC: "mac-3"},
					Info:            nil,
					PublicKey:       "",
					TenantID:        "00000000-0000-4000-0000-000000000000",
					Online:          false,
					Status:          "accepted",
					RemoteAddr:      "",
					Position:        nil,
					Tags:            []string{"tag-1"},
					Acceptable:      false,
					Namespace:       "namespace-1",
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() { assert.NoError(t, srv.Reset()) })

			dev, err := s.DeviceResolve(context.Background(), tc.tenantID, tc.resolver, tc.value)
			assert.Equal(t, tc.expected, Expected{dev: dev, err: err})
		})
	}
}

func TestDeviceGet(t *testing.T) {
	type Expected struct {
		dev *models.Device
		err error
	}
	cases := []struct {
		description string
		uid         models.UID
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when namespace is not found",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			fixtures:    []string{fixtureDevices},
			expected: Expected{
				dev: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "fails when device is not found",
			uid:         models.UID("nonexistent"),
			fixtures:    []string{fixtureNamespaces, fixtureDevices},
			expected: Expected{
				dev: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "fails when device is not found due to tenant",
			uid:         models.UID("5600560h6ed5h960969e7f358g4568491247198ge8537e9g448609fff1b231f"),
			fixtures:    []string{fixtureNamespaces, fixtureDevices},
			expected: Expected{
				dev: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when device is found",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			fixtures:    []string{fixtureNamespaces, fixtureDevices},
			expected: Expected{
				dev: &models.Device{
					CreatedAt:       time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
					StatusUpdatedAt: time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
					LastSeen:        time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
					UID:             "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
					Name:            "device-3",
					Identity:        &models.DeviceIdentity{MAC: "mac-3"},
					Info:            nil,
					PublicKey:       "",
					TenantID:        "00000000-0000-4000-0000-000000000000",
					Online:          false,
					Namespace:       "namespace-1",
					Status:          "accepted",
					RemoteAddr:      "",
					Position:        nil,
					Tags:            []string{"tag-1"},
					Acceptable:      false,
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

			dev, err := s.DeviceGet(ctx, tc.uid)
			assert.Equal(t, tc.expected, Expected{dev: dev, err: err})
		})
	}
}

func TestDeviceGetByMac(t *testing.T) {
	type Expected struct {
		dev *models.Device
		err error
	}
	cases := []struct {
		description string
		mac         string
		tenant      string
		status      models.DeviceStatus
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when device is not found due to mac",
			mac:         "nonexistent",
			tenant:      "00000000-0000-4000-0000-000000000000",
			status:      models.DeviceStatus(""),
			fixtures:    []string{fixtureDevices},
			expected: Expected{
				dev: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "fails when device is not found due to tenant",
			mac:         "mac-3",
			tenant:      "nonexistent",
			status:      models.DeviceStatus(""),
			fixtures:    []string{fixtureDevices},
			expected: Expected{
				dev: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when device is found",
			mac:         "mac-3",
			tenant:      "00000000-0000-4000-0000-000000000000",
			status:      models.DeviceStatus(""),
			fixtures:    []string{fixtureDevices},
			expected: Expected{
				dev: &models.Device{
					CreatedAt:       time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
					StatusUpdatedAt: time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
					LastSeen:        time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
					UID:             "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
					Name:            "device-3",
					Identity:        &models.DeviceIdentity{MAC: "mac-3"},
					Info:            nil,
					PublicKey:       "",
					TenantID:        "00000000-0000-4000-0000-000000000000",
					Online:          false,
					Status:          "accepted",
					RemoteAddr:      "",
					Position:        nil,
					Tags:            []string{"tag-1"},
					Acceptable:      false,
				},
				err: nil,
			},
		},
		{
			description: "succeeds when device with status is found",
			mac:         "mac-3",
			tenant:      "00000000-0000-4000-0000-000000000000",
			status:      models.DeviceStatus("accepted"),
			fixtures:    []string{fixtureDevices},
			expected: Expected{
				dev: &models.Device{
					CreatedAt:       time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
					StatusUpdatedAt: time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
					LastSeen:        time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
					UID:             "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
					Name:            "device-3",
					Identity:        &models.DeviceIdentity{MAC: "mac-3"},
					Info:            nil,
					PublicKey:       "",
					TenantID:        "00000000-0000-4000-0000-000000000000",
					Online:          false,
					Status:          "accepted",
					RemoteAddr:      "",
					Position:        nil,
					Tags:            []string{"tag-1"},
					Acceptable:      false,
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

			dev, err := s.DeviceGetByMac(ctx, tc.mac, tc.tenant, tc.status)
			assert.Equal(t, tc.expected, Expected{dev: dev, err: err})
		})
	}
}

func TestDeviceLookup(t *testing.T) {
	type Expected struct {
		dev *models.Device
		err error
	}
	cases := []struct {
		description string
		namespace   string
		hostname    string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when namespace does not exist",
			namespace:   "nonexistent",
			hostname:    "device-3",
			fixtures:    []string{fixtureNamespaces, fixtureDevices},
			expected: Expected{
				dev: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "fails when device does not exist due to name",
			namespace:   "namespace-1",
			hostname:    "nonexistent",
			fixtures:    []string{fixtureNamespaces, fixtureDevices},
			expected: Expected{
				dev: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "fails when device does not exist due to tenant-id",
			namespace:   "namespace-1",
			hostname:    "invalid_tenant",
			fixtures:    []string{fixtureNamespaces, fixtureDevices},
			expected: Expected{
				dev: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "fails when device does not exist due to status other than accepted",
			namespace:   "namespace-1",
			hostname:    "pending",
			fixtures:    []string{fixtureNamespaces, fixtureDevices},
			expected: Expected{
				dev: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when namespace exists and hostname status is accepted",
			namespace:   "namespace-1",
			hostname:    "device-3",
			fixtures:    []string{fixtureNamespaces, fixtureDevices},
			expected: Expected{
				dev: &models.Device{
					CreatedAt:       time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
					StatusUpdatedAt: time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
					LastSeen:        time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
					UID:             "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
					Name:            "device-3",
					Identity:        &models.DeviceIdentity{MAC: "mac-3"},
					Info:            nil,
					PublicKey:       "",
					TenantID:        "00000000-0000-4000-0000-000000000000",
					Online:          false,
					Status:          "accepted",
					RemoteAddr:      "",
					Position:        nil,
					Tags:            []string{"tag-1"},
					Acceptable:      false,
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

			dev, err := s.DeviceLookup(ctx, tc.namespace, tc.hostname)
			assert.Equal(t, tc.expected, Expected{dev: dev, err: err})
		})
	}
}

func TestDeviceCreate(t *testing.T) {
	cases := []struct {
		description string
		hostname    string
		device      models.Device
		fixtures    []string
		expected    error
	}{
		{
			description: "succeeds when all data is valid",
			hostname:    "device-3",
			device: models.Device{
				UID: "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
				Identity: &models.DeviceIdentity{
					MAC: "mac-3",
				},
				TenantID: "00000000-0000-4000-0000-000000000000",
				LastSeen: clock.Now(),
			},
			fixtures: []string{},
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

			err := s.DeviceCreate(ctx, tc.device, tc.hostname)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestDeviceRename(t *testing.T) {
	cases := []struct {
		description string
		uid         models.UID
		hostname    string
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when the device is not found",
			uid:         models.UID("nonexistent"),
			hostname:    "new_hostname",
			fixtures:    []string{fixtureDevices},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when the device is found",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			hostname:    "new_hostname",
			fixtures:    []string{fixtureDevices},
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

			err := s.DeviceRename(ctx, tc.uid, tc.hostname)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestDeviceUpdateStatus(t *testing.T) {
	cases := []struct {
		description string
		uid         models.UID
		status      string
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when the device is not found",
			uid:         models.UID("nonexistent"),
			status:      "accepted",
			fixtures:    []string{fixtureDevices},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when the device is found",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			status:      "accepted",
			fixtures:    []string{fixtureDevices},
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

			err := s.DeviceUpdateStatus(ctx, tc.uid, models.DeviceStatus(tc.status))
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestDeviceSetPosition(t *testing.T) {
	cases := []struct {
		description string
		uid         models.UID
		position    models.DevicePosition
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when the device is not found",
			uid:         models.UID("nonexistent"),
			position: models.DevicePosition{
				Longitude: 1,
				Latitude:  1,
			},
			fixtures: []string{fixtureDevices},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when the device is found",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			position: models.DevicePosition{
				Longitude: 1,
				Latitude:  1,
			},
			fixtures: []string{fixtureDevices},
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

			err := s.DeviceSetPosition(ctx, tc.uid, tc.position)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestDeviceChooser(t *testing.T) {
	cases := []struct {
		description string
		tenant      string
		chosen      []string
		fixtures    []string
		expected    error
	}{
		{
			description: "",
			tenant:      "00000000-0000-4000-0000-000000000000",
			chosen:      []string{""},
			fixtures:    []string{fixtureDevices},
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

			err := s.DeviceChooser(ctx, tc.tenant, tc.chosen)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestDeviceConflicts(t *testing.T) {
	type Expected struct {
		conflicts []string
		ok        bool
		err       error
	}

	cases := []struct {
		description string
		target      *models.DeviceConflicts
		fixtures    []string
		expected    Expected
	}{
		{
			description: "no conflicts when target is empty",
			target:      &models.DeviceConflicts{},
			fixtures:    []string{fixtureDevices},
			expected:    Expected{[]string{}, false, nil},
		},
		{
			description: "no conflicts with non existing email",
			target:      &models.DeviceConflicts{Name: "nonexistent"},
			fixtures:    []string{fixtureDevices},
			expected:    Expected{[]string{}, false, nil},
		},
		{
			description: "conflict detected with existing email",
			target:      &models.DeviceConflicts{Name: "device-1"},
			fixtures:    []string{fixtureDevices},
			expected:    Expected{[]string{"name"}, true, nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			require.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() { require.NoError(t, srv.Reset()) })

			conflicts, ok, err := s.DeviceConflicts(ctx, tc.target)
			require.Equal(t, tc.expected, Expected{conflicts, ok, err})
		})
	}
}

func TestDeviceUpdate(t *testing.T) {
	cases := []struct {
		description string
		tenantID    string
		uid         string
		changes     *models.DeviceChanges
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when the device is not found due to uid",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			uid:         "nonexistent",
			changes:     &models.DeviceChanges{},
			fixtures:    []string{fixtureDevices},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "fails when the device is not found due to tenantID",
			tenantID:    "nonexistent",
			uid:         "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
			changes:     &models.DeviceChanges{},
			fixtures:    []string{fixtureDevices},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when the device is found with tenant",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			uid:         "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
			changes:     &models.DeviceChanges{},
			fixtures:    []string{fixtureDevices},
			expected:    nil,
		},
		{
			description: "succeeds when the device is found without tenant",
			tenantID:    "",
			uid:         "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
			changes:     &models.DeviceChanges{},
			fixtures:    []string{fixtureDevices},
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

			err := s.DeviceUpdate(ctx, tc.tenantID, tc.uid, tc.changes)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestDeviceBulkUpdate(t *testing.T) {
	type Expected struct {
		modifiedCount int64
		err           error
	}

	cases := []struct {
		description string
		uids        []string
		changes     *models.DeviceChanges
		fixtures    []string
		expected    Expected
	}{
		{
			description: "succeeds when a device does not matches",
			uids:        []string{"0000000000000000000000000000000000000000000000000000000000000000"},
			changes:     &models.DeviceChanges{},
			fixtures:    []string{fixtureDevices},
			expected:    Expected{int64(0), nil},
		},
		{
			description: "succeeds when devices matches but nothing is updated",
			uids:        []string{"2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c", "4300430e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809e"},
			changes:     &models.DeviceChanges{},
			fixtures:    []string{fixtureDevices},
			expected:    Expected{int64(0), nil},
		},
		{
			description: "succeeds when devices matches",
			uids:        []string{"2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c", "4300430e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809e"},
			changes:     &models.DeviceChanges{LastSeen: time.Now()},
			fixtures:    []string{fixtureDevices},
			expected:    Expected{int64(2), nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			modifiedCount, err := s.DeviceBulkUpdate(ctx, tc.uids, tc.changes)
			require.Equal(t, Expected{modifiedCount: modifiedCount, err: err}, tc.expected)

			cursor, err := db.Collection("devices").Find(ctx, bson.M{"uid": bson.M{"$in": tc.uids}})
			require.NoError(t, err)

			for cursor.Next(ctx) {
				device := new(models.Device)
				require.NoError(t, cursor.Decode(device))

				if tc.changes.LastSeen != (time.Time{}) {
					require.WithinDuration(t, tc.changes.LastSeen, device.LastSeen, 2*time.Second)
				}
			}
		})
	}
}

func TestDeviceDelete(t *testing.T) {
	cases := []struct {
		description string
		uid         models.UID
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when device is not found",
			uid:         models.UID("nonexistent"),
			fixtures:    []string{fixtureDevices},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when device is found",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			fixtures:    []string{fixtureDevices},
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

			err := s.DeviceDelete(ctx, tc.uid)
			assert.Equal(t, tc.expected, err)
		})
	}
}
