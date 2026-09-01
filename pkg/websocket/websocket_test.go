package websocket

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGorillaWebSocketUpgraderChecksSameOrigin(t *testing.T) {
	t.Parallel()

	upgrader := NewGorillaWebSocketUpgrader()

	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		conn, err := upgrader.Upgrade(res, req)
		if err != nil {
			return
		}

		conn.Close() //nolint:errcheck
	}))
	t.Cleanup(server.Close)

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	host := strings.TrimPrefix(server.URL, "http://")

	cases := []struct {
		name    string
		origin  string
		upgrade bool
	}{
		{
			name:    "no origin header is accepted",
			origin:  "",
			upgrade: true,
		},
		{
			name:    "matching origin is accepted",
			origin:  "http://" + host,
			upgrade: true,
		},
		{
			name:    "foreign origin is rejected",
			origin:  "http://evil.example.com",
			upgrade: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			header := http.Header{}
			if tc.origin != "" {
				header.Set("Origin", tc.origin)
			}

			conn, res, err := websocket.DefaultDialer.Dial(url, header)
			if res != nil {
				defer res.Body.Close() //nolint:errcheck
			}

			if !tc.upgrade {
				require.ErrorIs(t, err, websocket.ErrBadHandshake)
				assert.Equal(t, http.StatusForbidden, res.StatusCode)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, http.StatusSwitchingProtocols, res.StatusCode)
			conn.Close() //nolint:errcheck
		})
	}
}
