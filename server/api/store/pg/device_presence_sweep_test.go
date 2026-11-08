package pg_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/shellhub-io/shellhub/server/api/store/storetest"
	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeviceSweepStalePresenceRechecksARowWrittenMeanwhile(t *testing.T) {
	ctx := context.Background()

	provider, err := pgprovider.NewProvider(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { _ = provider.Close(t) })

	st := provider.Store()
	suite := storetest.NewSuite(provider)
	now := provider.Clock().Now().UTC().Truncate(time.Second)
	lastSeen := now.Add(-time.Hour)
	staleBefore := now.Add(-10 * time.Minute)
	closedAt := now.Add(-30 * time.Minute)

	cases := []struct {
		description    string
		write          func(ctx context.Context, uid string) error
		disconnectedAt *time.Time
		online         bool
	}{
		{
			description: "a heartbeat",
			write: func(ctx context.Context, uid string) error {
				_, err := st.DeviceHeartbeat(ctx, []store.DeviceBeat{{UID: uid, At: now}}, 0)

				return err
			},
			online: true,
		},
		{
			description: "a written close",
			write: func(ctx context.Context, uid string) error {
				return st.DeviceOffline(ctx, uid, closedAt)
			},
			disconnectedAt: &closedAt,
		},
		{
			description: "another sweep",
			write: func(ctx context.Context, _ string) error {
				_, err := st.DeviceSweepStalePresence(ctx, staleBefore, 1000)

				return err
			},
			disconnectedAt: new(lastSeen.Add(models.DeviceOnlineWindow)),
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			require.NoError(t, provider.CleanDatabase(t))

			uid := string(suite.CreateDevice(t, storetest.WithDeviceLastSeen(lastSeen)))

			written := make(chan error, 1)
			commit := make(chan struct{})
			release := sync.OnceFunc(func() { close(commit) })
			committed := make(chan error, 1)
			writerDone := make(chan struct{})

			defer func() {
				release()
				<-writerDone
			}()

			go func() {
				defer close(writerDone)

				committed <- st.WithTransaction(ctx, func(ctx context.Context) error {
					err := tc.write(ctx, uid)
					written <- err
					<-commit

					return err
				})
			}()

			require.NoError(t, <-written)

			type sweepResult struct {
				stamped int64
				err     error
			}

			swept := make(chan sweepResult, 1)

			go func() {
				stamped, err := st.DeviceSweepStalePresence(ctx, staleBefore, 1000)
				swept <- sweepResult{stamped: stamped, err: err}
			}()

			require.Eventually(t, func() bool {
				var waiting int
				err := provider.DB().NewRaw(
					"SELECT count(*) FROM pg_stat_activity WHERE datname = current_database() AND cardinality(pg_blocking_pids(pid)) > 0",
				).Scan(ctx, &waiting)

				return err == nil && waiting == 1
			}, 10*time.Second, 20*time.Millisecond, "the sweep never waited on the row the other transaction holds")

			release()
			require.NoError(t, <-committed)

			result := <-swept
			require.NoError(t, result.err)
			assert.Equal(t, int64(0), result.stamped)

			device, err := st.DeviceResolve(ctx, scope.NewUnbounded("test: reading back the device the sweep raced"), store.DeviceUIDResolver, uid)
			require.NoError(t, err)
			assert.Equal(t, tc.online, device.Online)

			if tc.disconnectedAt == nil {
				assert.Nil(t, device.DisconnectedAt)
			} else {
				require.NotNil(t, device.DisconnectedAt)
				assert.WithinDuration(t, *tc.disconnectedAt, *device.DisconnectedAt, time.Millisecond)
			}
		})
	}
}
