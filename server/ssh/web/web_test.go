package web

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	cachemock "github.com/shellhub-io/shellhub/pkg/cache/mocks"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/websocket"
)

// TestExitLogLevel verifies that exitLogLevel returns logrus.WarnLevel for
// expected/banner-derived errors and logrus.ErrorLevel for genuine server faults.
func TestExitLogLevel(t *testing.T) {
	tests := []struct {
		description string
		err         error
		expected    logrus.Level
	}{
		{
			description: "ErrConnect uses Warn",
			err:         ErrConnect,
			expected:    logrus.WarnLevel,
		},
		{
			description: "ErrAccessDenied uses Warn",
			err:         ErrAccessDenied,
			expected:    logrus.WarnLevel,
		},
		{
			description: "ErrInvalidSSHID uses Warn",
			err:         ErrInvalidSSHID,
			expected:    logrus.WarnLevel,
		},
		{
			description: "ErrAuthentication uses Warn",
			err:         ErrAuthentication,
			expected:    logrus.WarnLevel,
		},
		{
			description: "ErrGetAuth uses Warn",
			err:         ErrGetAuth,
			expected:    logrus.WarnLevel,
		},
		{
			description: "ErrFindDevice uses Warn",
			err:         ErrFindDevice,
			expected:    logrus.WarnLevel,
		},
		{
			description: "ErrForbiddenPublicKey uses Warn",
			err:         ErrForbiddenPublicKey,
			expected:    logrus.WarnLevel,
		},
		{
			description: "ErrBridgeCredentialsNotFound uses Warn",
			err:         ErrBridgeCredentialsNotFound,
			expected:    logrus.WarnLevel,
		},
		{
			description: "ErrWebSocketGetToken uses Warn",
			err:         ErrWebSocketGetToken,
			expected:    logrus.WarnLevel,
		},
		{
			description: "ErrWebSocketGetDimensions uses Warn",
			err:         ErrWebSocketGetDimensions,
			expected:    logrus.WarnLevel,
		},
		{
			description: "ErrWebSocketGetIP uses Warn",
			err:         ErrWebSocketGetIP,
			expected:    logrus.WarnLevel,
		},
		{
			description: "ErrSession uses Error",
			err:         ErrSession,
			expected:    logrus.ErrorLevel,
		},
		{
			description: "ErrPty uses Error",
			err:         ErrPty,
			expected:    logrus.ErrorLevel,
		},
		{
			description: "ErrShell uses Error",
			err:         ErrShell,
			expected:    logrus.ErrorLevel,
		},
		{
			description: "unrecognised error uses Error",
			err:         ErrPublicKey,
			expected:    logrus.ErrorLevel,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			assert.Equal(t, test.expected, exitLogLevel(test.err))
		})
	}
}

func TestNewSSHServerBridge_CredentialsNotFound(t *testing.T) {
	e := echo.New()
	cache := new(cachemock.MockCache)

	NewSSHServerBridge(e, cache)

	server := httptest.NewServer(e)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/ssh?token=nonexistent&cols=80&rows=24"
	origin := server.URL

	assert.NotPanics(t, func() {
		cfg, err := websocket.NewConfig(wsURL, origin)
		require.NoError(t, err)

		cfg.Header.Set("X-Real-Ip", "127.0.0.1")

		conn, err := websocket.DialConfig(cfg)
		require.NoError(t, err)
		defer conn.Close() //nolint:errcheck

		var raw []byte
		err = websocket.Message.Receive(conn, &raw)
		require.NoError(t, err)

		var msg Message
		require.NoError(t, json.Unmarshal(raw, &msg))
		assert.Equal(t, messageKindError, msg.Kind)

		data, ok := msg.Data.(string)
		require.True(t, ok)
		assert.Contains(t, data, ErrBridgeCredentialsNotFound.Error())
	}, "handler must not panic when credentials are not found")
}
