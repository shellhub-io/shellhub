package tunnel

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"testing"
	"time"

	"github.com/multiformats/go-multistream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeStream struct {
	io.Reader
	io.Writer
}

func (fakeStream) Close() error { return nil }

func newFakeStream(wire []byte) io.ReadWriteCloser {
	return fakeStream{Reader: bytes.NewReader(wire), Writer: io.Discard}
}

// TestStreamKeepsThePayloadTheHeaderDecoderBuffered is the framing contract both ends of a V2
// stream depend on. The server writes the header and the payload's first bytes back to back,
// so they arrive in one read, and json.Decoder reads in blocks: it takes the payload into its
// own buffer along with the header. Reading the connection after that yields nothing.
func TestStreamKeepsThePayloadTheHeaderDecoderBuffered(t *testing.T) {
	const payload = "SSH-2.0-OpenSSH_9.6\r\n"

	rwc := newFakeStream([]byte(`{"id":"session-uid"}` + "\n" + payload))

	ctx := NewContext(context.Background(), rwc)

	headers, err := ctx.Headers()
	require.NoError(t, err)
	assert.Equal(t, "session-uid", headers["id"])

	rest, err := io.ReadAll(ctx.Stream())
	require.NoError(t, err)

	assert.Equal(t, payload, string(rest),
		"the payload that followed the header in the same read must still reach the handler")
}

// TestHandleServesTheStreamThatKeptThePayload drives a real negotiated stream, because the
// contract that broke was the dispatch one: a handler given the raw connection instead of the
// stream reads a payload with its first bytes missing, and the SSH handshake on it dies with
// EOF against a server that never sees a version banner.
func TestHandleServesTheStreamThatKeptThePayload(t *testing.T) {
	const protocol = "/ssh/open/1.0.0"
	const payload = "SSH-2.0-OpenSSH_9.6\r\n"

	tun := &TunnelV2{mux: multistream.NewMultistreamMuxer[string]()} //nolint:exhaustruct // the muxer is all Handle and the dispatch under test touch

	served := make(chan string, 1)
	failed := make(chan error, 1)

	tun.Handle(protocol, func(ctx Context, rwc io.ReadWriteCloser) error {
		headers, err := ctx.Headers()
		if err != nil {
			failed <- err

			return err
		}

		got := make([]byte, len(payload))
		if _, err := io.ReadFull(rwc, got); err != nil {
			failed <- err

			return err
		}

		served <- headers["id"] + "|" + string(got)

		return nil
	})

	server, client := net.Pipe()
	defer client.Close() //nolint:errcheck // the test fails on the assertions, not on the teardown

	go tun.mux.Handle(server) //nolint:errcheck // the assertions below read the outcome

	require.NoError(t, client.SetDeadline(time.Now().Add(5*time.Second))) //nolint:forbidigo // a deadline on the pipe, so a regression fails the test instead of hanging it
	require.NoError(t, multistream.SelectProtoOrFail(protocol, client))

	frame, err := json.Marshal(map[string]string{"id": "session-uid"})
	require.NoError(t, err)

	_, err = client.Write(append(append(frame, '\n'), []byte(payload)...))
	require.NoError(t, err)

	select {
	case got := <-served:
		assert.Equal(t, "session-uid|"+payload, got,
			"the handler must be served the stream that kept the payload, not the raw connection")
	case err := <-failed:
		t.Fatalf("the handler could not read the payload that followed the header: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("the handler never saw the payload that followed the header in the same write")
	}
}
