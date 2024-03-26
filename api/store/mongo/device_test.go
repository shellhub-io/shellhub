package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
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
			fixtures:    []string{fixtures.FixtureNamespaces, fixtures.FixtureDevices, fixtures.FixtureConnectedDevices},
			expected: Expected{
				dev: []models.Device{
					{
						CreatedAt:        time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt:  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						LastSeen:         time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						UID:              "5300530e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809f",
						Name:             "device-1",
						Identity:         &models.DeviceIdentity{MAC: "mac-1"},
						Info:             nil,
						PublicKey:        "",
						TenantID:         "00000000-0000-4000-0000-000000000000",
						Online:           false,
						Namespace:        "namespace-1",
						Status:           "accepted",
						RemoteAddr:       "",
						Position:         nil,
						Tags:             []string{"tag-1"},
						PublicURL:        false,
						PublicURLAddress: "",
						Acceptable:       false,
					},
					{
						CreatedAt:        time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt:  time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						LastSeen:         time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						UID:              "4300430e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809e",
						Name:             "device-2",
						Identity:         &models.DeviceIdentity{MAC: "mac-2"},
						Info:             nil,
						PublicKey:        "",
						TenantID:         "00000000-0000-4000-0000-000000000000",
						Online:           false,
						Namespace:        "namespace-1",
						Status:           "accepted",
						RemoteAddr:       "",
						Position:         nil,
						Tags:             []string{},
						PublicURL:        false,
						PublicURLAddress: "",
						Acceptable:       false,
					},
					{
						CreatedAt:        time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt:  time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						LastSeen:         time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						UID:              "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
						Name:             "device-3",
						Identity:         &models.DeviceIdentity{MAC: "mac-3"},
						Info:             nil,
						PublicKey:        "",
						TenantID:         "00000000-0000-4000-0000-000000000000",
						Online:           true,
						Namespace:        "namespace-1",
						Status:           "accepted",
						RemoteAddr:       "",
						Position:         nil,
						Tags:             []string{"tag-1"},
						PublicURL:        false,
						PublicURLAddress: "",
						Acceptable:       false,
					},
					{
						CreatedAt:        time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt:  time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						LastSeen:         time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						UID:              "3300330e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809d",
						Name:             "device-4",
						Identity:         &models.DeviceIdentity{MAC: "mac-4"},
						Info:             nil,
						PublicKey:        "",
						TenantID:         "00000000-0000-4000-0000-000000000000",
						Online:           false,
						Namespace:        "namespace-1",
						Status:           "pending",
						RemoteAddr:       "",
						Position:         nil,
						Tags:             []string{},
						PublicURL:        false,
						PublicURLAddress: "",
						Acceptable:       true,
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
			fixtures:    []string{fixtures.FixtureNamespaces, fixtures.FixtureDevices, fixtures.FixtureConnectedDevices},
			expected: Expected{
				dev: []models.Device{
					{
						CreatedAt:        time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt:  time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						LastSeen:         time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						UID:              "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
						Name:             "device-3",
						Identity:         &models.DeviceIdentity{MAC: "mac-3"},
						Info:             nil,
						PublicKey:        "",
						TenantID:         "00000000-0000-4000-0000-000000000000",
						Online:           true,
						Namespace:        "namespace-1",
						Status:           "accepted",
						RemoteAddr:       "",
						Position:         nil,
						Tags:             []string{"tag-1"},
						PublicURL:        false,
						PublicURLAddress: "",
						Acceptable:       false,
					},
					{
						CreatedAt:        time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt:  time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						LastSeen:         time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						UID:              "3300330e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809d",
						Name:             "device-4",
						Identity:         &models.DeviceIdentity{MAC: "mac-4"},
						Info:             nil,
						PublicKey:        "",
						TenantID:         "00000000-0000-4000-0000-000000000000",
						Online:           false,
						Namespace:        "namespace-1",
						Status:           "pending",
						RemoteAddr:       "",
						Position:         nil,
						Tags:             []string{},
						PublicURL:        false,
						PublicURLAddress: "",
						Acceptable:       true,
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
			fixtures:    []string{fixtures.FixtureNamespaces, fixtures.FixtureDevices, fixtures.FixtureConnectedDevices},
			expected: Expected{
				dev: []models.Device{
					{
						CreatedAt:        time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt:  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						LastSeen:         time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						UID:              "5300530e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809f",
						Name:             "device-1",
						Identity:         &models.DeviceIdentity{MAC: "mac-1"},
						Info:             nil,
						PublicKey:        "",
						TenantID:         "00000000-0000-4000-0000-000000000000",
						Online:           false,
						Namespace:        "namespace-1",
						Status:           "accepted",
						RemoteAddr:       "",
						Position:         nil,
						Tags:             []string{"tag-1"},
						PublicURL:        false,
						PublicURLAddress: "",
						Acceptable:       false,
					},
					{
						CreatedAt:        time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt:  time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						LastSeen:         time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						UID:              "4300430e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809e",
						Name:             "device-2",
						Identity:         &models.DeviceIdentity{MAC: "mac-2"},
						Info:             nil,
						PublicKey:        "",
						TenantID:         "00000000-0000-4000-0000-000000000000",
						Online:           false,
						Namespace:        "namespace-1",
						Status:           "accepted",
						RemoteAddr:       "",
						Position:         nil,
						Tags:             []string{},
						PublicURL:        false,
						PublicURLAddress: "",
						Acceptable:       false,
					},
					{
						CreatedAt:        time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt:  time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						LastSeen:         time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						UID:              "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
						Name:             "device-3",
						Identity:         &models.DeviceIdentity{MAC: "mac-3"},
						Info:             nil,
						PublicKey:        "",
						TenantID:         "00000000-0000-4000-0000-000000000000",
						Online:           true,
						Namespace:        "namespace-1",
						Status:           "accepted",
						RemoteAddr:       "",
						Position:         nil,
						Tags:             []string{"tag-1"},
						PublicURL:        false,
						PublicURLAddress: "",
						Acceptable:       false,
					},
					{
						CreatedAt:        time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt:  time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						LastSeen:         time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						UID:              "3300330e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809d",
						Name:             "device-4",
						Identity:         &models.DeviceIdentity{MAC: "mac-4"},
						Info:             nil,
						PublicKey:        "",
						TenantID:         "00000000-0000-4000-0000-000000000000",
						Online:           false,
						Namespace:        "namespace-1",
						Status:           "pending",
						RemoteAddr:       "",
						Position:         nil,
						Tags:             []string{},
						PublicURL:        false,
						PublicURLAddress: "",
						Acceptable:       true,
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
			fixtures:    []string{fixtures.FixtureNamespaces, fixtures.FixtureDevices, fixtures.FixtureConnectedDevices},
			expected: Expected{
				dev: []models.Device{
					{
						CreatedAt:        time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt:  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						LastSeen:         time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						UID:              "5300530e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809f",
						Name:             "device-1",
						Identity:         &models.DeviceIdentity{MAC: "mac-1"},
						Info:             nil,
						PublicKey:        "",
						TenantID:         "00000000-0000-4000-0000-000000000000",
						Online:           false,
						Namespace:        "namespace-1",
						Status:           "accepted",
						RemoteAddr:       "",
						Position:         nil,
						Tags:             []string{"tag-1"},
						PublicURL:        false,
						PublicURLAddress: "",
						Acceptable:       false,
					},
					{
						CreatedAt:        time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt:  time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						LastSeen:         time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						UID:              "4300430e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809e",
						Name:             "device-2",
						Identity:         &models.DeviceIdentity{MAC: "mac-2"},
						Info:             nil,
						PublicKey:        "",
						TenantID:         "00000000-0000-4000-0000-000000000000",
						Online:           false,
						Namespace:        "namespace-1",
						Status:           "accepted",
						RemoteAddr:       "",
						Position:         nil,
						Tags:             []string{},
						PublicURL:        false,
						PublicURLAddress: "",
						Acceptable:       false,
					},
					{
						CreatedAt:        time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt:  time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						LastSeen:         time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						UID:              "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
						Name:             "device-3",
						Identity:         &models.DeviceIdentity{MAC: "mac-3"},
						Info:             nil,
						PublicKey:        "",
						TenantID:         "00000000-0000-4000-0000-000000000000",
						Online:           true,
						Namespace:        "namespace-1",
						Status:           "accepted",
						RemoteAddr:       "",
						Position:         nil,
						Tags:             []string{"tag-1"},
						PublicURL:        false,
						PublicURLAddress: "",
						Acceptable:       false,
					},
					{
						CreatedAt:        time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt:  time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						LastSeen:         time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						UID:              "3300330e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809d",
						Name:             "device-4",
						Identity:         &models.DeviceIdentity{MAC: "mac-4"},
						Info:             nil,
						PublicKey:        "",
						TenantID:         "00000000-0000-4000-0000-000000000000",
						Online:           false,
						Namespace:        "namespace-1",
						Status:           "pending",
						RemoteAddr:       "",
						Position:         nil,
						Tags:             []string{},
						PublicURL:        false,
						PublicURLAddress: "",
						Acceptable:       true,
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
			fixtures:    []string{fixtures.FixtureNamespaces, fixtures.FixtureDevices, fixtures.FixtureConnectedDevices},
			expected: Expected{
				dev: []models.Device{
					{
						CreatedAt:        time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt:  time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						LastSeen:         time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						UID:              "3300330e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809d",
						Name:             "device-4",
						Identity:         &models.DeviceIdentity{MAC: "mac-4"},
						Info:             nil,
						PublicKey:        "",
						TenantID:         "00000000-0000-4000-0000-000000000000",
						Online:           false,
						Namespace:        "namespace-1",
						Status:           "pending",
						RemoteAddr:       "",
						Position:         nil,
						Tags:             []string{},
						PublicURL:        false,
						PublicURLAddress: "",
						Acceptable:       true,
					},
					{
						CreatedAt:        time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt:  time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						LastSeen:         time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
						UID:              "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
						Name:             "device-3",
						Identity:         &models.DeviceIdentity{MAC: "mac-3"},
						Info:             nil,
						PublicKey:        "",
						TenantID:         "00000000-0000-4000-0000-000000000000",
						Online:           true,
						Namespace:        "namespace-1",
						Status:           "accepted",
						RemoteAddr:       "",
						Position:         nil,
						Tags:             []string{"tag-1"},
						PublicURL:        false,
						PublicURLAddress: "",
						Acceptable:       false,
					},
					{
						CreatedAt:        time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt:  time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						LastSeen:         time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
						UID:              "4300430e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809e",
						Name:             "device-2",
						Identity:         &models.DeviceIdentity{MAC: "mac-2"},
						Info:             nil,
						PublicKey:        "",
						TenantID:         "00000000-0000-4000-0000-000000000000",
						Online:           false,
						Namespace:        "namespace-1",
						Status:           "accepted",
						RemoteAddr:       "",
						Position:         nil,
						Tags:             []string{},
						PublicURL:        false,
						PublicURLAddress: "",
						Acceptable:       false,
					},
					{
						CreatedAt:        time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt:  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						LastSeen:         time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						UID:              "5300530e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809f",
						Name:             "device-1",
						Identity:         &models.DeviceIdentity{MAC: "mac-1"},
						Info:             nil,
						PublicKey:        "",
						TenantID:         "00000000-0000-4000-0000-000000000000",
						Online:           false,
						Namespace:        "namespace-1",
						Status:           "accepted",
						RemoteAddr:       "",
						Position:         nil,
						Tags:             []string{"tag-1"},
						PublicURL:        false,
						PublicURLAddress: "",
						Acceptable:       false,
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
			fixtures:    []string{fixtures.FixtureNamespaces, fixtures.FixtureDevices, fixtures.FixtureConnectedDevices},
			expected: Expected{
				dev: []models.Device{
					{
						CreatedAt:        time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						StatusUpdatedAt:  time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						LastSeen:         time.Date(2023, 1, 4, 12, 0, 0, 0, time.UTC),
						UID:              "3300330e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809d",
						Name:             "device-4",
						Identity:         &models.DeviceIdentity{MAC: "mac-4"},
						Info:             nil,
						PublicKey:        "",
						TenantID:         "00000000-0000-4000-0000-000000000000",
						Online:           false,
						Namespace:        "namespace-1",
						Status:           "pending",
						RemoteAddr:       "",
						Position:         nil,
						Tags:             []string{},
						PublicURL:        false,
						PublicURLAddress: "",
						Acceptable:       true,
					},
				},
				len: 1,
				err: nil,
			},
		},
	}

	ctx := context.TODO()

	mongostore := GetMongoStore()

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			// assert.NoError(t, fixtures.Apply(tc.fixtures...))
			// defer fixtures.Teardown() // nolint: errcheck

			if tc.expected.len > 0 {
				mockConnectedDevices := []interface{}{
					bson.M{
						"uid":       "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
						"tenant_id": "00000000-0000-4000-0000-000000000000",
						"status":    "online",
					},
					bson.M{
						"uid":       "3300330e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809d",
						"tenant_id": "00000000-0000-4000-0000-000000000000",
						"status":    "online",
					},
				}
				connectedDevicesCollection := mongostore.db.Collection("connected_devices")
				if err := dbtest.InsertMockData(ctx, connectedDevicesCollection, mockConnectedDevices); err != nil {
					t.Fatalf("failed to insert mock data for connected_devices: %v", err)
				}

				var testData []interface{}
				for _, item := range tc.expected.dev {
					testData = append(testData, item)
				}
				collection := mongostore.db.Collection("devices")

				if err := dbtest.InsertMockData(ctx, collection, testData); err != nil {
					t.Fatalf("failed to insert documents: %v", err)
				}

				mockNamespaces := []interface{}{
					bson.M{
						"name":      "namespace-1",
						"tenant_id": "00000000-0000-4000-0000-000000000",
					},
				}
				namespacesCollection := mongostore.db.Collection("namespaces")
				if err := dbtest.InsertMockData(ctx, namespacesCollection, mockNamespaces); err != nil {
					t.Fatalf("failed to insert mock data for namespaces: %v", err)
				}
			}

			dev, count, err := mongostore.DeviceList(
				context.TODO(),
				tc.status,
				tc.paginator,
				tc.filters,
				tc.sorter,
				store.DeviceAcceptableIfNotAccepted,
			)
			assert.ObjectsAreEqual(tc.expected, Expected{dev: dev, len: count, err: err})

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
			tenant:      "",
			fixtures:    []string{fixtures.FixtureSessions},
			expected: Expected{
				uid: []models.UID{},
				len: 0,
				err: nil,
			},
		},
		{
			description: "succeeds when has 1 or more device sessions",
			tenant:      "00000000-0000-4000-0000-000000000000",
			fixtures:    []string{fixtures.FixtureSessions},
			expected: Expected{
				uid: []models.UID{},
				len: 0,
				err: nil,
			},
		},
	}

	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")
	collection := mongostore.db.Collection("sessions")
	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			if tc.tenant != "" {
				mockData := bson.M{"device_uid": "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"}
				if err := dbtest.InsertMockData(context.TODO(), collection, []interface{}{mockData}); err != nil {
					t.Fatalf("failed to insert device document: %v", err)
				}
			}

			uids, err := mongostore.DeviceListByUsage(context.TODO(), tc.tenant)
			assert.Equal(t, tc.expected, Expected{uid: uids, len: len(uids), err: err})
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
			description: "succeeds when device is found",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			fixtures:    []string{fixtures.FixtureNamespaces, fixtures.FixtureDevices, fixtures.FixtureConnectedDevices},
			expected: Expected{
				dev: &models.Device{
					UID:              "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
					Name:             "device-3",
					Identity:         &models.DeviceIdentity{MAC: "mac-3"},
					TenantID:         "00000000-0000-4000-0000-000000000000",
					Online:           true,
					Namespace:        "namespace-1",
					Status:           models.DeviceStatusAccepted,
					RemoteAddr:       "",
					Position:         nil,
					Tags:             []string{"tag-1"},
					PublicURL:        false,
					PublicURLAddress: "",
					Acceptable:       false,
				},
				err: nil,
			},
		},
	}

	ctx := context.TODO()

	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			if len(tc.expected.dev.Namespace) > 0 {
				mockNamespace := bson.M{
					"name":      tc.expected.dev.Namespace,
					"tenant_id": tc.expected.dev.TenantID,
				}
				collection := mongostore.db.Collection("namespaces")
				if err := dbtest.InsertMockData(ctx, collection, []interface{}{mockNamespace}); err != nil {
					t.Fatalf("failed to insert mock data for namespaces: %v", err)
				}
			}

			if len(tc.expected.dev.UID) > 0 {
				mockConnectedDevice := bson.M{
					"uid":       tc.expected.dev.UID,
					"tenant_id": tc.expected.dev.TenantID,
				}
				collection := mongostore.db.Collection("connected_devices")
				if err := dbtest.InsertMockData(ctx, collection, []interface{}{mockConnectedDevice}); err != nil {
					t.Fatalf("failed to insert mock data for connected_devices: %v", err)
				}
			}

			if tc.uid != "" {
				mockDevice := bson.M{
					"uid":             tc.expected.dev.UID,
					"name":            tc.expected.dev.Name,
					"identity":        bson.M{"mac": tc.expected.dev.Identity.MAC},
					"tenant_id":       tc.expected.dev.TenantID,
					"online":          tc.expected.dev.Online,
					"namespace":       tc.expected.dev.Namespace,
					"status":          string(tc.expected.dev.Status),
					"tags":            tc.expected.dev.Tags,
					"public_url":      tc.expected.dev.PublicURL,
					"public_url_addr": tc.expected.dev.PublicURLAddress,
					"acceptable":      tc.expected.dev.Acceptable,
				}

				collection := mongostore.db.Collection("devices")
				if err := dbtest.InsertMockData(ctx, collection, []interface{}{mockDevice}); err != nil {
					t.Fatalf("failed to insert mock data for device: %v", err)
				}
			}

			dev, err := mongostore.DeviceGet(context.TODO(), tc.uid)
			if err != nil {
				assert.EqualError(t, err, tc.expected.err.Error(), "unexpected error")
				return
			}

			assert.Equal(t, tc.expected.dev, dev, "devices are not equal")
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
			mac:         "",
			tenant:      "00000000-0000-4000-0000-000000000000",
			status:      models.DeviceStatus(""),
			fixtures:    []string{fixtures.FixtureDevices},
			expected: Expected{
				dev: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "fails when device is not found due to tenant",
			mac:         "mac-3",
			tenant:      "",
			status:      models.DeviceStatus(""),
			fixtures:    []string{fixtures.FixtureDevices},
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
			fixtures:    []string{fixtures.FixtureDevices},
			expected: Expected{
				dev: &models.Device{
					UID:              "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
					Name:             "device-3",
					Identity:         &models.DeviceIdentity{MAC: "mac-3"},
					Info:             nil,
					PublicKey:        "",
					TenantID:         "00000000-0000-4000-0000-000000000000",
					Online:           false,
					Status:           "accepted",
					RemoteAddr:       "",
					Position:         nil,
					Tags:             []string{"tag-1"},
					PublicURL:        false,
					PublicURLAddress: "",
					Acceptable:       false,
				},
				err: nil,
			},
		},
		{
			description: "succeeds when device with status is found",
			mac:         "mac-3",
			tenant:      "00000000-0000-4000-0000-000000000000",
			status:      models.DeviceStatus("accepted"),
			fixtures:    []string{fixtures.FixtureDevices},
			expected: Expected{
				dev: &models.Device{
					UID:              "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
					Name:             "device-3",
					Identity:         &models.DeviceIdentity{MAC: "mac-3"},
					Info:             nil,
					PublicKey:        "",
					TenantID:         "00000000-0000-4000-0000-000000000000",
					Online:           false,
					Status:           "accepted",
					RemoteAddr:       "",
					Position:         nil,
					Tags:             []string{"tag-1"},
					PublicURL:        false,
					PublicURLAddress: "",
					Acceptable:       false,
				},
				err: nil,
			},
		},
	}

	ctx := context.TODO()

	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck
			if tc.mac != "" && tc.tenant != "" {
				collection := mongostore.db.Collection("devices")
				mockData := bson.M{
					"uid":        tc.expected.dev.UID,
					"name":       tc.expected.dev.Name,
					"identity":   bson.M{"mac": tc.expected.dev.Identity.MAC},
					"tenant_id":  tc.expected.dev.TenantID,
					"online":     tc.expected.dev.Online,
					"status":     tc.expected.dev.Status,
					"tags":       tc.expected.dev.Tags,
					"public_url": tc.expected.dev.PublicURL,
					"acceptable": tc.expected.dev.Acceptable,
				}

				if err := dbtest.InsertMockData(ctx, collection, []interface{}{mockData}); err != nil {
					t.Fatalf("failed to insert device document: %v", err)
				}
			}
			dev, err := mongostore.DeviceGetByMac(context.TODO(), tc.mac, tc.tenant, tc.status)
			assert.Equal(t, tc.expected, Expected{dev: dev, err: err})
		})
	}
}

func TestDeviceGetByName(t *testing.T) {
	type Expected struct {
		dev *models.Device
		err error
	}
	cases := []struct {
		description string
		hostname    string
		tenant      string
		status      models.DeviceStatus
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when device is not found due to name",
			hostname:    "",
			tenant:      "00000000-0000-4000-0000-000000000000",
			status:      models.DeviceStatusAccepted,
			fixtures:    []string{fixtures.FixtureDevices},
			expected: Expected{
				dev: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "fails when device is not found due to tenant",
			hostname:    "device-3",
			tenant:      "",
			status:      models.DeviceStatusAccepted,
			fixtures:    []string{fixtures.FixtureDevices},
			expected: Expected{
				dev: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when device is found",
			hostname:    "device-3",
			tenant:      "00000000-0000-4000-0000-000000000000",
			status:      models.DeviceStatusAccepted,
			fixtures:    []string{fixtures.FixtureDevices},
			expected: Expected{
				dev: &models.Device{
					UID:              "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
					Name:             "device-3",
					Identity:         &models.DeviceIdentity{MAC: "mac-3"},
					Info:             nil,
					PublicKey:        "",
					TenantID:         "00000000-0000-4000-0000-000000000000",
					Online:           false,
					Status:           "accepted",
					RemoteAddr:       "",
					Position:         nil,
					Tags:             []string{"tag-1"},
					PublicURL:        false,
					PublicURLAddress: "",
					Acceptable:       false,
				},
				err: nil,
			},
		},
	}

	ctx := context.TODO()

	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			if tc.hostname != "" && tc.tenant != "" {
				collection := mongostore.db.Collection("devices")
				mockData := bson.M{
					"uid":        tc.expected.dev.UID,
					"name":       tc.expected.dev.Name,
					"identity":   bson.M{"mac": tc.expected.dev.Identity.MAC},
					"tenant_id":  tc.expected.dev.TenantID,
					"online":     tc.expected.dev.Online,
					"status":     tc.expected.dev.Status,
					"tags":       tc.expected.dev.Tags,
					"public_url": tc.expected.dev.PublicURL,
					"acceptable": tc.expected.dev.Acceptable,
				}

				if err := dbtest.InsertMockData(ctx, collection, []interface{}{mockData}); err != nil {
					t.Fatalf("failed to insert device document: %v", err)
				}
			}

			dev, err := mongostore.DeviceGetByName(context.TODO(), tc.hostname, tc.tenant, tc.status)
			assert.Equal(t, tc.expected, Expected{dev: dev, err: err})
		})
	}
}

func TestDeviceGetByUID(t *testing.T) {
	type Expected struct {
		dev *models.Device
		err error
	}
	cases := []struct {
		description string
		uid         models.UID
		tenant      string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when device is not found due to UID",
			uid:         models.UID(""),
			tenant:      "00000000-0000-4000-0000-000000000000",
			fixtures:    []string{fixtures.FixtureDevices},
			expected: Expected{
				dev: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "fails when device is not found due to tenant",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			tenant:      "",
			fixtures:    []string{fixtures.FixtureDevices},
			expected: Expected{
				dev: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when device is found",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			tenant:      "00000000-0000-4000-0000-000000000000",
			fixtures:    []string{fixtures.FixtureDevices},
			expected: Expected{
				dev: &models.Device{
					UID:              "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
					Name:             "device-3",
					Identity:         &models.DeviceIdentity{MAC: "mac-3"},
					Info:             nil,
					PublicKey:        "",
					TenantID:         "00000000-0000-4000-0000-000000000000",
					Online:           false,
					Status:           "accepted",
					RemoteAddr:       "",
					Position:         nil,
					Tags:             []string{"tag-1"},
					PublicURL:        false,
					PublicURLAddress: "",
					Acceptable:       false,
				},
				err: nil,
			},
		},
	}

	ctx := context.TODO()

	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck
			if tc.uid != "" && tc.tenant != "" {
				collection := mongostore.db.Collection("devices")
				mockData := bson.M{
					"uid":        tc.expected.dev.UID,
					"name":       tc.expected.dev.Name,
					"identity":   bson.M{"mac": tc.expected.dev.Identity.MAC},
					"tenant_id":  tc.expected.dev.TenantID,
					"online":     tc.expected.dev.Online,
					"status":     tc.expected.dev.Status,
					"tags":       tc.expected.dev.Tags,
					"public_url": tc.expected.dev.PublicURL,
					"acceptable": tc.expected.dev.Acceptable,
				}

				if err := dbtest.InsertMockData(ctx, collection, []interface{}{mockData}); err != nil {
					t.Fatalf("failed to insert device document: %v", err)
				}
			}

			dev, err := mongostore.DeviceGetByUID(context.TODO(), tc.uid, tc.tenant)
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
			namespace:   "",
			hostname:    "device-3",
			fixtures:    []string{fixtures.FixtureNamespaces, fixtures.FixtureDevices},
			expected: Expected{
				dev: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "fails when device does not exist due to name",
			namespace:   "namespace-1",
			hostname:    "",
			fixtures:    []string{fixtures.FixtureNamespaces, fixtures.FixtureDevices},
			expected: Expected{
				dev: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "fails when device does not exist due to tenant-id",
			namespace:   "namespace-1",
			hostname:    "invalid_tenant",
			fixtures:    []string{fixtures.FixtureNamespaces, fixtures.FixtureDevices},
			expected: Expected{
				dev: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "fails when device does not exist due to status other than accepted",
			namespace:   "namespace-1",
			hostname:    "pending",
			fixtures:    []string{fixtures.FixtureNamespaces, fixtures.FixtureDevices},
			expected: Expected{
				dev: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when namespace exists and hostname status is accepted",
			namespace:   "namespace-1",
			hostname:    "device-3",
			fixtures:    []string{fixtures.FixtureNamespaces, fixtures.FixtureDevices},
			expected: Expected{
				dev: &models.Device{
					CreatedAt:        time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
					StatusUpdatedAt:  time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
					LastSeen:         time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
					UID:              "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
					Name:             "device-3",
					Identity:         &models.DeviceIdentity{MAC: "mac-3"},
					Info:             nil,
					PublicKey:        "",
					TenantID:         "00000000-0000-4000-0000-000000000000",
					Online:           false,
					Status:           "accepted",
					RemoteAddr:       "",
					Position:         nil,
					Tags:             []string{"tag-1"},
					PublicURL:        false,
					PublicURLAddress: "",
					Acceptable:       false,
				},
				err: nil,
			},
		},
	}

	ctx := context.TODO()

	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			if tc.hostname != "" && tc.namespace != "" {
				collection := mongostore.db.Collection("namespaces")
				mockData := bson.M{"name": tc.namespace, "tenant_id": "tenant"}

				if err := dbtest.InsertMockData(ctx, collection, []interface{}{mockData}); err != nil {
					t.Fatalf("failed to insert device document: %v", err)
				}

				collection = mongostore.db.Collection("devices")
				mockData = bson.M{"tenant_id": "tenant", "name": tc.namespace, "status": "accepted"}

				if err := dbtest.InsertMockData(ctx, collection, []interface{}{mockData}); err != nil {
					t.Fatalf("failed to insert device document: %v", err)
				}
			}
			dev, err := mongostore.DeviceLookup(context.TODO(), tc.namespace, tc.hostname)
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

	ctx := context.TODO()

	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			if tc.device.UID != "" {
				collection := mongostore.db.Collection("devices")
				mockData := bson.M{"uid": tc.device.UID}

				if err := dbtest.InsertMockData(ctx, collection, []interface{}{mockData}); err != nil {
					t.Fatalf("failed to insert device document: %v", err)
				}
			}
			err := mongostore.DeviceCreate(context.TODO(), tc.device, tc.hostname)
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
			uid:         models.UID(""),
			hostname:    "new_hostname",
			fixtures:    []string{fixtures.FixtureDevices},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when the device is found",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			hostname:    "new_hostname",
			fixtures:    []string{fixtures.FixtureDevices},
			expected:    nil,
		},
	}

	ctx := context.TODO()

	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			if tc.uid != "" {
				collection := mongostore.db.Collection("devices")
				mockData := bson.M{"uid": tc.uid}

				if err := dbtest.InsertMockData(ctx, collection, []interface{}{mockData}); err != nil {
					t.Fatalf("failed to insert device document: %v", err)
				}
			}

			err := mongostore.DeviceRename(context.TODO(), tc.uid, tc.hostname)
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
			uid:         models.UID(""),
			status:      "accepted",
			fixtures:    []string{fixtures.FixtureDevices},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when the device is found",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			status:      "accepted",
			fixtures:    []string{fixtures.FixtureDevices},
			expected:    nil,
		},
	}

	ctx := context.TODO()

	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			if tc.uid != "" {
				collection := mongostore.db.Collection("devices")
				mockData := bson.M{"uid": tc.uid}

				if err := dbtest.InsertMockData(ctx, collection, []interface{}{mockData}); err != nil {
					t.Fatalf("failed to insert device document: %v", err)
				}
			}

			err := mongostore.DeviceUpdateStatus(context.TODO(), tc.uid, models.DeviceStatus(tc.status))
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestDeviceUpdateOnline(t *testing.T) {
	cases := []struct {
		description string
		uid         models.UID
		online      bool
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when the device is not found",
			uid:         models.UID(""),
			online:      true,
			fixtures:    []string{fixtures.FixtureDevices},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when the device is found",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			online:      true,
			fixtures:    []string{fixtures.FixtureDevices},
			expected:    nil,
		},
	}

	ctx := context.TODO()

	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			if tc.uid != "" {
				collection := mongostore.db.Collection("devices")
				mockData := bson.M{"uid": tc.uid}

				if err := dbtest.InsertMockData(ctx, collection, []interface{}{mockData}); err != nil {
					t.Fatalf("failed to insert device document: %v", err)
				}
			}

			err := mongostore.DeviceUpdateOnline(context.TODO(), tc.uid, tc.online)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestDeviceUpdateLastSeen(t *testing.T) {
	cases := []struct {
		description string
		uid         models.UID
		now         time.Time
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when the device is not found",
			uid:         models.UID(""),
			now:         time.Now(),
			fixtures:    []string{fixtures.FixtureDevices},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when the device is found",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			now:         time.Now(),
			fixtures:    []string{fixtures.FixtureDevices},
			expected:    nil,
		},
	}

	ctx := context.TODO()

	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck
			if tc.uid != "" {
				collection := mongostore.db.Collection("devices")
				mockData := bson.M{"uid": tc.uid}

				if err := dbtest.InsertMockData(ctx, collection, []interface{}{mockData}); err != nil {
					t.Fatalf("failed to insert device document: %v", err)
				}
			}
			err := mongostore.DeviceUpdateLastSeen(context.TODO(), tc.uid, tc.now)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestDeviceSetOnline(t *testing.T) {
	cases := []struct {
		description string
		uid         models.UID
		online      bool
		fixtures    []string
		expected    error
	}{
		{
			description: "succeeds when UID is valid and online is true",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			online:      true,
			fixtures:    []string{fixtures.FixtureDevices},
			expected:    nil,
		},
		{
			description: "succeeds when UID is valid and online is false",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			online:      false,
			fixtures:    []string{fixtures.FixtureDevices},
			expected:    nil,
		},
	}

	ctx := context.TODO()

	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			collection := mongostore.db.Collection("devices")
			mockData := bson.M{"uid": tc.uid, "online": tc.online}

			if err := dbtest.InsertMockData(ctx, collection, []interface{}{mockData}); err != nil {
				t.Fatalf("failed to insert device document: %v", err)
			}

			err := mongostore.DeviceSetOnline(context.TODO(), tc.uid, time.Now(), tc.online)
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
			uid:         models.UID(""),
			position: models.DevicePosition{
				Longitude: 1,
				Latitude:  1,
			},
			fixtures: []string{fixtures.FixtureDevices},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when the device is found",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			position: models.DevicePosition{
				Longitude: 1,
				Latitude:  1,
			},
			fixtures: []string{fixtures.FixtureDevices},
			expected: nil,
		},
	}

	ctx := context.TODO()

	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			if tc.uid != "" {
				collection := mongostore.db.Collection("devices")
				mockData := bson.M{"uid": tc.uid}

				if err := dbtest.InsertMockData(ctx, collection, []interface{}{mockData}); err != nil {
					t.Fatalf("failed to insert device document: %v", err)
				}
			}

			err := mongostore.DeviceSetPosition(context.TODO(), tc.uid, tc.position)
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
			description: "fails when tenant is nil",
			tenant:      "",
			chosen:      []string{""},
			fixtures:    []string{fixtures.FixtureDevices},
			expected:    nil,
		},
	}

	ctx := context.TODO()

	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			if tc.tenant != "" {
				collection := mongostore.db.Collection("devices")
				mockData := bson.M{"tenant_id": tc.tenant}

				if err := dbtest.InsertMockData(ctx, collection, []interface{}{mockData}); err != nil {
					t.Fatalf("failed to insert device document: %v", err)
				}
			}

			err := mongostore.DeviceChooser(context.TODO(), tc.tenant, tc.chosen)
			assert.Equal(t, tc.expected, err)
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
			uid:         models.UID(""),
			fixtures:    []string{fixtures.FixtureDevices},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when device is found",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			fixtures:    []string{fixtures.FixtureDevices},
			expected:    nil,
		},
	}

	ctx := context.TODO()

	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			if tc.uid != "" {
				collection := mongostore.db.Collection("devices")
				mockData := bson.M{"uid": tc.uid}

				if err := dbtest.InsertMockData(ctx, collection, []interface{}{mockData}); err != nil {
					t.Fatalf("failed to insert device document: %v", err)
				}
			}
			err := mongostore.DeviceDelete(context.TODO(), tc.uid)
			assert.Equal(t, tc.expected, err)
		})
	}
}
