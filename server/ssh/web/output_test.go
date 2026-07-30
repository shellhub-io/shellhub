package web

import (
	"bytes"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/websocket"
)

// scriptedReader hands out the stream in the chunk sizes it was scripted with,
// so a read boundary can be placed on any byte.
type scriptedReader struct {
	steps [][]byte
	i     int
}

func (r *scriptedReader) Read(p []byte) (int, error) {
	if r.i >= len(r.steps) {
		return 0, io.EOF
	}

	step := r.steps[r.i]
	if len(step) > len(p) {
		return 0, io.ErrShortBuffer
	}

	r.i++

	return copy(p, step), nil
}

// framingSocket records every write as a distinct frame.
type framingSocket struct {
	frames [][]byte
}

func (s *framingSocket) Write(p []byte) (int, error) {
	s.frames = append(s.frames, bytes.Clone(p))

	return len(p), nil
}

func (s *framingSocket) Read(_ []byte) (int, error) { return 0, io.EOF }

func (s *framingSocket) Close() error { return nil }

func (s *framingSocket) joined() []byte {
	return bytes.Join(s.frames, nil)
}

func TestOutputWriterPreservesBytes(t *testing.T) {
	cases := []struct {
		description string
		steps       [][]byte
	}{
		{
			description: "utf-8 split mid-character across reads",
			steps:       [][]byte{{0xe4, 0xb8}, {0xad, 0xe6, 0x96, 0x87}},
		},
		{
			description: "gbk bytes",
			steps:       [][]byte{{0xd6, 0xd0, 0xce, 0xc4, 0xc4, 0xbf, 0xc2, 0xbc}},
		},
		{
			description: "latin-1 high bytes",
			steps:       [][]byte{{0x41, 0xe9, 0xe8, 0x42}},
		},
		{
			description: "a chunk of only continuation bytes",
			steps:       [][]byte{{0x80, 0x81, 0x82}},
		},
		{
			description: "a zero-byte read followed by more data",
			steps:       [][]byte{{}, {0x41, 0x42, 0x43}},
		},
		{
			description: "every byte on its own read",
			steps:       [][]byte{{0xe4}, {0xb8}, {0xad}},
		},
		{
			description: "arbitrary binary output",
			steps:       [][]byte{{0x00, 0x01, 0xff, 0xfe, 0x7f, 0x80}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			socket := &framingSocket{}
			writer := newOutputWriter(NewConn(socket))

			_, err := io.Copy(writer, &scriptedReader{steps: tc.steps})
			require.NoError(t, err)

			assert.Equal(t, bytes.Join(tc.steps, nil), socket.joined())
		})
	}
}

// TestOutputWriterSendsBinaryFrames pins the frame type over a real connection.
// The package's socket mock cannot show it, because WriteBinary falls back to a
// plain write when the socket is not a websocket.
func TestOutputWriterSendsBinaryFrames(t *testing.T) {
	// Invalid UTF-8: a text frame carrying these bytes violates RFC 6455 and the
	// browser is entitled to fail the connection over it.
	payload := []byte{0xd6, 0xd0, 0xce, 0xc4}

	server := httptest.NewServer(websocket.Handler(func(socket *websocket.Conn) {
		writer := newOutputWriter(NewConn(socket))

		if _, err := writer.Write(payload); err != nil {
			t.Errorf("failed to write the output: %v", err)
		}
	}))

	defer server.Close()

	client, err := websocket.Dial("ws"+strings.TrimPrefix(server.URL, "http"), "", server.URL)
	require.NoError(t, err)

	defer client.Close() //nolint:errcheck

	var frame capturedFrame

	require.NoError(t, frameCapture.Receive(client, &frame))

	assert.Equal(t, byte(websocket.BinaryFrame), frame.payloadType)
	assert.Equal(t, payload, frame.data)
}

type capturedFrame struct {
	payloadType byte
	data        []byte
}

// frameCapture exposes the frame's opcode, which the stock Message codec
// discards when unmarshalling.
var frameCapture = websocket.Codec{
	Marshal: func(_ any) ([]byte, byte, error) {
		return nil, websocket.UnknownFrame, websocket.ErrNotSupported
	},
	Unmarshal: func(msg []byte, payloadType byte, v any) error {
		frame, ok := v.(*capturedFrame)
		if !ok {
			return websocket.ErrNotSupported
		}

		frame.payloadType = payloadType
		frame.data = msg

		return nil
	},
}
