package web

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	servicemocks "github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/webhandoff"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/websocket"
)

// writeHostKey writes a throwaway host key for the bridge to pin to, and returns its path.
func writeHostKey(t *testing.T) string {
	t.Helper()

	_, key, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	block, err := ssh.MarshalPrivateKey(key, "")
	require.NoError(t, err)

	path := filepath.Join(t.TempDir(), "ssh.key")
	require.NoError(t, os.WriteFile(path, pem.EncodeToMemory(block), 0o600))

	return path
}

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

	require.NoError(t, NewSSHServerBridge(e, nil, nil, webhandoff.NewStore(), &Config{HostKeyFile: writeHostKey(t)}))

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

const tenantID = "00000000-0000-4000-0000-000000000000"

// deviceInNamespace is what the store answers for a device lookup bounded to the caller's
// namespace: the device when it belongs there, and nothing when it does not.
func deviceInNamespace(inScope bool) (*models.Device, error) {
	if !inScope {
		return nil, store.ErrNoDocuments
	}

	return &models.Device{UID: "device-uid", TenantID: tenantID}, nil //nolint:exhaustruct // the handler only checks that the device resolves
}

func TestWebSessionRouteRequiresTheConnectPermission(t *testing.T) {
	tests := []struct {
		description   string
		role          string
		deviceInScope bool
		expectedCode  int
	}{
		{
			description:  "observer is refused",
			role:         authorizer.RoleObserver.String(),
			expectedCode: http.StatusForbidden,
		},
		{
			description:  "an identity carrying no role is refused",
			role:         "",
			expectedCode: http.StatusForbidden,
		},
		{
			description:   "operator is served a token for a device in their namespace",
			role:          authorizer.RoleOperator.String(),
			deviceInScope: true,
			expectedCode:  http.StatusOK,
		},
		{
			description:  "a device outside the caller's namespace is refused",
			role:         authorizer.RoleOperator.String(),
			expectedCode: http.StatusForbidden,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			service := servicemocks.NewMockService(t)
			if test.role == authorizer.RoleOperator.String() {
				service.On("GetDevice", mock.Anything, mock.Anything, models.UID("device-uid")).
					Return(deviceInNamespace(test.deviceInScope)).Once()
			}

			e := echo.New()
			e.Use(gateway.WithContext(nil))

			require.NoError(t, NewSSHServerBridge(e, nil, service, webhandoff.NewStore(), &Config{HostKeyFile: writeHostKey(t)}))

			server := httptest.NewServer(e)
			defer server.Close()

			body := strings.NewReader(`{"device":"device-uid","username":"root","password":"secret"}`)

			req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, server.URL+WebSessionRoute, body)
			require.NoError(t, err)

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Tenant-ID", tenantID)

			if test.role != "" {
				req.Header.Set("X-Role", test.role)
			}

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			defer res.Body.Close() //nolint:errcheck

			assert.Equal(t, test.expectedCode, res.StatusCode)

			if test.expectedCode != http.StatusOK {
				return
			}

			var success struct {
				Token string `json:"token"`
			}

			require.NoError(t, json.NewDecoder(res.Body).Decode(&success))
			assert.NotEmpty(t, success.Token)
		})
	}
}

func TestNewSSHServerBridge_MissingHostKey(t *testing.T) {
	err := NewSSHServerBridge(echo.New(), nil, nil, webhandoff.NewStore(), &Config{HostKeyFile: filepath.Join(t.TempDir(), "absent.key")})

	assert.ErrorIs(t, err, ErrBridgeReadHostKey)
}
