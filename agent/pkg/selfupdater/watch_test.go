package selfupdater

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Masterminds/semver/v3"
	log "github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type serverVersions struct {
	mu      sync.Mutex
	version *semver.Version
	checks  int
}

func (s *serverVersions) check() (*semver.Version, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.checks++

	return s.version, nil
}

func (s *serverVersions) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.checks
}

type recordingUpdater struct {
	mu      sync.Mutex
	applied []string
	err     error
}

func (r *recordingUpdater) CurrentVersion() (*semver.Version, error) {
	return semver.NewVersion("0.20.0")
}

func (r *recordingUpdater) ApplyUpdate(v *semver.Version) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.applied = append(r.applied, v.String())

	return r.err
}

func (r *recordingUpdater) CompleteUpdate() error {
	return nil
}

func (r *recordingUpdater) appliedVersions() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	return append([]string(nil), r.applied...)
}

func TestWatch(t *testing.T) {
	t.Parallel()

	t.Run("a versioned agent checks again after the interval", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		server := &serverVersions{version: semver.MustParse("0.21.0")}
		updater := new(recordingUpdater)
		tick := make(chan time.Time)

		done := make(chan struct{})
		go func() {
			defer close(done)
			Watch(ctx, log.NewEntry(log.StandardLogger()), "0.20.0", semver.MustParse("0.20.0"), server.check, updater, tick)
		}()

		require.Eventually(t, func() bool { return server.count() == 1 }, time.Second, time.Millisecond, "the agent did not check as it started")

		select {
		case tick <- time.Time{}:
		case <-time.After(time.Second):
			require.FailNow(t, "the agent is not waiting for the interval to pass")
		}

		require.Eventually(t, func() bool { return len(updater.appliedVersions()) == 2 }, time.Second, time.Millisecond, "the agent did not check again after the interval")
		assert.Equal(t, 2, server.count())
		assert.Equal(t, []string{"0.21.0", "0.21.0"}, updater.appliedVersions())

		cancel()
		<-done
	})

	t.Run("a failed update is not reported as applied", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		logger, hook := logtest.NewNullLogger()
		server := &serverVersions{version: semver.MustParse("0.21.0")}
		updater := &recordingUpdater{err: errors.New("the image could not be pulled")}

		done := make(chan struct{})
		go func() {
			defer close(done)
			Watch(ctx, log.NewEntry(logger), "0.20.0", semver.MustParse("0.20.0"), server.check, updater, make(chan time.Time))
		}()

		require.Eventually(t, func() bool {
			return hook.LastEntry() != nil && hook.LastEntry().Message == "Waiting for the next update check"
		}, time.Second, time.Millisecond, "the agent did not finish its first check")

		messages := make([]string, 0)
		for _, entry := range hook.AllEntries() {
			messages = append(messages, entry.Message)
		}

		assert.Equal(t, []string{"0.21.0"}, updater.appliedVersions())
		assert.Contains(t, messages, "Failed to apply update")
		assert.NotContains(t, messages, "Update successfully applied")

		cancel()
		<-done
	})

	t.Run("an agent reporting latest never checks for updates", func(t *testing.T) {
		t.Parallel()

		server := &serverVersions{version: semver.MustParse("0.21.0")}
		updater := new(recordingUpdater)

		done := make(chan struct{})
		go func() {
			defer close(done)
			Watch(t.Context(), log.NewEntry(log.StandardLogger()), "latest", new(semver.Version), server.check, updater, make(chan time.Time))
		}()

		select {
		case <-done:
		case <-time.After(time.Second):
			require.FailNow(t, "the latest agent is watching for updates")
		}

		assert.Zero(t, server.count())
		assert.Empty(t, updater.appliedVersions())
	})
}
