package srdp

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/shellhub-io/shellhub/pkg/srdp/displays"
	"github.com/shellhub-io/shellhub/pkg/srdp/encoders"
	log "github.com/sirupsen/logrus"
)

type Connection struct {
	connID  uint64
	conn    io.ReadWriteCloser
	reader  *bufio.Reader
	writer  *bufio.Writer
	width   uint16
	height  uint16
	fps     uint32
	encoder encoders.Encoder
	display displays.Display
	logger  *log.Entry

	version string
}

func NewConnection(conn io.ReadWriteCloser, display displays.Display, logger *log.Entry) (*Connection, error) {
	width, height := display.GetScreenSize()

	if logger == nil {
		logger = log.NewEntry(log.StandardLogger())
	}

	var fps uint32 = 30 // Target frames per second

	return &Connection{
		connID:  uint64(time.Now().UnixNano() % 0xFFFFFFFF), //nolint:gosec
		conn:    conn,
		reader:  bufio.NewReader(conn),
		writer:  bufio.NewWriter(conn),
		width:   width,
		height:  height,
		fps:     fps,
		display: display,
		logger:  logger,
		// TODO: Add configurable version.
		version: "SRDP 000.001\n",
	}, nil
}

// exhangeVersion handles the initial protocol version exchange
func (c *Connection) exhangeVersion(version string) error {
	// Send protocol version
	if _, err := c.writer.WriteString(version); err != nil {
		return err
	}

	if err := c.writer.Flush(); err != nil {
		return err
	}

	// Read client protocol version
	clientVersion := make([]byte, 13)
	if _, err := io.ReadFull(c.reader, clientVersion); err != nil {
		return err
	}

	if string(clientVersion[:]) != version {
		return fmt.Errorf("unsupported protocol version: %s", string(clientVersion))
	}

	return nil
}

const (
	// SecurityTypeResponseOK indicates successful security negotiation.
	//
	// https://www.rfc-editor.org/rfc/rfc6143.html#section-7.1.3
	SecurityTypeResponseOK = 0
	// SecurityTypeResponseFailed indicates failed security negotiation.
	//
	// https://www.rfc-editor.org/rfc/rfc6143.html#section-7.1.3
	SecurityTypeResponseFailed = 1
)

// securityHandshake handles security type negotiation
func (c *Connection) securityHandshake(auths []Auth) error {
	// Send number of security types
	n := len(auths)
	if _, err := c.writer.Write([]byte{byte(n)}); err != nil {
		return err
	}

	c.logger.Tracef("Supported security types: %d", n)

	var types []byte
	for _, auth := range auths {
		types = append(types, auth.SecurityType(c))
	}

	// Send security types
	if _, err := c.writer.Write(types); err != nil {
		return err
	}

	c.logger.Tracef("Sent security types: %v", types)

	if err := c.writer.Flush(); err != nil {
		return err
	}

	c.logger.Trace("Waiting for client to select security type")

	// Read client's chosen security type
	var clientSecurityType [1]byte
	if _, err := io.ReadFull(c.reader, clientSecurityType[:]); err != nil {
		return err
	}

	c.logger.Tracef("Client selected security type: %d", clientSecurityType[0])

	securityResult := make([]byte, 4)

	switch clientSecurityType[0] {
	case SecurityTypeNone:
	case SecurityTypeSRDPAuth:
		// Send security result (failed)
		binary.BigEndian.PutUint32(securityResult, SecurityTypeResponseFailed) // 1 = failed
		if _, err := c.writer.Write(securityResult); err != nil {
			return err
		}

		return fmt.Errorf("SRDP authentication not implemented")
	default:
		binary.BigEndian.PutUint32(securityResult, SecurityTypeResponseFailed) // 1 = failed
		if _, err := c.writer.Write(securityResult); err != nil {
			return err
		}

		return fmt.Errorf("unsupported security type: %d", clientSecurityType[0])
	}

	// Send security result (OK)
	binary.BigEndian.PutUint32(securityResult, SecurityTypeResponseOK) // 0 = OK
	if _, err := c.writer.Write(securityResult); err != nil {
		return err
	}

	return c.writer.Flush()
}

func (c *Connection) clientInit() error {
	buf := make([]byte, 1)
	if _, err := io.ReadFull(c.reader, buf); err != nil {
		return err
	}

	numEncoders := buf[0] // Number of supported encoders
	c.logger.Tracef("Client supports %d encoders", numEncoders)

	encodersList := make([]byte, numEncoders) // List of supported encoders
	if _, err := io.ReadFull(c.reader, encodersList); err != nil {
		return err
	}

	c.logger.Tracef("Client encoders: %v", encodersList)

	for _, enc := range encodersList {
		switch enc {
		// NOTE: only H.264 is supported for now.
		// TODO: Add property to connection with all supported encoder by the device, filling it when creating the SRDP
		// server on Agent.
		case EncoderH264:
			encoder, err := encoders.NewH264(int(c.width), int(c.height), c.fps)
			if err != nil {
				return err
			}

			c.encoder = encoder

			c.logger.Infof("Selected encoder: H.264")
		default:
			c.logger.Warnf("Unsupported encoder from client: %d", enc)
		}
	}

	if c.encoder == nil {
		c.writer.WriteByte(EncoderInvalid)
	} else {
		c.writer.WriteByte(byte(c.encoder.Code()))
	}

	return c.writer.Flush()
}

// serverInit sends the server initialization message
func (c *Connection) serverInit() error {
	buf := make([]byte, 4)

	binary.BigEndian.PutUint16(buf[0:2], c.width)
	binary.BigEndian.PutUint16(buf[2:4], c.height)
	if _, err := c.writer.Write(buf); err != nil {
		return err
	}

	return c.writer.Flush()
}

func (c *Connection) Close() error {
	return c.conn.Close()
}

func (c *Connection) sendVideoFrame() error {
	data, err := c.display.CaptureScreen()
	if err != nil {
		return err
	}

	encoded, err := c.encoder.Encode(c.width, c.height, data)
	if err != nil {
		return fmt.Errorf("H.264 encoding failed: %v", err)
	}

	c.writer.WriteByte(ServerFrameUpdate) // Message type

	header := make([]byte, 8)
	binary.BigEndian.PutUint64(header[0:8], uint64(len(encoded))) // Message length
	if _, err := c.writer.Write(header); err != nil {
		return err
	}

	if _, err := c.writer.Write(encoded); err != nil { // Encoded pixel data.
		return err
	}

	return c.writer.Flush()
}

func (c *Connection) startVideoStream() {
	ticker := time.NewTicker(1000 / time.Duration(c.fps))
	defer ticker.Stop()
	defer c.Close()

	for range ticker.C {
		if err := c.sendVideoFrame(); err != nil {
			if !errors.Is(err, io.EOF) {
				c.logger.Errorf("handle framebuffer update request error: %v", err)

				return
			}

			return
		}
	}
}

func (c *Connection) loop() error {
	c.logger.Info("Entering message loop")
	defer c.logger.Info("Exiting message loop")

	go c.startVideoStream()

	for {
		var msgType [1]byte
		if _, err := io.ReadFull(c.reader, msgType[:]); err != nil {
			if err == io.EOF {
				return nil // Client disconnected
			}

			return err
		}

		switch msgType[0] {
		case ClientKeyEvent:
			c.logger.Trace("Received key event from client")

			if err := c.handleKeyEvent(); err != nil {
				return err
			}
		case ClientPointerEvent:
			c.logger.Trace("Received pointer event from client")

			if err := c.handlePointerEvent(); err != nil {
				return err
			}
		default:
			c.logger.Warnf("Unknown client message type: %d", msgType[0])

			return fmt.Errorf("unknown client message type: %d", msgType[0])
		}
	}
}

func (c *Connection) handleKeyEvent() error {
	var event [7]byte
	if _, err := io.ReadFull(c.reader, event[:]); err != nil {
		return err
	}

	downFlag := event[0]
	key := binary.BigEndian.Uint32(event[3:7])

	return c.display.SendKeyEvent(key, downFlag != 0)
}

func (c *Connection) handlePointerEvent() error {
	var event [5]byte
	if _, err := io.ReadFull(c.reader, event[:]); err != nil {
		return err
	}

	buttonMask := event[0]
	x := binary.BigEndian.Uint16(event[1:3])
	y := binary.BigEndian.Uint16(event[3:5])

	return c.display.SendPointerEvent(x, y, buttonMask)
}
