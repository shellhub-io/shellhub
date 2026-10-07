package dialer

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/multiformats/go-multistream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const greeting = "220 smtp.example.com ESMTP ready\r\n"

const okReply = `{"status":"ok"}`

const httpResponse = "HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n"

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

func answerHTTPProxy(agent net.Conn, reply string) {
	mux := multistream.NewMultistreamMuxer[string]()
	mux.AddHandler(ProtoHTTPProxy, nil)

	if _, _, err := mux.Negotiate(agent); err != nil {
		return
	}

	headers := map[string]string{}
	if err := json.NewDecoder(agent).Decode(&headers); err != nil {
		return
	}

	agent.Write([]byte(reply)) //nolint:errcheck // a lost reply fails the test through prepare's decode
}

func TestHTTPProxyTargetKeepsGreetingSentWithTheReply(t *testing.T) {
	client, agent := pipeWithDeadline(t)

	go answerHTTPProxy(agent, okReply+greeting)

	conn, err := HTTPProxyTarget{Host: "127.0.0.1", Port: 5432}.
		prepare(context.Background(), client, TransportVersion2)
	require.NoError(t, err)

	buf := make([]byte, len(greeting))
	_, err = io.ReadFull(conn, buf)
	require.NoError(t, err)
	assert.Equal(t, greeting, string(buf))
}

func TestHTTPProxyTargetDropsTheNewlineEndingTheReply(t *testing.T) {
	client, agent := pipeWithDeadline(t)

	go answerHTTPProxy(agent, okReply+"\n"+httpResponse)

	conn, err := HTTPProxyTarget{Host: "127.0.0.1", Port: 3000}.
		prepare(context.Background(), client, TransportVersion2)
	require.NoError(t, err)

	buf := make([]byte, len(httpResponse))
	_, err = io.ReadFull(conn, buf)
	require.NoError(t, err)
	assert.Equal(t, httpResponse, string(buf))
}

func TestHTTPProxyTargetReturnsBeforeTheDeviceSendsAnything(t *testing.T) {
	client, agent := pipeWithDeadline(t)

	go func() {
		answerHTTPProxy(agent, okReply)

		if _, err := http.ReadRequest(bufio.NewReader(agent)); err != nil {
			return
		}

		agent.Write([]byte(httpResponse)) //nolint:errcheck // a lost response fails the test through ReadFull
	}()

	conn, err := HTTPProxyTarget{Host: "127.0.0.1", Port: 3000}.
		prepare(context.Background(), client, TransportVersion2)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://127.0.0.1:3000/", nil)
	require.NoError(t, err)
	require.NoError(t, req.Write(conn))

	buf := make([]byte, len(httpResponse))
	_, err = io.ReadFull(conn, buf)
	require.NoError(t, err)
	assert.Equal(t, httpResponse, string(buf))
}

func TestHTTPProxyTargetDropsANewlineThatArrivesAfterAFailedRead(t *testing.T) {
	client, agent := pipeWithDeadline(t)

	firstReadFailed := make(chan struct{})

	go func() {
		answerHTTPProxy(agent, okReply)

		<-firstReadFailed

		agent.Write([]byte("\n" + httpResponse)) //nolint:errcheck // a lost response fails the test through ReadFull
	}()

	conn, err := HTTPProxyTarget{Host: "127.0.0.1", Port: 3000}.
		prepare(context.Background(), client, TransportVersion2)
	require.NoError(t, err)

	require.NoError(t, conn.SetReadDeadline(time.Now().Add(50*time.Millisecond))) //nolint:forbidigo // a deadline, so the first read times out before the newline arrives
	_, err = conn.Read(make([]byte, 1))
	require.ErrorIs(t, err, os.ErrDeadlineExceeded)

	require.NoError(t, conn.SetReadDeadline(time.Now().Add(5*time.Second))) //nolint:forbidigo // a deadline, so a regression fails the test instead of hanging it
	close(firstReadFailed)

	buf := make([]byte, len(httpResponse))
	_, err = io.ReadFull(conn, buf)
	require.NoError(t, err)
	assert.Equal(t, httpResponse, string(buf))
}

func TestHTTPProxyTargetV1KeepsGreetingSentWithTheReply(t *testing.T) {
	client, agent := pipeWithDeadline(t)

	go func() {
		if _, err := http.ReadRequest(bufio.NewReader(agent)); err != nil {
			return
		}

		agent.Write([]byte(httpResponse + greeting)) //nolint:errcheck // a lost reply fails the test through ReadFull
	}()

	conn, err := HTTPProxyTarget{Host: "127.0.0.1", Port: 5432}.
		prepare(context.Background(), client, TransportVersion1)
	require.NoError(t, err)

	buf := make([]byte, len(greeting))
	_, err = io.ReadFull(conn, buf)
	require.NoError(t, err)
	assert.Equal(t, greeting, string(buf))
}

func TestHTTPProxyTargetReportsTheAgentsReason(t *testing.T) {
	client, agent := pipeWithDeadline(t)

	go answerHTTPProxy(agent, `{"error":"address not found on the device"}`)

	_, err := HTTPProxyTarget{Host: "127.0.0.1", Port: 5432}.
		prepare(context.Background(), client, TransportVersion2)

	var refused *ProxyRefusedError
	require.ErrorAs(t, err, &refused)
	assert.Equal(t, "address not found on the device", refused.Reason)
}
