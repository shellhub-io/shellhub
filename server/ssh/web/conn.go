package web

import (
	"encoding/json"
	"errors"
	"io"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/shellhub-io/shellhub/pkg/clock"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/websocket"
)

// Socket is the transport under a [Conn]. It is an interface so tests can drive a connection
// without a real WebSocket.
type Socket interface {
	io.ReadWriteCloser
}

// Conn is the browser's end of a web terminal session. Writes are serialised, because the
// terminal, the ping loop and the signing exchange all write to the same socket.
type Conn struct {
	// Socket is the internal websocket connection the messages come from.
	Socket Socket
	// Pinger is reponsable to inform the server that a SSH session is open.
	Pinger *time.Ticker

	writes sync.Mutex
}

// NewConn wraps socket, starting the ticker that keeps the session marked as live.
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

// ReadMessage decodes the next message into message, returning the bytes consumed. The read
// is bounded by [ReadMessageBufferSize] so that a client cannot force unbounded buffering.
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

		if utf8.RuneCountInString(str) > TermniosMaxLineLength {
			return 0, errors.Join(ErrConnReadMessageInputTooLong)
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

// WriteMessage sends message as JSON, returning the bytes written.
func (c *Conn) WriteMessage(message *Message) (int, error) {
	c.writes.Lock()
	defer c.writes.Unlock()

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

// WriteBinary sends data as a binary frame, which is how terminal output is carried. Non-
// WebSocket sockets fall back to a plain write.
func (c *Conn) WriteBinary(data []byte) (int, error) {
	c.writes.Lock()
	defer c.writes.Unlock()

	socket, ok := c.Socket.(*websocket.Conn)
	if !ok {
		return c.Socket.Write(data)
	}

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

// WritePing sends a ping frame, telling the browser the session is still open.
func (c *Conn) WritePing() error {
	c.writes.Lock()
	defer c.writes.Unlock()

	socket, ok := c.Socket.(*websocket.Conn)
	if !ok {
		return nil
	}

	frame, err := socket.NewFrameWriter(websocket.PingFrame)
	if err != nil {
		return err
	}

	if _, err := frame.Write([]byte{}); err != nil {
		return err
	}

	return nil
}

func (c *Conn) Read(buffer []byte) (int, error) {
	return c.Socket.Read(buffer)
}

func (c *Conn) Write(buffer []byte) (int, error) {
	c.writes.Lock()
	defer c.writes.Unlock()

	return c.Socket.Write(buffer)
}

// Close stops the ping ticker and closes the socket.
func (c *Conn) Close() error {
	c.Pinger.Stop()

	return c.Socket.Close()
}

// KeepAlive pings on every tick until the connection is closed. It is meant to be run in its
// own goroutine.
func (c *Conn) KeepAlive() {
	socket, ok := c.Socket.(*websocket.Conn)
	if !ok {
		return
	}

	for {
		if err := socket.SetDeadline(clock.Now().Add((time.Second * 30) * 2)); err != nil {
			return
		}

		if err := c.WritePing(); err != nil {
			return
		}

		if _, running := <-c.Pinger.C; !running {
			return
		}
	}
}
