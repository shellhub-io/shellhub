package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/shellhub-io/shellhub/server/ssh/web/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/websocket"
)

func TestConnReadMessage_input(t *testing.T) {
	socket := new(mocks.MockSocket)
	conn := NewConn(socket)

	type Expected struct {
		message *Message
		read    int
		err     error
	}

	tests := []struct {
		description   string
		requiredMocks func()
		expect        Expected
	}{
		{
			description: "fail when socket reading fail",
			requiredMocks: func() {
				socket.On("Read", mock.AnythingOfType("[]uint8")).Return(0, errors.New("")).Once()
			},
			expect: Expected{
				message: &Message{
					Data: new(json.RawMessage),
				},
				read: 0,
				err:  ErrConnReadMessageJSONInvalid,
			},
		},
		{
			description: "fail when data read is not a JSON object",
			requiredMocks: func() {
				socket.On("Read", mock.AnythingOfType("[]uint8")).Return(512, nil).Once()
			},
			expect: Expected{
				message: &Message{Data: new(json.RawMessage)},
				read:    0,
				err:     ErrConnReadMessageJSONInvalid,
			},
		},
		{
			description: "success to read the message",
			requiredMocks: func() {
				socket.On("Read", mock.AnythingOfType("[]uint8")).Return(21, nil).Run(func(args mock.Arguments) {
					b := args.Get(0).([]byte)

					buf, _ := json.Marshal(Message{
						Kind: messageKindInput,
						Data: "a",
					})

					copy(b, buf)
				}).Once()
			},
			expect: Expected{
				message: &Message{
					Kind: messageKindInput,
					Data: "a",
				},
				read: 21,
				err:  nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.requiredMocks()

			var message Message
			read, err := conn.ReadMessage(&message)

			assert.Equal(t, test.expect.message, &message)
			assert.Equal(t, test.expect.read, read)
			assert.ErrorIs(t, err, test.expect.err)
		})
	}
}

func TestConnReadMessage_resize(t *testing.T) {
	socket := new(mocks.MockSocket)
	conn := NewConn(socket)

	type Expected struct {
		message *Message
		read    int
		err     error
	}

	tests := []struct {
		description   string
		requiredMocks func()
		expect        Expected
	}{
		{
			description: "fail when socket reading fail",
			requiredMocks: func() {
				socket.On("Read", mock.AnythingOfType("[]uint8")).Return(0, errors.New("")).Once()
			},
			expect: Expected{
				message: &Message{Data: new(json.RawMessage)},
				read:    0,
				err:     ErrConnReadMessageJSONInvalid,
			},
		},
		{
			description: "fail when data read is not a JSON object",
			requiredMocks: func() {
				socket.On("Read", mock.AnythingOfType("[]uint8")).Return(512, nil).Once()
			},
			expect: Expected{
				message: &Message{Data: new(json.RawMessage)},
				read:    0,
				err:     ErrConnReadMessageJSONInvalid,
			},
		},
		{
			description: "success to read the message",
			requiredMocks: func() {
				socket.On("Read", mock.AnythingOfType("[]uint8")).Return(40, nil).Run(func(args mock.Arguments) {
					b := args.Get(0).([]byte)

					buf, _ := json.Marshal(Message{
						Kind: messageKindResize,
						Data: Dimensions{Cols: 100, Rows: 50},
					})

					copy(b, buf)
				}).Once()
			},
			expect: Expected{
				message: func() *Message {
					return &Message{
						Kind: messageKindResize,
						Data: Dimensions{Cols: 100, Rows: 50},
					}
				}(),
				read: 40,
				err:  nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.requiredMocks()

			var message Message
			read, err := conn.ReadMessage(&message)

			assert.Equal(t, test.expect.message, &message)
			assert.Equal(t, test.expect.read, read)
			assert.ErrorIs(t, err, test.expect.err)
		})
	}
}

// TestConnConcurrentWritesDoNotRace pins the synchronisation on Conn rather
// than on any one writer: output frames, control messages and keep-alive pings
// all share the socket, and the frame writer underneath is not goroutine-safe.
func TestConnConcurrentWritesDoNotRace(t *testing.T) {
	const rounds = 50

	output := bytes.Repeat([]byte{'o'}, 128)

	server := httptest.NewServer(websocket.Handler(func(socket *websocket.Conn) {
		conn := NewConn(socket)

		wg := sync.WaitGroup{}
		wg.Add(3)

		go func() {
			defer wg.Done()

			for range rounds {
				if _, err := conn.WriteBinary(output); err != nil {
					return
				}
			}
		}()

		go func() {
			defer wg.Done()

			for range rounds {
				if _, err := conn.WriteMessage(&Message{Kind: messageKindSession, Data: "uid"}); err != nil {
					return
				}
			}
		}()

		go func() {
			defer wg.Done()

			for range rounds {
				if err := conn.WritePing(); err != nil {
					return
				}
			}
		}()

		wg.Wait()
	}))

	defer server.Close()

	client, err := websocket.Dial("ws"+strings.TrimPrefix(server.URL, "http"), "", server.URL)
	require.NoError(t, err)

	defer client.Close() //nolint:errcheck

	binary := 0

	for range rounds * 2 {
		var frame capturedFrame

		if err := frameCapture.Receive(client, &frame); err != nil {
			break
		}

		if frame.payloadType == websocket.BinaryFrame {
			binary++

			// A frame carrying anything other than the whole payload means two
			// writers interleaved on the shared frame writer.
			assert.Equal(t, output, frame.data)
		}
	}

	assert.Equal(t, rounds, binary)
}
