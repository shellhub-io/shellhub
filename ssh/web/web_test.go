package web

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	cachemock "github.com/shellhub-io/shellhub/pkg/cache/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/websocket"
)

func TestNewSSHServerBridge_CredentialsNotFound(t *testing.T) {
	e := echo.New()
	cache := new(cachemock.Cache)

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
		defer conn.Close()

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
