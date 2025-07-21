package web

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/ssh/web/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConnReadMessage_input(t *testing.T) {
	socket := new(mocks.Socket)
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
				socket.On("Read", mock.AnythingOfType("[]uint8")).Return(24, nil).Run(func(args mock.Arguments) {
					b := args.Get(0).([]byte)

					buf, _ := json.Marshal(Message{
						Kind: messageKindInput,
						Data: []byte("a"),
					})

					copy(b, buf)
				}).Once()
			},
			expect: Expected{
				message: &Message{
					Kind: messageKindInput,
					Data: []byte("a"),
				},
				read: 24,
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
	socket := new(mocks.Socket)
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
