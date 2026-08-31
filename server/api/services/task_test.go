package services

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/clock"
	clockmock "github.com/shellhub-io/shellhub/pkg/clock/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	storemock "github.com/shellhub-io/shellhub/server/api/store/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

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
	storeMock.On("Options").Return(queryOptionsMock).Maybe()

	thirtyDaysAgo := now.AddDate(0, 0, -30)
	sorter := query.Sorter{By: "removed_at", Order: query.OrderAsc, Tiebreak: "id"}

	countOpts := mock.MatchedBy(func(opts []store.QueryOption) bool {
		return len(opts) == 1
	})

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
					On("DeviceList", ctx, mock.Anything, store.DeviceAcceptableAsFalse, countOpts).
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
					On("DeviceList", ctx, mock.Anything, store.DeviceAcceptableAsFalse, countOpts).
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
					On("DeviceList", ctx, mock.Anything, store.DeviceAcceptableAsFalse, countOpts).
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
					On("DeviceList", ctx, mock.Anything, store.DeviceAcceptableAsFalse, pageOpts).
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
					On("DeviceList", ctx, mock.Anything, store.DeviceAcceptableAsFalse, countOpts).
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
					On("DeviceList", ctx, mock.Anything, store.DeviceAcceptableAsFalse, pageOpts).
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
					On("DeviceList", ctx, mock.Anything, store.DeviceAcceptableAsFalse, countOpts).
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
					On("DeviceList", ctx, mock.Anything, store.DeviceAcceptableAsFalse, pageOpts).
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
					On("NamespaceIncrementDeviceCount", ctx, scope.MustBounded("tenant-1"), models.DeviceStatusRemoved, int64(-2)).
					Return(nil).
					Once()
				storeMock.
					On("NamespaceIncrementDeviceCount", ctx, scope.MustBounded("tenant-2"), models.DeviceStatusRemoved, int64(-1)).
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
					On("DeviceList", ctx, mock.Anything, store.DeviceAcceptableAsFalse, countOpts).
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
					On("DeviceList", ctx, mock.Anything, store.DeviceAcceptableAsFalse, pageOpts).
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
					On("NamespaceIncrementDeviceCount", ctx, scope.MustBounded("tenant-1"), models.DeviceStatusRemoved, int64(-2)).
					Return(nil).
					Once()
				storeMock.
					On("NamespaceIncrementDeviceCount", ctx, scope.MustBounded("tenant-2"), models.DeviceStatusRemoved, int64(-1)).
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
					On("DeviceList", ctx, mock.Anything, store.DeviceAcceptableAsFalse, countOpts).
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
					On("DeviceList", ctx, mock.Anything, store.DeviceAcceptableAsFalse, pageOpts).
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
					On("DeviceList", ctx, mock.Anything, store.DeviceAcceptableAsFalse, pageOpts).
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
					On("DeviceList", ctx, mock.Anything, store.DeviceAcceptableAsFalse, pageOpts).
					Return([]models.Device{}, 2001, nil).
					Once()
				storeMock.
					On("NamespaceIncrementDeviceCount", ctx, scope.MustBounded("tenant-1"), models.DeviceStatusRemoved, int64(-1)).
					Return(nil).
					Once()
				storeMock.
					On("NamespaceIncrementDeviceCount", ctx, scope.MustBounded("tenant-2"), models.DeviceStatusRemoved, int64(-1)).
					Return(nil).
					Once()
			},
			expected: nil,
		},
	}

	s := NewService(storeMock, privateKey, publicKey, cache.NewNullCache())

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

	s := NewService(storeMock, privateKey, publicKey, cache.NewNullCache())

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			ctx := context.Background()
			tc.requiredMocks(ctx)
			require.Equal(tt, tc.expected, s.NamespaceDeviceCountSync()(ctx))
		})
	}
}

func TestService_SessionCleanup(t *testing.T) {
	now := time.Date(2026, 8, 11, 12, 0, 0, 0, time.UTC)
	retention := 180 * 24 * time.Hour
	cutoff := now.Add(-retention)

	storeMock := storemock.NewMockStore(t)
	clockMock := clockmock.NewMockClock(t)

	prevClock := clock.DefaultBackend
	t.Cleanup(func() { clock.DefaultBackend = prevClock })
	clock.DefaultBackend = clockMock
	clockMock.On("Now").Return(now).Maybe()

	expired := func(n int) []store.ExpiredSession {
		batch := make([]store.ExpiredSession, n)
		for i := range batch {
			batch[i] = store.ExpiredSession{UID: fmt.Sprintf("session-%d", i)}
		}

		return batch
	}

	uids := func(sessions []store.ExpiredSession) []string {
		out := make([]string, len(sessions))
		for i, session := range sessions {
			out[i] = session.UID
		}

		return out
	}

	full := expired(sessionCleanupBatchSize)

	cases := []struct {
		description   string
		retention     time.Duration
		requiredMocks func(context.Context, *mockSessionRecordingPruner)
		withPruner    bool
		expected      error
	}{
		{
			description: "does not prune when retention is not positive",
			retention:   0,
			requiredMocks: func(_ context.Context, _ *mockSessionRecordingPruner) {
			},
			expected: nil,
		},
		{
			description: "fails when listing fails",
			retention:   retention,
			requiredMocks: func(ctx context.Context, _ *mockSessionRecordingPruner) {
				storeMock.
					On("SessionListExpired", ctx, cutoff, sessionCleanupBatchSize).
					Return(nil, errors.New("database error")).
					Once()
			},
			expected: errors.New("database error"),
		},
		{
			description: "stops when nothing is expired",
			retention:   retention,
			requiredMocks: func(ctx context.Context, _ *mockSessionRecordingPruner) {
				storeMock.
					On("SessionListExpired", ctx, cutoff, sessionCleanupBatchSize).
					Return([]store.ExpiredSession{}, nil).
					Once()
			},
			expected: nil,
		},
		{
			description: "fails when deleting fails",
			retention:   retention,
			requiredMocks: func(ctx context.Context, _ *mockSessionRecordingPruner) {
				storeMock.
					On("SessionListExpired", ctx, cutoff, sessionCleanupBatchSize).
					Return(expired(3), nil).
					Once()
				storeMock.
					On("SessionDeleteMany", ctx, uids(expired(3))).
					Return(int64(0), errors.New("database error")).
					Once()
			},
			expected: errors.New("database error"),
		},
		{
			description: "stops after a partial batch",
			retention:   retention,
			requiredMocks: func(ctx context.Context, _ *mockSessionRecordingPruner) {
				batch := expired(sessionCleanupBatchSize - 1)
				storeMock.
					On("SessionListExpired", ctx, cutoff, sessionCleanupBatchSize).
					Return(batch, nil).
					Once()
				storeMock.
					On("SessionDeleteMany", ctx, uids(batch)).
					Return(int64(len(batch)), nil).
					Once()
			},
			expected: nil,
		},
		{
			description: "keeps batching while each batch comes back full",
			retention:   retention,
			requiredMocks: func(ctx context.Context, _ *mockSessionRecordingPruner) {
				storeMock.
					On("SessionListExpired", ctx, cutoff, sessionCleanupBatchSize).
					Return(full, nil).
					Once()
				storeMock.
					On("SessionDeleteMany", ctx, uids(full)).
					Return(int64(sessionCleanupBatchSize), nil).
					Once()
				storeMock.
					On("SessionListExpired", ctx, cutoff, sessionCleanupBatchSize).
					Return([]store.ExpiredSession{}, nil).
					Once()
			},
			expected: nil,
		},
		{
			description: "stops at the per-run cap and leaves the rest for the next run",
			retention:   retention,
			requiredMocks: func(ctx context.Context, _ *mockSessionRecordingPruner) {
				storeMock.
					On("SessionListExpired", ctx, cutoff, sessionCleanupBatchSize).
					Return(full, nil).
					Times(sessionCleanupMaxBatches)
				storeMock.
					On("SessionDeleteMany", ctx, uids(full)).
					Return(int64(sessionCleanupBatchSize), nil).
					Times(sessionCleanupMaxBatches)
			},
			expected: nil,
		},
		{
			description: "does not reach for the bucket when nothing in the batch was recorded",
			retention:   retention,
			withPruner:  true,
			requiredMocks: func(ctx context.Context, _ *mockSessionRecordingPruner) {
				storeMock.
					On("SessionListExpired", ctx, cutoff, sessionCleanupBatchSize).
					Return(expired(3), nil).
					Once()
				storeMock.
					On("SessionDeleteMany", ctx, uids(expired(3))).
					Return(int64(3), nil).
					Once()
			},
			expected: nil,
		},
		{
			description: "purges a recorded session's recording before deleting its row",
			retention:   retention,
			withPruner:  true,
			requiredMocks: func(ctx context.Context, pruner *mockSessionRecordingPruner) {
				batch := []store.ExpiredSession{
					{UID: "plain"},
					{UID: "recorded", Recorded: true},
				}

				storeMock.
					On("SessionListExpired", ctx, cutoff, sessionCleanupBatchSize).
					Return(batch, nil).
					Once()

				purged := false
				pruner.
					On("DeleteRecordings", ctx, []string{"recorded"}).
					Run(func(_ mock.Arguments) { purged = true }).
					Return([]string{"recorded"}, nil).
					Once()
				storeMock.
					On("SessionDeleteMany", ctx, []string{"plain", "recorded"}).
					Run(func(_ mock.Arguments) {
						assert.True(t, purged, "the row must not be deleted before its recording")
					}).
					Return(int64(2), nil).
					Once()
			},
			expected: nil,
		},
		{
			description: "keeps the row of a session whose recording could not be purged",
			retention:   retention,
			withPruner:  true,
			requiredMocks: func(ctx context.Context, pruner *mockSessionRecordingPruner) {
				batch := []store.ExpiredSession{
					{UID: "ok", Recorded: true},
					{UID: "stuck", Recorded: true},
				}

				storeMock.
					On("SessionListExpired", ctx, cutoff, sessionCleanupBatchSize).
					Return(batch, nil).
					Once()
				pruner.
					On("DeleteRecordings", ctx, []string{"ok", "stuck"}).
					Return([]string{"ok"}, nil).
					Once()
				storeMock.
					On("SessionDeleteMany", ctx, []string{"ok"}).
					Return(int64(1), nil).
					Once()
			},
			expected: nil,
		},
		{
			description: "ends the run when no session in the batch can be deleted",
			retention:   retention,
			withPruner:  true,
			requiredMocks: func(ctx context.Context, pruner *mockSessionRecordingPruner) {
				storeMock.
					On("SessionListExpired", ctx, cutoff, sessionCleanupBatchSize).
					Return([]store.ExpiredSession{{UID: "stuck", Recorded: true}}, nil).
					Once()
				pruner.
					On("DeleteRecordings", ctx, []string{"stuck"}).
					Return([]string{}, nil).
					Once()
			},
			expected: nil,
		},
		{
			description: "fails when the pruner reports the batch is moot",
			retention:   retention,
			withPruner:  true,
			requiredMocks: func(ctx context.Context, pruner *mockSessionRecordingPruner) {
				storeMock.
					On("SessionListExpired", ctx, cutoff, sessionCleanupBatchSize).
					Return([]store.ExpiredSession{{UID: "a", Recorded: true}}, nil).
					Once()
				pruner.
					On("DeleteRecordings", ctx, []string{"a"}).
					Return(nil, context.Canceled).
					Once()
			},
			expected: context.Canceled,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			ctx := context.Background()

			pruner := new(mockSessionRecordingPruner)
			tt.Cleanup(func() { pruner.AssertExpectations(tt) })
			tc.requiredMocks(ctx, pruner)

			opts := []Option{}
			if tc.withPruner {
				opts = append(opts, WithSessionRecordingPruner(pruner))
			}

			s := NewService(storeMock, privateKey, publicKey, cache.NewNullCache(), opts...)
			require.Equal(tt, tc.expected, s.sessionCleanup(ctx, tc.retention, 0))
		})
	}

	t.Run("the cron handler prunes nothing when retention is not positive", func(tt *testing.T) {
		s := NewService(storeMock, privateKey, publicKey, cache.NewNullCache())
		require.NoError(tt, s.SessionCleanup(0)(context.Background()))
	})
}

type mockSessionRecordingPruner struct {
	mock.Mock
}

func (m *mockSessionRecordingPruner) DeleteRecordings(ctx context.Context, uids []string) ([]string, error) {
	args := m.Called(ctx, uids)

	purged, _ := args.Get(0).([]string)

	return purged, args.Error(1)
}
