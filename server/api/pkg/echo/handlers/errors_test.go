package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/errors"
	routes "github.com/shellhub-io/shellhub/server/api/routes/errors"
	"github.com/shellhub-io/shellhub/server/api/services"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	s.once.Do(func() { close(s.done) })
}

func (s *spyTransport) Flush(_ time.Duration) bool              { return true }
func (s *spyTransport) FlushWithContext(_ context.Context) bool { return true }
func (s *spyTransport) Close()                                  {}

func (s *spyTransport) reported() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.events)
}

func (s *spyTransport) firstMessage() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.events) == 0 {
		return ""
	}

	return s.events[0].Message
}

func (s *spyTransport) waitForEvent(timeout time.Duration) bool {
	select {
	case <-s.done:
		return true
	case <-time.After(timeout):
		return false
	}
}

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

func newEchoCtx() (*echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	return e.NewContext(req, rec), rec
}

func TestNewErrors(t *testing.T) {
	cases := []struct {
		description  string
		err          error
		wantStatus   int
		wantBody     responses.Error
		wantNoBody   bool
		wantReported bool
	}{
		{
			description:  "store.ErrInternal yields a generic 500 body and is reported to sentry",
			err:          errors.Wrap(store.ErrInternal, errors.New("some internal detail", store.ErrLayer, store.ErrCodeInternal)),
			wantStatus:   http.StatusInternalServerError,
			wantBody:     responses.Error{Message: "internal server error"},
			wantReported: true,
		},
		{
			description:  "store.ErrNoDocuments yields a generic 500 body without reporting to sentry",
			err:          store.ErrNoDocuments,
			wantStatus:   http.StatusInternalServerError,
			wantBody:     responses.Error{Message: "internal server error"},
			wantReported: false,
		},
		{
			description:  "context.Canceled yields a generic 500 body without reporting to sentry",
			err:          context.Canceled,
			wantStatus:   http.StatusInternalServerError,
			wantBody:     responses.Error{Message: "internal server error"},
			wantReported: false,
		},
		{
			description:  "an echo error yields its status and message",
			err:          echo.ErrNotFound,
			wantStatus:   http.StatusNotFound,
			wantBody:     responses.Error{Message: "Not Found"},
			wantReported: false,
		},
		{
			description:  "an echo HTTPError yields its status and message",
			err:          echo.NewHTTPError(http.StatusRequestTimeout, "request timeout"),
			wantStatus:   http.StatusRequestTimeout,
			wantBody:     responses.Error{Message: "request timeout"},
			wantReported: false,
		},
		{
			description:  "a binding error yields 400 without echoing the underlying cause",
			err:          echo.NewBindingError("page", []string{"abc"}, "failed to bind field value to int", strconv.ErrSyntax),
			wantStatus:   http.StatusBadRequest,
			wantBody:     responses.Error{Message: "failed to bind field value to int"},
			wantReported: false,
		},
		{
			description:  "a service not-found error yields 404 and its message",
			err:          services.ErrUserNotFound,
			wantStatus:   http.StatusNotFound,
			wantBody:     responses.Error{Message: "user not found"},
			wantReported: false,
		},
		{
			description:  "a service invalid error yields 400 and its message",
			err:          services.ErrUserInvalid,
			wantStatus:   http.StatusBadRequest,
			wantBody:     responses.Error{Message: "user invalid"},
			wantReported: false,
		},
		{
			description:  "a service duplicated error yields 409 and its message",
			err:          services.ErrUserDuplicated,
			wantStatus:   http.StatusConflict,
			wantBody:     responses.Error{Message: "user duplicated"},
			wantReported: false,
		},
		{
			description:  "a service forbidden error yields 403 and its message",
			err:          services.ErrUserNotConfirmed,
			wantStatus:   http.StatusForbidden,
			wantBody:     responses.Error{Message: "user not confirmed"},
			wantReported: false,
		},
		{
			description:  "a service locked error yields 423 and its message",
			err:          services.ErrUserAwaitingApproval,
			wantStatus:   http.StatusLocked,
			wantBody:     responses.Error{Message: "user awaiting approval"},
			wantReported: false,
		},
		{
			description:  "a service payment error yields 402 and its message",
			err:          services.ErrPaymentRequired,
			wantStatus:   http.StatusPaymentRequired,
			wantBody:     responses.Error{Message: "payment required"},
			wantReported: false,
		},
		{
			description:  "a service limit error yields 403 and its message",
			err:          services.ErrMaxTagReached,
			wantStatus:   http.StatusForbidden,
			wantBody:     responses.Error{Message: "tag limit reached"},
			wantReported: false,
		},
		{
			description:  "a service store error yields a generic 500 body",
			err:          services.ErrUserUpdate,
			wantStatus:   http.StatusInternalServerError,
			wantBody:     responses.Error{Message: "internal server error"},
			wantReported: false,
		},
		{
			description:  "a service error carrying field detail yields 400 with the fields",
			err:          services.NewErrInstallKeyInvalidField(map[string]string{"usage_limit": "must be greater than zero"}),
			wantStatus:   http.StatusBadRequest,
			wantBody:     responses.Error{Message: "install key field is invalid", Fields: map[string]string{"usage_limit": "must be greater than zero"}},
			wantReported: false,
		},
		{
			description:  "a route invalid-entity error yields 400 with the fields",
			err:          routes.NewErrInvalidEntity(map[string]string{"username": "required"}),
			wantStatus:   http.StatusBadRequest,
			wantBody:     responses.Error{Message: "invalid entity", Fields: map[string]string{"username": "required"}},
			wantReported: false,
		},
		{
			description:  "a route unauthorized error yields 401 and its message",
			err:          routes.NewErrUnauthorized(nil),
			wantStatus:   http.StatusUnauthorized,
			wantBody:     responses.Error{Message: "unauthorized"},
			wantReported: false,
		},
		{
			description:  "a route unprocessable-entity error yields 422 and its message",
			err:          routes.NewErrUnprocessableEntity(nil),
			wantStatus:   http.StatusUnprocessableEntity,
			wantBody:     responses.Error{Message: "unprocessable entity"},
			wantReported: false,
		},
		{
			description:  "a service no-content-change error yields a bare 204",
			err:          services.ErrNoContentChange,
			wantStatus:   http.StatusNoContent,
			wantNoBody:   true,
			wantReported: false,
		},
		{
			description:  "an error from an unrecognised layer yields a generic 500 body",
			err:          scope.ErrEmptyTenantID,
			wantStatus:   http.StatusInternalServerError,
			wantBody:     responses.Error{Message: "internal server error"},
			wantReported: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			client, spy := newSpyClient(t)
			handler := NewErrors(client)

			ctx, rec := newEchoCtx()
			handler(ctx, tc.err)

			assert.Equal(t, tc.wantStatus, rec.Code, "unexpected HTTP status code")

			if tc.wantNoBody {
				assert.Empty(t, rec.Body.Bytes(), "expected no response body")
			} else {
				body := responses.Error{} //nolint:exhaustruct
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body), "response body is not a JSON error")
				assert.Equal(t, tc.wantBody, body, "unexpected response body")
			}

			if tc.wantReported {
				got := spy.waitForEvent(500 * time.Millisecond)
				require.True(t, got, "expected exactly one sentry event, got none within 500ms")
				assert.Equal(t, 1, spy.reported(), "expected exactly one sentry event")
				assert.Contains(t, spy.firstMessage(), "some internal detail", "the reporter must receive the real message")
			} else {
				time.Sleep(50 * time.Millisecond)
				assert.Equal(t, 0, spy.reported(), "expected no sentry events")
			}
		})
	}
}
