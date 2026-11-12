package app

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServeError(t *testing.T) {
	t.Run("treats a closed HTTP server as a graceful stop", func(t *testing.T) {
		assert.NoError(t, serveError(http.ErrServerClosed))
		assert.NoError(t, serveError(fmt.Errorf("wrapped: %w", http.ErrServerClosed)))
	})

	t.Run("treats a closed SSH server as a graceful stop", func(t *testing.T) {
		assert.NoError(t, serveError(gliderssh.ErrServerClosed))
	})

	t.Run("keeps other serving errors", func(t *testing.T) {
		sentinel := errors.New("address already in use")

		require.ErrorIs(t, serveError(sentinel), sentinel)
	})

	t.Run("waits for shutdown to finish before returning on a signal", func(t *testing.T) {
		s := &fakeLifecycle{stopped: make(chan struct{})}
		sigs := make(chan os.Signal, 1)

		go func() {
			sigs <- syscall.SIGTERM
		}()

		require.NoError(t, runServer(s, sigs))
		assert.Equal(t, int32(1), s.shutdownCalls.Load())
		assert.True(t, s.shutdownFinished.Load(), "runServer returned before Shutdown finished")
	})

	t.Run("runs shutdown when Start returns without a signal", func(t *testing.T) {
		s := &fakeLifecycle{}

		require.NoError(t, runServer(s, make(chan os.Signal, 1)))
		assert.Equal(t, int32(1), s.shutdownCalls.Load())
		assert.True(t, s.shutdownFinished.Load())
	})

	t.Run("returns a start error without shutting down", func(t *testing.T) {
		sentinel := errors.New("listen failed")
		s := &fakeLifecycle{startErr: sentinel}

		require.ErrorIs(t, runServer(s, make(chan os.Signal, 1)), sentinel)
		assert.Equal(t, int32(0), s.shutdownCalls.Load())
	})
}

type fakeLifecycle struct {
	startErr         error
	stopped          chan struct{}
	shutdownCalls    atomic.Int32
	shutdownFinished atomic.Bool
}

func (f *fakeLifecycle) Start() error {
	if f.stopped != nil {
		<-f.stopped
	}

	return f.startErr
}

func (f *fakeLifecycle) Shutdown() {
	f.shutdownCalls.Add(1)

	if f.stopped != nil {
		close(f.stopped)
	}

	time.Sleep(50 * time.Millisecond)
	f.shutdownFinished.Store(true)
}
