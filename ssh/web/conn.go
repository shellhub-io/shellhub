package web

import (
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/shellhub-io/shellhub/pkg/clock"
	log "github.com/sirupsen/logrus"
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

// CharacterSize is the size of a single character in bytes when encoded in UTF-8.
const CharacterSize = 4

// TermniosMaxLineLength is the maximum line length for a terminal input in characters.
//
// [termios] is a POSIX-defined API for configuring terminal I/O settings in Unix-like systems (Linux, macOS, *BSD, etc.).
// It provides fine-grained control over how terminals (TTYs and PTYs) handle input, output, and line discipline
// features like canonical mode, echo, signals, and baud rates.
//
// Essentially, [termios] settings control how the PTY slave, used by our web terminal, processes input and output data.
// It also affects how the slave buffers input, handles special chars (like Ctrl-C), line editing, etc. In canonical
// mode, the terminal processes input line-by-line, meaning it waits for a newline character before sending the data to
// the application. The maximum line length is 4096 characters, any input longer than that is truncated.
//
// [termios] documentation says:
//
//	The maximum line length is 4096 chars (including the
//	terminating newline character); lines longer than 4096 chars
//	are truncated.  After 4095 characters, input processing (e.g.,
//	ISIG and ECHO* processing) continues, but any input data after
//	4095 characters up to (but not including) any terminating
//	newline is discarded.  This ensures that the terminal can
//	always receive more input until at least one line can be read.
//
// [termios]: https://www.man7.org/linux/man-pages/man3/termios.3.html
const TermniosMaxLineLength = 4096

// ReadMessageBufferSize is the size of the buffer used to read messages from the websocket connection.
//
// As we read JSON messages from the websocket connection, we need to ensure that the buffer size is large enough
// so, we have decided to use a buffer size of 16404 bytes, which is the [TermniosMaxLineLength] plus the size of the
// minimum message size [MessageMinSize].
const ReadMessageBufferSize = MessageMinSize + (TermniosMaxLineLength * CharacterSize)

func (c *Conn) ReadMessage(message *Message) (int, error) {
	limit := io.LimitReader(c.Socket, ReadMessageBufferSize)
	decoder := json.NewDecoder(limit)

	var data json.RawMessage
	message.Data = &data

	if err := decoder.Decode(message); err != nil {
		log.WithError(err).Error("failed to read a line from the websocket connection")

		return 0, errors.Join(ErrConnReadMessageJSONInvalid, err)
	}

	switch message.Kind {
	case messageKindInput:
		var str string

		if err := json.Unmarshal(data, &str); err != nil {
			return 0, errors.Join(ErrConnReadMessageJSONInvalid)
		}

		message.Data = str
	case messageKindResize:
		var dim Dimensions

		if err := json.Unmarshal(data, &dim); err != nil {
			return 0, errors.Join(ErrConnReadMessageJSONInvalid)
		}

		message.Data = dim
	case messageKindSignature:
		var sig string

		if err := json.Unmarshal(data, &sig); err != nil {
			return 0, errors.Join(ErrConnReadMessageJSONInvalid)
		}

		message.Data = sig
	default:
		return 0, errors.Join(ErrConnReadMessageKindInvalid)
	}

	return int(decoder.InputOffset()), nil
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

func (c *Conn) WriteBinary(data []byte) (int, error) {
	socket := c.Socket.(*websocket.Conn)

	frame, err := socket.NewFrameWriter(websocket.BinaryFrame)
	if err != nil {
		return 0, errors.Join(ErrConnWriteMessageFailedFrame, err)
	}

	wrote, err := frame.Write(data)
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
