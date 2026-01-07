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
		opts        []store.QueryOption
		fixtures    []string
		expected    Expected
	}{
		{
			description: "succeeds when no devices are found",
			opts: []store.QueryOption{
				s.Options().Match(&query.Filters{}),
				s.Options().Sort(&query.Sorter{By: "last_seen", Order: query.OrderAsc}),
				s.Options().Paginate(&query.Paginator{Page: -1, PerPage: -1}),
			},
			fixtures: []string{},
			expected: Expected{
				dev: []models.Device{},
				len: 0,
				err: nil,
			},
		},
		{
			description: "succeeds when devices are found",
			opts: []store.QueryOption{
				s.Options().Match(&query.Filters{}),
				s.Options().Sort(&query.Sorter{By: "last_seen", Order: query.OrderAsc}),
				s.Options().Paginate(&query.Paginator{Page: -1, PerPage: -1}),
			},
			fixtures: []string{fixtureNamespaces, fixtureTags, fixtureDevices},
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
						Acceptable:      false,
						Taggable: models.Taggable{
							TagIDs: []string{"6791d3ae04ba86e6d7a0514d"},
							Tags: []models.Tag{
								{
									ID:        "6791d3ae04ba86e6d7a0514d",
									CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
									UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
									Name:      "production",
									TenantID:  "00000000-0000-4000-0000-000000000000",
								},
							},
						},
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
						Acceptable:      false,
						Taggable: models.Taggable{
							TagIDs: []string{},
							Tags:   []models.Tag{},
						},
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
						Acceptable:      false,
						Taggable: models.Taggable{
							TagIDs: []string{"6791d3ae04ba86e6d7a0514d", "6791d3be5a201d874c4c2885"},
							Tags: []models.Tag{
								{
									ID:        "6791d3ae04ba86e6d7a0514d",
									CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
									UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
									Name:      "production",
									TenantID:  "00000000-0000-4000-0000-000000000000",
								},
								{
									ID:        "6791d3be5a201d874c4c2885",
									CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
									UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
									Name:      "development",
									TenantID:  "00000000-0000-4000-0000-000000000000",
								},
							},
						},
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
						Acceptable:      true,
						Taggable: models.Taggable{
							TagIDs: []string{},
							Tags:   []models.Tag{},
						},
					},
				},
				len: 4,
				err: nil,
			},
		},
		{
			description: "succeeds when devices are found with limited page and page size",
			opts: []store.QueryOption{
				s.Options().Match(&query.Filters{}),
				s.Options().Sort(&query.Sorter{By: "last_seen", Order: query.OrderAsc}),
				s.Options().Paginate(&query.Paginator{Page: 2, PerPage: 2}),
			},
			fixtures: []string{fixtureNamespaces, fixtureTags, fixtureDevices},
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
						Acceptable:      false,
						Taggable: models.Taggable{
							TagIDs: []string{"6791d3ae04ba86e6d7a0514d", "6791d3be5a201d874c4c2885"},
							Tags: []models.Tag{
								{
									ID:        "6791d3ae04ba86e6d7a0514d",
									CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
									UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
									Name:      "production",
									TenantID:  "00000000-0000-4000-0000-000000000000",
								},
								{
									ID:        "6791d3be5a201d874c4c2885",
									CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
									UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
									Name:      "development",
									TenantID:  "00000000-0000-4000-0000-000000000000",
								},
							},
						},
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
						Acceptable:      true,
						Taggable: models.Taggable{
							TagIDs: []string{},
							Tags:   []models.Tag{},
						},
					},
				},
				len: 4,
				err: nil,
			},
		},
		{
			description: "succeeds when devices are found with order asc",
			opts: []store.QueryOption{
				s.Options().Match(&query.Filters{}),
				s.Options().Sort(&query.Sorter{By: "last_seen", Order: query.OrderAsc}),
				s.Options().Paginate(&query.Paginator{Page: -1, PerPage: -1}),
			},
			fixtures: []string{fixtureNamespaces, fixtureTags, fixtureDevices},
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
						Acceptable:      false,
						Taggable: models.Taggable{
							TagIDs: []string{"6791d3ae04ba86e6d7a0514d"},
							Tags: []models.Tag{
								{
									ID:        "6791d3ae04ba86e6d7a0514d",
									CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
									UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
									Name:      "production",
									TenantID:  "00000000-0000-4000-0000-000000000000",
								},
							},
						},
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
						Acceptable:      false,
						Taggable: models.Taggable{
							TagIDs: []string{},
							Tags:   []models.Tag{},
						},
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
						Acceptable:      false,
						Taggable: models.Taggable{
							TagIDs: []string{"6791d3ae04ba86e6d7a0514d", "6791d3be5a201d874c4c2885"},
							Tags: []models.Tag{
								{
									ID:        "6791d3ae04ba86e6d7a0514d",
									CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
									UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
									Name:      "production",
									TenantID:  "00000000-0000-4000-0000-000000000000",
								},
								{
									ID:        "6791d3be5a201d874c4c2885",
									CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
									UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
									Name:      "development",
									TenantID:  "00000000-0000-4000-0000-000000000000",
								},
							},
						},
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
						Acceptable:      true,
						Taggable: models.Taggable{
							TagIDs: []string{},
							Tags:   []models.Tag{},
						},
					},
				},
				len: 4,
				err: nil,
			},
		},
		{
			description: "succeeds when devices are found with order desc",
			opts: []store.QueryOption{
				s.Options().Match(&query.Filters{}),
				s.Options().Sort(&query.Sorter{By: "last_seen", Order: query.OrderDesc}),
				s.Options().Paginate(&query.Paginator{Page: -1, PerPage: -1}),
			},
			fixtures: []string{fixtureNamespaces, fixtureTags, fixtureDevices},
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
						Acceptable:      true,
						Taggable: models.Taggable{
							TagIDs: []string{},
							Tags:   []models.Tag{},
						},
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
						Acceptable:      false,
						Taggable: models.Taggable{
							TagIDs: []string{"6791d3ae04ba86e6d7a0514d", "6791d3be5a201d874c4c2885"},
							Tags: []models.Tag{
								{
									ID:        "6791d3ae04ba86e6d7a0514d",
									CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
									UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
									Name:      "production",
									TenantID:  "00000000-0000-4000-0000-000000000000",
								},
								{
									ID:        "6791d3be5a201d874c4c2885",
									CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
									UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
									Name:      "development",
									TenantID:  "00000000-0000-4000-0000-000000000000",
								},
							},
						},
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
						Acceptable:      false,
						Taggable: models.Taggable{
							TagIDs: []string{},
							Tags:   []models.Tag{},
						},
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
						Acceptable:      false,
						Taggable: models.Taggable{
							TagIDs: []string{"6791d3ae04ba86e6d7a0514d"},
							Tags: []models.Tag{
								{
									ID:        "6791d3ae04ba86e6d7a0514d",
									CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
									UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
									Name:      "production",
									TenantID:  "00000000-0000-4000-0000-000000000000",
								},
							},
						},
					},
				},
				len: 4,
				err: nil,
			},
		},
		{
			description: "succeeds when devices are found filtering status",
			opts: []store.QueryOption{
				s.Options().WithDeviceStatus(models.DeviceStatusPending),
				s.Options().Match(&query.Filters{}),
				s.Options().Sort(&query.Sorter{By: "last_seen", Order: query.OrderAsc}),
				s.Options().Paginate(&query.Paginator{Page: -1, PerPage: -1}),
			},
			fixtures: []string{fixtureNamespaces, fixtureTags, fixtureDevices},
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
						Acceptable:      true,
						Taggable: models.Taggable{
							TagIDs: []string{},
							Tags:   []models.Tag{},
						},
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

			dev, count, err := s.DeviceList(ctx, store.DeviceAcceptableIfNotAccepted, tc.opts...)
			assert.Equal(t, tc.expected, Expected{dev: dev, len: count, err: err})
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
		resolver    store.DeviceResolver
		value       string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when device not found by UID",
			resolver:    store.DeviceUIDResolver,
			value:       "nonexistent",
			fixtures:    []string{fixtureDevices},
			expected: Expected{
				dev: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds resolving device by UID",
			resolver:    store.DeviceUIDResolver,
			value:       "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
			fixtures:    []string{fixtureNamespaces, fixtureTags, fixtureDevices},
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
					Namespace:       "namespace-1",
					Acceptable:      false,
					Taggable: models.Taggable{
						TagIDs: []string{"6791d3ae04ba86e6d7a0514d", "6791d3be5a201d874c4c2885"},
						Tags: []models.Tag{
							{
								ID:        "6791d3ae04ba86e6d7a0514d",
								CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
								UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
								Name:      "production",
								TenantID:  "00000000-0000-4000-0000-000000000000",
							},
							{
								ID:        "6791d3be5a201d874c4c2885",
								CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
								UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
								Name:      "development",
								TenantID:  "00000000-0000-4000-0000-000000000000",
							},
						},
					},
				},
				err: nil,
			},
		},
		{
			description: "succeeds resolving device by hostname",
			resolver:    store.DeviceHostnameResolver,
			value:       "device-3",
			fixtures:    []string{fixtureNamespaces, fixtureTags, fixtureDevices},
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
					Namespace:       "namespace-1",
					Acceptable:      false,
					Taggable: models.Taggable{
						TagIDs: []string{"6791d3ae04ba86e6d7a0514d", "6791d3be5a201d874c4c2885"},
						Tags: []models.Tag{
							{
								ID:        "6791d3ae04ba86e6d7a0514d",
								CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
								UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
								Name:      "production",
								TenantID:  "00000000-0000-4000-0000-000000000000",
							},
							{
								ID:        "6791d3be5a201d874c4c2885",
								CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
								UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
								Name:      "development",
								TenantID:  "00000000-0000-4000-0000-000000000000",
							},
						},
					},
				},
				err: nil,
			},
		},
		{
			description: "succeeds resolving device by MAC",
			resolver:    store.DeviceMACResolver,
			value:       "mac-3",
			fixtures:    []string{fixtureNamespaces, fixtureTags, fixtureDevices},
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
					Namespace:       "namespace-1",
					Acceptable:      false,
					Taggable: models.Taggable{
						TagIDs: []string{"6791d3ae04ba86e6d7a0514d", "6791d3be5a201d874c4c2885"},
						Tags: []models.Tag{
							{
								ID:        "6791d3ae04ba86e6d7a0514d",
								CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
								UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
								Name:      "production",
								TenantID:  "00000000-0000-4000-0000-000000000000",
							},
							{
								ID:        "6791d3be5a201d874c4c2885",
								CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
								UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
								Name:      "development",
								TenantID:  "00000000-0000-4000-0000-000000000000",
							},
						},
					},
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() { assert.NoError(t, srv.Reset()) })

			dev, err := s.DeviceResolve(context.Background(), tc.resolver, tc.value)
			assert.Equal(t, tc.expected, Expected{dev: dev, err: err})
		})
	}
}

func TestDeviceCreate(t *testing.T) {
	type Expected struct {
		insertedUID string
		err         error
	}

	cases := []struct {
		description string
		device      *models.Device
		fixtures    []string
		expected    Expected
	}{
		{
			description: "succeeds when creating new device",
			device: &models.Device{
				UID: "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
				Identity: &models.DeviceIdentity{
					MAC: "mac-3",
				},
				TenantID: "00000000-0000-4000-0000-000000000000",
				LastSeen: clock.Now(),
			},
			fixtures: []string{},
			expected: Expected{
				insertedUID: "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
				err:         nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			assert.NoError(tt, srv.Apply(tc.fixtures...))
			tt.Cleanup(func() { assert.NoError(tt, srv.Reset()) })

			insertedUID, err := s.DeviceCreate(context.Background(), tc.device)
			assert.Equal(tt, tc.expected, Expected{insertedUID, err})
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
		{
			description: "no conflict with removed device name",
			target:      &models.DeviceConflicts{Name: "device-removed"},
			fixtures:    []string{fixtureDevicesWithRemoved},
			expected:    Expected{[]string{}, false, nil},
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
		device      *models.Device
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when the device is not found due to uid",
			device: &models.Device{
				UID:      "nonexistent",
				TenantID: "00000000-0000-4000-0000-000000000000",
			},
			fixtures: []string{fixtureDevices},
			expected: store.ErrNoDocuments,
		},
		{
			description: "fails when the device is not found due to tenantID",
			device: &models.Device{
				UID:      "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
				TenantID: "nonexistent",
			},
			fixtures: []string{fixtureDevices},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when the device",
			device: &models.Device{
				UID:      "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
				TenantID: "00000000-0000-4000-0000-000000000000",
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

			err := s.DeviceUpdate(ctx, tc.device)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestDeviceHeartbeat(t *testing.T) {
	type Expected struct {
		modifiedCount int64
		err           error
	}

	cases := []struct {
		description string
		uids        []string
		lastSeen    time.Time
		fixtures    []string
		expected    Expected
	}{
		{
			description: "succeeds when no devices match",
			uids:        []string{"nonexistent1", "nonexistent2"},
			lastSeen:    time.Now(),
			fixtures:    []string{fixtureDevices},
			expected: Expected{
				modifiedCount: 0,
				err:           nil,
			},
		},
		{
			description: "succeeds when devices match",
			uids: []string{
				"2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
				"4300430e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809e",
			},
			lastSeen: time.Now(),
			fixtures: []string{fixtureDevices},
			expected: Expected{
				modifiedCount: 2,
				err:           nil,
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

			modifiedCount, err := s.DeviceHeartbeat(ctx, tc.uids, tc.lastSeen)
			require.Equal(t, Expected{modifiedCount: modifiedCount, err: err}, tc.expected)

			if tc.expected.modifiedCount > 0 {
				cursor, err := db.Collection("devices").Find(ctx, bson.M{"uid": bson.M{"$in": tc.uids}})
				require.NoError(t, err)

				for cursor.Next(ctx) {
					device := new(models.Device)
					require.NoError(t, cursor.Decode(device))
					require.WithinDuration(t, tc.lastSeen, device.LastSeen, 2*time.Second)
					require.Nil(t, device.DisconnectedAt)
				}
			}
		})
	}
}

func TestDeviceDelete(t *testing.T) {
	cases := []struct {
		description string
		device      *models.Device
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when device is not found",
			device: &models.Device{
				UID: "nonexistent",
			},
			fixtures: []string{fixtureDevices},
			expected: store.ErrNoDocuments,
		},
		{
			description: "succeeds when device is found",
			device: &models.Device{
				UID: "2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
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

			err := s.DeviceDelete(ctx, tc.device)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestDeviceDeleteMany(t *testing.T) {
	type Expected struct {
		deletedCount int64
		err          error
	}

	cases := []struct {
		description string
		uids        []string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "succeeds when no devices match",
			uids:        []string{},
			fixtures:    []string{fixtureDevices},
			expected: Expected{
				deletedCount: 0,
				err:          nil,
			},
		},
		{
			description: "succeeds when devices match",
			uids: []string{
				"2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c",
				"4300430e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809e",
			},
			fixtures: []string{fixtureDevices},
			expected: Expected{
				deletedCount: 2,
				err:          nil,
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

			deletedCount, err := s.DeviceDeleteMany(ctx, tc.uids)
			require.Equal(t, tc.expected, Expected{deletedCount, err})
			if tc.expected.deletedCount > 0 {
				for _, uid := range tc.uids {
					count, err := db.Collection("devices").CountDocuments(ctx, bson.M{"uid": uid})
					require.NoError(t, err)
					require.Equal(t, int64(0), count)
				}
			}
		})
	}
}
