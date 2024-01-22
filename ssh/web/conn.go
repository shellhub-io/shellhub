package web

import (
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/shellhub-io/shellhub/pkg/clock"
	"golang.org/x/net/websocket"
)

//go:generate mockery --name Socket --filename socket.go
type Socket interface {
	io.ReadWriteCloser
}

type Conn struct {
	// Socket is the internal websocket connection the messages come from.
	Socket Socket
	// Pinger is reponsable to inform the server that a SSH session is open.
	Pinger *time.Ticker
}

func NewConn(socket Socket) *Conn {
	return &Conn{
		Socket: socket,
		Pinger: time.NewTicker(30 * time.Second),
	}
}

var (
	ErrConnReadMessageSocketRead  = errors.New("failed to read the message from socket")
	ErrConnReadMessageSocketWrite = errors.New("failed to write the message's data to socket")
	ErrConnReadMessageJSONInvalid = errors.New("failed to parse the message from json")
	ErrConnReadMessageKindInvalid = errors.New("this kind of message is invalid")
)

func (c *Conn) ReadMessage(message *Message) (int, error) {
	buffer := make([]byte, 1024)

	read, err := c.Socket.Read(buffer)
	if err != nil {
		return read, errors.Join(ErrConnReadMessageSocketRead, err)
	}

	var data json.RawMessage
	message.Data = &data

	if err = json.Unmarshal(buffer[:read], &message); err != nil {
		return 0, errors.Join(ErrConnReadMessageJSONInvalid)
	}

	switch message.Kind {
	case MessageKindInput:
		var bytes []byte

		if err = json.Unmarshal(data, &bytes); err != nil {
			return 0, errors.Join(ErrConnReadMessageJSONInvalid)
		}

		message.Data = bytes
	case MessageKindResize:
		var dim Dimensions

		if err = json.Unmarshal(data, &dim); err != nil {
			return 0, errors.Join(ErrConnReadMessageJSONInvalid)
		}

		message.Data = dim
	default:
		return 0, errors.Join(ErrConnReadMessageKindInvalid)
	}

	return read, nil
}

func (c *Conn) WriteMessage(message *Message) (int, error) {
	buffer, err := json.Marshal(message)
	if err != nil {
		return 0, errors.Join(ErrConnReadMessageJSONInvalid)
	}

	wrote, err := c.Socket.Write(buffer)
	if err != nil {
		return wrote, errors.Join(ErrConnReadMessageSocketWrite, err)
	}

	return wrote, nil
}

func (c *Conn) Read(buffer []byte) (int, error) {
	return c.Socket.Read(buffer)
}

func (c *Conn) Write(buffer []byte) (int, error) {
	return c.Socket.Write(buffer)
}

func (c *Conn) Close() error {
	c.Pinger.Stop()

	return c.Socket.Close()
}

func (c *Conn) KeepAlive() {
	socket, ok := c.Socket.(*websocket.Conn)
	if !ok {
		return
	}

	for {
		if err := socket.SetDeadline(clock.Now().Add((time.Second * 30) * 2)); err != nil {
			return
		}

		if fw, err := socket.NewFrameWriter(websocket.PingFrame); err != nil {
			return
		} else if _, err = fw.Write([]byte{}); err != nil {
			return
		}

		if _, running := <-c.Pinger.C; !running {
			return
		}
	}
}
