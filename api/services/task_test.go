package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	storemocks "github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/clock"
	clockmock "github.com/shellhub-io/shellhub/pkg/clock/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_DevicesHeartbeat(t *testing.T) {
	storeMock := new(storemocks.Store)
	clockMock := new(clockmock.Clock)

	clock.DefaultBackend = clockMock

	clockMock.On("Now").Return(now)

	cases := []struct {
		description   string
		payload       []byte
		requiredMocks func(context.Context)
		expected      error
	}{
		{
			description: "fails when cannot set the status",
			payload:     []byte("0000000000000000000000000000000000000000000000000000000000000000\n0000000000000000000000000000000000000000000000000000000000000001"),
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On(
						"DeviceBulkUpdate",
						ctx,
						[]string{"0000000000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000000001"},
						&models.DeviceChanges{LastSeen: now, DisconnectedAt: nil},
					).
					Return(int64(0), errors.New("error")).
					Once()
			},
			expected: errors.New("error"),
		},
		{
			description: "succeeds with duplicated IDs",
			payload:     []byte("0000000000000000000000000000000000000000000000000000000000000000\n0000000000000000000000000000000000000000000000000000000000000001\n0000000000000000000000000000000000000000000000000000000000000000"),
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On(
						"DeviceBulkUpdate",
						ctx,
						[]string{"0000000000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000000001"},
						&models.DeviceChanges{LastSeen: now, DisconnectedAt: nil},
					).
					Return(int64(2), nil).
					Once()
			},
			expected: nil,
		},
		{
			description: "succeeds",
			payload:     []byte("0000000000000000000000000000000000000000000000000000000000000000\n0000000000000000000000000000000000000000000000000000000000000001"),
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On(
						"DeviceBulkUpdate",
						ctx,
						[]string{"0000000000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000000001"},
						&models.DeviceChanges{LastSeen: now, DisconnectedAt: nil},
					).
					Return(int64(2), nil).
					Once()
			},
			expected: nil,
		},
	}

	s := NewService(storeMock, privateKey, publicKey, cache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			ctx := context.Background()
			tc.requiredMocks(ctx)
			require.Equal(tt, tc.expected, s.DevicesHeartbeat()(ctx, tc.payload))
		})
	}
}

func TestService_DeviceCleanup(t *testing.T) {
	ctx := context.Background()

	matchFilter := func() func(*query.Filters) bool {
		return func(filters *query.Filters) bool {
			if len(filters.Data) != 2 {
				return false
			}

			matchStatus := func() bool {
				filter := filters.Data[0]
				if filter.Type != query.FilterTypeProperty {
					return false
				}

				params, ok := filter.Params.(*query.FilterProperty)
				if !ok {
					return false
				}

				return params.Name == "status" && params.Operator == "eq" && params.Value == string(models.DeviceStatusRemoved)
			}

			matchTime := func() bool {
				filter := filters.Data[1]
				if filter.Type != query.FilterTypeProperty {
					return false
				}

				params, ok := filter.Params.(*query.FilterProperty)
				if !ok {
					return false
				}

				if params.Name != "status_updated_at" || params.Operator != "lt" {
					return false
				}

				timeValue, isTime := params.Value.(time.Time)
				if !isTime {
					return false
				}

				expectedTime := time.Now().AddDate(0, 0, -30)
				timeDiff := timeValue.Sub(expectedTime)
				if timeDiff < 0 {
					timeDiff = -timeDiff
				}

				return timeDiff <= time.Second // allow 1 seconds tolerance
			}

			return matchStatus() && matchTime()
		}
	}

	storeMock := new(storemocks.Store)
	clockMock := new(clockmock.Clock)

	clock.DefaultBackend = clockMock

	now := time.Now()
	clockMock.On("Now").Return(now)

	queryOptionsMock := new(storemocks.QueryOptions)
	storeMock.On("Options").Return(queryOptionsMock)

	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	sorter := query.Sorter{
		By:    "status_updated_at",
		Order: query.OrderAsc,
	}

	cases := []struct {
		description   string
		requiredMocks func()
		expected      error
	}{
		{
			description: "fails when cannot get total count of removed devices",
			requiredMocks: func() {
				queryOptionsMock.
					On("Match", mock.MatchedBy(matchFilter())).
					Return(nil).
					Once()
				storeMock.
					On("DeviceList", ctx, store.DeviceAcceptableAsFalse, mock.AnythingOfType("store.QueryOption")).
					Return([]models.Device{}, 0, errors.New("database error")).
					Once()
			},
			expected: errors.New("database error"),
		},
		{
			description: "succeeds with no removed devices to cleanup",
			requiredMocks: func() {
				queryOptionsMock.
					On("Match", mock.MatchedBy(matchFilter())).
					Return(nil).
					Once()
				storeMock.
					On("DeviceList", ctx, store.DeviceAcceptableAsFalse, mock.AnythingOfType("store.QueryOption")).
					Return([]models.Device{}, 0, nil).
					Once()
			},
			expected: nil,
		},
		{
			description: "fails when cannot list devices for a page",
			requiredMocks: func() {
				queryOptionsMock.
					On("Match", mock.MatchedBy(matchFilter())).
					Return(nil).
					Once()
				storeMock.
					On("DeviceList", ctx, store.DeviceAcceptableAsFalse, mock.AnythingOfType("store.QueryOption")).
					Return([]models.Device{}, 1000, nil).
					Once()
				queryOptionsMock.
					On("Match", mock.MatchedBy(matchFilter())).
					Return(nil).
					Once()
				queryOptionsMock.
					On("Sort", &sorter).
					Return(nil).
					Once()
				queryOptionsMock.
					On("Paginate", &query.Paginator{Page: 0, PerPage: 1000}).
					Return(nil).
					Once()
				storeMock.
					On("DeviceList", ctx, store.DeviceAcceptableAsFalse, mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).
					Return([]models.Device{}, 0, errors.New("page error")).
					Once()
			},
			expected: errors.New("page error"),
		},
		{
			description: "fails when cannot delete some devices",
			requiredMocks: func() {
				queryOptionsMock.
					On("Match", mock.MatchedBy(matchFilter())).
					Return(nil).
					Once()
				storeMock.
					On("DeviceList", ctx, store.DeviceAcceptableAsFalse, mock.AnythingOfType("store.QueryOption")).
					Return([]models.Device{}, 2, nil).
					Once()
				queryOptionsMock.
					On("Match", mock.MatchedBy(matchFilter())).
					Return(nil).
					Once()
				queryOptionsMock.
					On("Sort", &sorter).
					Return(nil).
					Once()
				queryOptionsMock.
					On("Paginate", &query.Paginator{Page: 0, PerPage: 1000}).
					Return(nil).
					Once()
				storeMock.
					On("DeviceList", ctx, store.DeviceAcceptableAsFalse, mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).
					Return(
						[]models.Device{
							{UID: "device-1", TenantID: "tenant-1", StatusUpdatedAt: thirtyDaysAgo},
							{UID: "device-2", TenantID: "tenant-1", StatusUpdatedAt: thirtyDaysAgo},
						},
						2,
						nil,
					).
					Once()
				storeMock.
					On("DeviceDelete", ctx, models.UID("device-1")).
					Return(nil).
					Once()
				storeMock.
					On("DeviceDelete", ctx, models.UID("device-2")).
					Return(errors.New("delete error")).
					Once()
			},
			expected: errors.New("delete error"),
		},
		{
			description: "fails when cannot update counters",
			requiredMocks: func() {
				queryOptionsMock.
					On("Match", mock.MatchedBy(matchFilter())).
					Return(nil).
					Once()
				storeMock.
					On("DeviceList", ctx, store.DeviceAcceptableAsFalse, mock.AnythingOfType("store.QueryOption")).
					Return([]models.Device{}, 3, nil).
					Once()
				queryOptionsMock.
					On("Match", mock.MatchedBy(matchFilter())).
					Return(nil).
					Once()
				queryOptionsMock.
					On("Sort", &sorter).
					Return(nil).
					Once()
				queryOptionsMock.
					On("Paginate", &query.Paginator{Page: 0, PerPage: 1000}).
					Return(nil).
					Once()
				storeMock.
					On("DeviceList", ctx, store.DeviceAcceptableAsFalse, mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).
					Return(
						[]models.Device{
							{UID: "device-1", TenantID: "tenant-1", StatusUpdatedAt: thirtyDaysAgo},
							{UID: "device-2", TenantID: "tenant-1", StatusUpdatedAt: thirtyDaysAgo},
							{UID: "device-3", TenantID: "tenant-2", StatusUpdatedAt: thirtyDaysAgo},
						},
						3,
						nil,
					).
					Once()
				storeMock.
					On("DeviceDelete", ctx, models.UID("device-1")).
					Return(nil).
					Once()
				storeMock.
					On("DeviceDelete", ctx, models.UID("device-2")).
					Return(nil).
					Once()
				storeMock.
					On("DeviceDelete", ctx, models.UID("device-3")).
					Return(nil).
					Once()
				storeMock.
					On("NamespaceIncrementDeviceCount", ctx, "tenant-1", models.DeviceStatusRemoved, int64(-2)).
					Return(nil).
					Once()
				storeMock.
					On("NamespaceIncrementDeviceCount", ctx, "tenant-2", models.DeviceStatusRemoved, int64(-1)).
					Return(errors.New("update error")).
					Once()
			},
			expected: errors.New("update error"),
		},
		{
			description: "succeeds deleting all old removed devices and updates counters",
			requiredMocks: func() {
				queryOptionsMock.
					On("Match", mock.MatchedBy(matchFilter())).
					Return(nil).
					Once()
				storeMock.
					On("DeviceList", ctx, store.DeviceAcceptableAsFalse, mock.AnythingOfType("store.QueryOption")).
					Return([]models.Device{}, 3, nil).
					Once()
				queryOptionsMock.
					On("Match", mock.MatchedBy(matchFilter())).
					Return(nil).
					Once()
				queryOptionsMock.
					On("Sort", &sorter).
					Return(nil).
					Once()
				queryOptionsMock.
					On("Paginate", &query.Paginator{Page: 0, PerPage: 1000}).
					Return(nil).
					Once()
				storeMock.
					On("DeviceList", ctx, store.DeviceAcceptableAsFalse, mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).
					Return(
						[]models.Device{
							{UID: "device-1", TenantID: "tenant-1", StatusUpdatedAt: thirtyDaysAgo},
							{UID: "device-2", TenantID: "tenant-1", StatusUpdatedAt: thirtyDaysAgo},
							{UID: "device-3", TenantID: "tenant-2", StatusUpdatedAt: thirtyDaysAgo},
						},
						3,
						nil,
					).
					Once()
				storeMock.
					On("DeviceDelete", ctx, models.UID("device-1")).
					Return(nil).
					Once()
				storeMock.
					On("DeviceDelete", ctx, models.UID("device-2")).
					Return(nil).
					Once()
				storeMock.
					On("DeviceDelete", ctx, models.UID("device-3")).
					Return(nil).
					Once()
				storeMock.
					On("NamespaceIncrementDeviceCount", ctx, "tenant-1", models.DeviceStatusRemoved, int64(-2)).
					Return(nil).
					Once()
				storeMock.
					On("NamespaceIncrementDeviceCount", ctx, "tenant-2", models.DeviceStatusRemoved, int64(-1)).
					Return(nil).
					Once()
			},
			expected: nil,
		},
		{
			description: "succeeds with multiple pages and updates counters correctly",
			requiredMocks: func() {
				queryOptionsMock.
					On("Match", mock.MatchedBy(matchFilter())).
					Return(nil).
					Once()
				storeMock.
					On("DeviceList", ctx, store.DeviceAcceptableAsFalse, mock.AnythingOfType("store.QueryOption")).
					Return([]models.Device{}, 2001, nil).
					Once()
				queryOptionsMock.
					On("Match", mock.MatchedBy(matchFilter())).
					Return(nil).
					Once()
				queryOptionsMock.
					On("Sort", &sorter).
					Return(nil).
					Once()
				queryOptionsMock.
					On("Paginate", &query.Paginator{Page: 0, PerPage: 1000}).
					Return(nil).
					Once()
				storeMock.
					On("DeviceList", ctx, store.DeviceAcceptableAsFalse, mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).
					Return(
						[]models.Device{
							{UID: "device-1", TenantID: "tenant-1", StatusUpdatedAt: thirtyDaysAgo},
						},
						2001,
						nil,
					).
					Once()
				storeMock.
					On("DeviceDelete", ctx, models.UID("device-1")).
					Return(nil).
					Once()
				queryOptionsMock.
					On("Match", mock.MatchedBy(matchFilter())).
					Return(nil).
					Once()
				queryOptionsMock.
					On("Sort", &sorter).
					Return(nil).
					Once()
				queryOptionsMock.
					On("Paginate", &query.Paginator{Page: 1, PerPage: 1000}).
					Return(nil).
					Once()
				storeMock.
					On("DeviceList", ctx, store.DeviceAcceptableAsFalse, mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).
					Return(
						[]models.Device{
							{UID: "device-2", TenantID: "tenant-2", StatusUpdatedAt: thirtyDaysAgo},
						},
						2001,
						nil,
					).
					Once()
				storeMock.
					On("DeviceDelete", ctx, models.UID("device-2")).
					Return(nil).
					Once()
				queryOptionsMock.
					On("Match", mock.MatchedBy(matchFilter())).
					Return(nil).
					Once()
				queryOptionsMock.
					On("Sort", &sorter).
					Return(nil).
					Once()
				queryOptionsMock.
					On("Paginate", &query.Paginator{Page: 2, PerPage: 1000}).
					Return(nil).
					Once()
				storeMock.
					On("DeviceList", ctx, store.DeviceAcceptableAsFalse, mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).
					Return([]models.Device{}, 2001, nil).
					Once()
				storeMock.
					On("NamespaceIncrementDeviceCount", ctx, "tenant-1", models.DeviceStatusRemoved, int64(-1)).
					Return(nil).
					Once()
				storeMock.
					On("NamespaceIncrementDeviceCount", ctx, "tenant-2", models.DeviceStatusRemoved, int64(-1)).
					Return(nil).
					Once()
			},
			expected: nil,
		},
	}

	s := NewService(storeMock, privateKey, publicKey, cache.NewNullCache(), clientMock)

	storeMock.
		On("WithTransaction", ctx, mock.AnythingOfType("store.TransactionCb")).
		Return(func(ctx context.Context, cb store.TransactionCb) error { return cb(ctx) }).
		Times(len(cases))

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tc.requiredMocks()
			require.Equal(tt, tc.expected, s.DeviceCleanup()(ctx))
		})
	}
}
