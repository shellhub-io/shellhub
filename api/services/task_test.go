package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	storemock "github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/clock"
	clockmock "github.com/shellhub-io/shellhub/pkg/clock/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_DevicesHeartbeat(t *testing.T) {
	storeMock := storemock.NewMockStore(t)
	clockMock := clockmock.NewMockClock(t)

	prevClock := clock.DefaultBackend
	t.Cleanup(func() { clock.DefaultBackend = prevClock })
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
						"DeviceHeartbeat",
						ctx,
						[]string{"0000000000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000000001"},
						now,
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
						"DeviceHeartbeat",
						ctx,
						[]string{"0000000000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000000001"},
						now,
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
						"DeviceHeartbeat",
						ctx,
						[]string{"0000000000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000000001"},
						now,
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

	storeMock := storemock.NewMockStore(t)
	clockMock := clockmock.NewMockClock(t)

	prevClock := clock.DefaultBackend
	t.Cleanup(func() { clock.DefaultBackend = prevClock })
	clock.DefaultBackend = clockMock

	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	clockMock.On("Now").Return(now).Maybe()

	queryOptionsMock := storemock.NewMockQueryOptions(t)
	storeMock.On("Options").Return(queryOptionsMock)

	thirtyDaysAgo := now.AddDate(0, 0, -30)
	sorter := query.Sorter{By: "removed_at", Order: query.OrderAsc, Tiebreak: "id"}

	// countOpts matches the single-option DeviceList call used to retrieve the total
	// count (only a Match option; no Sort or Paginate).
	countOpts := mock.MatchedBy(func(opts []store.QueryOption) bool {
		return len(opts) == 1
	})

	// pageOpts matches the three-option DeviceList call used to retrieve a page
	// (Match + Sort + Paginate).
	pageOpts := mock.MatchedBy(func(opts []store.QueryOption) bool {
		return len(opts) == 3
	})

	matchFilter := func() func(*query.Filters) bool {
		return func(filters *query.Filters) bool {
			if len(filters.Data) != 1 {
				return false
			}

			matchTime := func() bool {
				filter := filters.Data[0]
				if filter.Type != query.FilterTypeProperty {
					return false
				}

				params, ok := filter.Params.(*query.FilterProperty)
				if !ok {
					return false
				}

				if params.Name != "removed_at" || params.Operator != "lt" {
					return false
				}

				timeValue, isTime := params.Value.(time.Time)
				if !isTime {
					return false
				}

				return timeValue.Equal(thirtyDaysAgo)
			}

			return matchTime()
		}
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
					On("DeviceList", ctx, store.DeviceAcceptableAsFalse, countOpts).
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
					On("DeviceList", ctx, store.DeviceAcceptableAsFalse, countOpts).
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
					On("DeviceList", ctx, store.DeviceAcceptableAsFalse, countOpts).
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
					On("Paginate", &query.Paginator{Page: 1, PerPage: 1000}).
					Return(nil).
					Once()
				storeMock.
					On("DeviceList", ctx, store.DeviceAcceptableAsFalse, pageOpts).
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
					On("DeviceList", ctx, store.DeviceAcceptableAsFalse, countOpts).
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
					On("Paginate", &query.Paginator{Page: 1, PerPage: 1000}).
					Return(nil).
					Once()
				storeMock.
					On("DeviceList", ctx, store.DeviceAcceptableAsFalse, pageOpts).
					Return(
						[]models.Device{
							{UID: "device-1", TenantID: "tenant-1", RemovedAt: &thirtyDaysAgo},
							{UID: "device-2", TenantID: "tenant-1", RemovedAt: &thirtyDaysAgo},
						},
						2,
						nil,
					).
					Once()
				storeMock.
					On("DeviceDeleteMany", ctx, []string{"device-1", "device-2"}).
					Return(int64(0), errors.New("delete error")).
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
					On("DeviceList", ctx, store.DeviceAcceptableAsFalse, countOpts).
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
					On("Paginate", &query.Paginator{Page: 1, PerPage: 1000}).
					Return(nil).
					Once()
				storeMock.
					On("DeviceList", ctx, store.DeviceAcceptableAsFalse, pageOpts).
					Return(
						[]models.Device{
							{UID: "device-1", TenantID: "tenant-1", RemovedAt: &thirtyDaysAgo},
							{UID: "device-2", TenantID: "tenant-1", RemovedAt: &thirtyDaysAgo},
							{UID: "device-3", TenantID: "tenant-2", RemovedAt: &thirtyDaysAgo},
						},
						3,
						nil,
					).
					Once()
				storeMock.
					On("DeviceDeleteMany", ctx, []string{"device-1", "device-2", "device-3"}).
					Return(int64(3), nil).
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
					On("DeviceList", ctx, store.DeviceAcceptableAsFalse, countOpts).
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
					On("Paginate", &query.Paginator{Page: 1, PerPage: 1000}).
					Return(nil).
					Once()
				storeMock.
					On("DeviceList", ctx, store.DeviceAcceptableAsFalse, pageOpts).
					Return(
						[]models.Device{
							{UID: "device-1", TenantID: "tenant-1", RemovedAt: &thirtyDaysAgo},
							{UID: "device-2", TenantID: "tenant-1", RemovedAt: &thirtyDaysAgo},
							{UID: "device-3", TenantID: "tenant-2", RemovedAt: &thirtyDaysAgo},
						},
						3,
						nil,
					).
					Once()
				storeMock.
					On("DeviceDeleteMany", ctx, []string{"device-1", "device-2", "device-3"}).
					Return(int64(3), nil).
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
					On("DeviceList", ctx, store.DeviceAcceptableAsFalse, countOpts).
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
					On("Paginate", &query.Paginator{Page: 1, PerPage: 1000}).
					Return(nil).
					Once()
				storeMock.
					On("DeviceList", ctx, store.DeviceAcceptableAsFalse, pageOpts).
					Return(
						[]models.Device{
							{UID: "device-1", TenantID: "tenant-1", RemovedAt: &thirtyDaysAgo},
						},
						2001,
						nil,
					).
					Once()
				storeMock.
					On("DeviceDeleteMany", ctx, []string{"device-1"}).
					Return(int64(1), nil).
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
					On("DeviceList", ctx, store.DeviceAcceptableAsFalse, pageOpts).
					Return(
						[]models.Device{
							{UID: "device-2", TenantID: "tenant-2", RemovedAt: &thirtyDaysAgo},
						},
						2001,
						nil,
					).
					Once()
				storeMock.
					On("DeviceDeleteMany", ctx, []string{"device-2"}).
					Return(int64(1), nil).
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
					On("Paginate", &query.Paginator{Page: 3, PerPage: 1000}).
					Return(nil).
					Once()
				storeMock.
					On("DeviceList", ctx, store.DeviceAcceptableAsFalse, pageOpts).
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

func TestService_NamespaceDeviceCountSync(t *testing.T) {
	storeMock := storemock.NewMockStore(t)

	cases := []struct {
		description   string
		requiredMocks func(context.Context)
		expected      error
	}{
		{
			description: "fails when sync fails",
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceSyncDeviceCounts", ctx).
					Return(errors.New("sync error")).
					Once()
			},
			expected: errors.New("sync error"),
		},
		{
			description: "succeeds",
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceSyncDeviceCounts", ctx).
					Return(nil).
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
			require.Equal(tt, tc.expected, s.NamespaceDeviceCountSync()(ctx))
		})
	}
}
