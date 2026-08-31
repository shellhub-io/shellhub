package dialer

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/multiformats/go-multistream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// greeting stands for the opening line of a server-first protocol, as SMTP and
// MySQL send on connect.
const greeting = "220 smtp.example.com ESMTP ready\r\n"

func pipeWithDeadline(t *testing.T) (net.Conn, net.Conn) {
	t.Helper()

	client, agent := net.Pipe()

	t.Cleanup(func() {
		_ = client.Close()
		_ = agent.Close()
	})

	require.NoError(t, client.SetDeadline(time.Now().Add(5*time.Second))) //nolint:forbidigo // a deadline, an elapsed-time measurement, or the clock mock itself
	require.NoError(t, agent.SetDeadline(time.Now().Add(5*time.Second)))  //nolint:forbidigo // a deadline, an elapsed-time measurement, or the clock mock itself

	return client, agent
}

func TestHTTPProxyTargetKeepsGreetingSentWithTheReply(t *testing.T) {
	client, agent := pipeWithDeadline(t)

	go func() {
		mux := multistream.NewMultistreamMuxer[string]()
		mux.AddHandler(ProtoHTTPProxy, nil)

		if _, _, err := mux.Negotiate(agent); err != nil {
			return
		}

		headers := map[string]string{}
		if err := json.NewDecoder(agent).Decode(&headers); err != nil {
			return
		}

		agent.Write([]byte(`{"status":"ok"}` + greeting)) //nolint:errcheck
	}()

	conn, err := HTTPProxyTarget{Host: "127.0.0.1", Port: 5432}.
		prepare(context.Background(), client, TransportVersion2)
	require.NoError(t, err)

	buf := make([]byte, len(greeting))
	_, err = io.ReadFull(conn, buf)
	require.NoError(t, err)
	assert.Equal(t, greeting, string(buf))
}

func TestHTTPProxyTargetV1KeepsGreetingSentWithTheReply(t *testing.T) {
	client, agent := pipeWithDeadline(t)

	go func() {
		if _, err := http.ReadRequest(bufio.NewReader(agent)); err != nil {
			return
		}

		agent.Write([]byte("HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n" + greeting)) //nolint:errcheck
	}()

	conn, err := HTTPProxyTarget{Host: "127.0.0.1", Port: 5432}.
		prepare(context.Background(), client, TransportVersion1)
	require.NoError(t, err)

	buf := make([]byte, len(greeting))
	_, err = io.ReadFull(conn, buf)
	require.NoError(t, err)
	assert.Equal(t, greeting, string(buf))
}
