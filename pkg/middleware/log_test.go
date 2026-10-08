package middleware

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	echo "github.com/labstack/echo/v5"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSlogLoggerLevels(t *testing.T) {
	cases := []struct {
		description string
		log         func(logger *slog.Logger)
		expected    logrus.Level
	}{
		{
			description: "maps debug",
			log:         func(logger *slog.Logger) { logger.Debug("message") },
			expected:    logrus.DebugLevel,
		},
		{
			description: "maps info",
			log:         func(logger *slog.Logger) { logger.Info("message") },
			expected:    logrus.InfoLevel,
		},
		{
			description: "maps warn",
			log:         func(logger *slog.Logger) { logger.Warn("message") },
			expected:    logrus.WarnLevel,
		},
		{
			description: "maps error",
			log:         func(logger *slog.Logger) { logger.Error("message") },
			expected:    logrus.ErrorLevel,
		},
		{
			description: "maps a level above error onto error",
			log:         func(logger *slog.Logger) { logger.Log(t.Context(), slog.LevelError+4, "message") },
			expected:    logrus.ErrorLevel,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			backend, hook := test.NewNullLogger()
			backend.SetLevel(logrus.TraceLevel)

			tc.log(NewSlogLogger(logrus.NewEntry(backend)))

			entry := hook.LastEntry()
			require.NotNil(t, entry)
			assert.Equal(t, tc.expected, entry.Level)
			assert.Equal(t, "message", entry.Message)
		})
	}
}

func TestSlogLoggerAttrs(t *testing.T) {
	cases := []struct {
		description string
		log         func(logger *slog.Logger)
		expected    logrus.Fields
	}{
		{
			description: "carries record attributes as fields",
			log:         func(logger *slog.Logger) { logger.Info("message", "tenant", "acme") },
			expected:    logrus.Fields{"tenant": "acme"},
		},
		{
			description: "carries attributes bound to the logger",
			log:         func(logger *slog.Logger) { logger.With("tenant", "acme").Info("message", "device", "d1") },
			expected:    logrus.Fields{"tenant": "acme", "device": "d1"},
		},
		{
			description: "qualifies keys with the open group",
			log:         func(logger *slog.Logger) { logger.WithGroup("request").Info("message", "id", "r1") },
			expected:    logrus.Fields{"request.id": "r1"},
		},
		{
			description: "qualifies keys with nested groups",
			log: func(logger *slog.Logger) {
				logger.WithGroup("request").WithGroup("route").Info("message", "id", "r1")
			},
			expected: logrus.Fields{"request.route.id": "r1"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			backend, hook := test.NewNullLogger()
			backend.SetLevel(logrus.TraceLevel)

			tc.log(NewSlogLogger(logrus.NewEntry(backend)))

			entry := hook.LastEntry()
			require.NotNil(t, entry)
			assert.Equal(t, tc.expected, entry.Data)
		})
	}
}

func TestSlogLoggerRespectsBackendLevel(t *testing.T) {
	backend, hook := test.NewNullLogger()
	backend.SetLevel(logrus.WarnLevel)

	logger := NewSlogLogger(logrus.NewEntry(backend))
	logger.Info("dropped")

	assert.Nil(t, hook.LastEntry())

	logger.Warn("kept")

	require.NotNil(t, hook.LastEntry())
	assert.Equal(t, "kept", hook.LastEntry().Message)
}

func TestLogRedactsCredentialsInTheQuery(t *testing.T) {
	cases := []struct {
		description string
		target      string
		handlerErr  error
		expected    string
	}{
		{
			description: "hides a web terminal token and keeps the terminal size",
			target:      "/ws/ssh?token=secret&cols=80&rows=24",
			expected:    "/ws/ssh?token=REDACTED&cols=80&rows=24",
		},
		{
			description: "hides an invitation code",
			target:      "/api/invitations/resolve?invite=secret",
			expected:    "/api/invitations/resolve?invite=REDACTED",
		},
		{
			description: "hides the email and token of an account confirmation",
			target:      "/api/user/validation_account?email=user%40example.com&token=secret",
			expected:    "/api/user/validation_account?email=REDACTED&token=REDACTED",
		},
		{
			description: "hides an approval code and keeps the fingerprint",
			target:      "/api/user/saml/reauth?approval_code=secret&fingerprint=SHA256:abc",
			expected:    "/api/user/saml/reauth?approval_code=REDACTED&fingerprint=SHA256:abc",
		},
		{
			description: "hides a percent-encoded value",
			target:      "/ws/ssh?token=%73ecret",
			expected:    "/ws/ssh?token=REDACTED",
		},
		{
			description: "hides the value of a percent-encoded name and keeps the name as sent",
			target:      "/ws/ssh?%74oken=secret",
			expected:    "/ws/ssh?%74oken=REDACTED",
		},
		{
			description: "hides every occurrence of a repeated parameter",
			target:      "/ws/ssh?token=first&token=second",
			expected:    "/ws/ssh?token=REDACTED&token=REDACTED",
		},
		{
			description: "hides a name sent in another case, as the binder matches names case-insensitively",
			target:      "/api/invitations/resolve?Invite=secret&TOKEN=secret",
			expected:    "/api/invitations/resolve?Invite=REDACTED&TOKEN=REDACTED",
		},
		{
			description: "hides the whole query when it cannot be parsed",
			target:      "/ws/ssh?token=%zzsecret&cols=80",
			expected:    "/ws/ssh?REDACTED",
		},
		{
			description: "hides the whole query when it uses a semicolon separator",
			target:      "/ws/ssh?cols=80;token=secret",
			expected:    "/ws/ssh?REDACTED",
		},
		{
			description: "hides a parameter sent without a value",
			target:      "/ws/ssh?token&cols=80",
			expected:    "/ws/ssh?token=REDACTED&cols=80",
		},
		{
			description: "hides a token in an absolute-form request URI",
			target:      "http://example.com/ws/ssh?token=secret&cols=80",
			expected:    "http://example.com/ws/ssh?token=REDACTED&cols=80",
		},
		{
			description: "keeps a URI with no query",
			target:      "/api/devices",
			expected:    "/api/devices",
		},
		{
			description: "hides a token on a request that fails",
			target:      "/ws/ssh?token=secret",
			handlerErr:  echo.ErrBadRequest,
			expected:    "/ws/ssh?token=REDACTED",
		},
	}

	hook := test.NewGlobal()
	t.Cleanup(hook.Reset)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			hook.Reset()

			router := echo.New()
			router.Use(Log)
			router.GET("/*", func(*echo.Context) error { return tc.handlerErr })

			router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(t.Context(), http.MethodGet, tc.target, nil))

			entry := hook.LastEntry()
			require.NotNil(t, entry)
			assert.Equal(t, tc.expected, entry.Data["uri"])
		})
	}
}
