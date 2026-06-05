package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// spyTransport is a sentry.Transport that records every event sent to it, for
// use in unit tests. It is safe to use from multiple goroutines.
type spyTransport struct {
	mu     sync.Mutex
	events []*sentry.Event
	done   chan struct{} // closed on first SendEvent call
	once   sync.Once
}

func newSpyTransport() *spyTransport {
	return &spyTransport{
		done: make(chan struct{}),
	}
}

func (s *spyTransport) Configure(_ sentry.ClientOptions) {}

func (s *spyTransport) SendEvent(e *sentry.Event) {
	s.mu.Lock()
	s.events = append(s.events, e)
	s.mu.Unlock()
	// Signal that at least one event was delivered.
	s.once.Do(func() { close(s.done) })
}

func (s *spyTransport) Flush(_ time.Duration) bool              { return true }
func (s *spyTransport) FlushWithContext(_ context.Context) bool { return true }
func (s *spyTransport) Close()                                  {}

// reported returns the number of events captured so far.
func (s *spyTransport) reported() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.events)
}

// waitForEvent blocks until at least one event has been delivered or the
// deadline is reached. Returns true if an event was received.
func (s *spyTransport) waitForEvent(timeout time.Duration) bool {
	select {
	case <-s.done:
		return true
	case <-time.After(timeout):
		return false
	}
}

// newSpyClient returns a *sentry.Client wired to a spyTransport together with
// the spy so tests can inspect captured events.
func newSpyClient(t *testing.T) (*sentry.Client, *spyTransport) {
	t.Helper()

	spy := newSpyTransport()

	client, err := sentry.NewClient(sentry.ClientOptions{ //nolint:exhaustruct
		Transport: spy,
		Dsn:       "https://public@sentry.example.com/1",
	})
	require.NoError(t, err)

	return client, spy
}

// newEchoCtx returns an echo.Context backed by an HTTP recorder for assertions.
func newEchoCtx() (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	return e.NewContext(req, rec), rec
}

func TestNewErrors(t *testing.T) {
	cases := []struct {
		description  string
		err          error
		wantStatus   int
		wantReported bool
	}{
		{
			// (1) store.ErrInternal must yield 500 AND invoke report().
			description:  "store.ErrInternal yields 500 and is reported to sentry",
			err:          errors.Wrap(store.ErrInternal, errors.New("some internal detail", store.ErrLayer, store.ErrCodeInternal)),
			wantStatus:   http.StatusInternalServerError,
			wantReported: true,
		},
		{
			// (2) A plain store-layer error that is NOT ErrInternal yields 500 but
			//     MUST NOT invoke report().
			description:  "store.ErrNoDocuments yields 500 without reporting to sentry",
			err:          store.ErrNoDocuments,
			wantStatus:   http.StatusInternalServerError,
			wantReported: false,
		},
		{
			// (3) context.Canceled must NOT be reported.
			description:  "context.Canceled yields 500 without reporting to sentry",
			err:          context.Canceled,
			wantStatus:   http.StatusInternalServerError,
			wantReported: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			client, spy := newSpyClient(t)
			handler := NewErrors(client)

			ctx, rec := newEchoCtx()
			handler(tc.err, ctx)

			assert.Equal(t, tc.wantStatus, rec.Code, "unexpected HTTP status code")

			if tc.wantReported {
				// report() fires a goroutine; wait for the event to arrive.
				got := spy.waitForEvent(500 * time.Millisecond)
				require.True(t, got, "expected exactly one sentry event, got none within 500ms")
				assert.Equal(t, 1, spy.reported(), "expected exactly one sentry event")
			} else {
				// Give any (erroneous) async report a chance to arrive, then assert silence.
				time.Sleep(50 * time.Millisecond)
				assert.Equal(t, 0, spy.reported(), "expected no sentry events")
			}
		})
	}
}
