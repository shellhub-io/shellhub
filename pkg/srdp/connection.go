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

const (
	DefaultFPS uint16 = 30 // Default frames per second

	// FPS constraints enforced by server
	MaxFPS uint16 = 60 // Maximum frames per second (clamped on server)
	MinFPS uint16 = 1  // Minimum frames per second (clamped on server)

	// Client options validation
	MinClientOptions = 0 // Minimum allowed client options (0 = defaults used)
	MaxClientOptions = 3 // Maximum allowed client options (WindowSize, FPS, Encoder)
)

type Options struct {
	Width    uint16
	Height   uint16
	FPS      uint16
	Encoders []byte
}

type Connection struct {
	version string
	connID  uint64
	conn    io.ReadWriteCloser
	reader  *bufio.Reader
	writer  *bufio.Writer
	encoder encoders.Encoder
	display displays.Display
	options Options

	logger *log.Entry
}

func NewConnection(conn io.ReadWriteCloser, display displays.Display, logger *log.Entry) (*Connection, error) {
	if logger == nil {
		logger = log.NewEntry(log.StandardLogger())
	}

	return &Connection{
		// TODO: Add configurable version.
		version: "SRDP 000.001\n",
		connID:  uint64(time.Now().UnixNano() % 0xFFFFFFFF), //nolint:gosec
		conn:    conn,
		reader:  bufio.NewReader(conn),
		writer:  bufio.NewWriter(conn),
		display: display,

		logger: logger,
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
	optionsBuf := make([]byte, 1)
	if _, err := io.ReadFull(c.reader, optionsBuf); err != nil {
		return err
	}

	numOptions := optionsBuf[0]
	c.logger.Tracef("Client sent %d options", numOptions)

	if numOptions == 0 {
		c.logger.Info("No client options sent; using default settings")

		return nil
	}

	if numOptions < MinClientOptions || numOptions > MaxClientOptions {
		return fmt.Errorf("invalid number of client options: %d", numOptions)
	}

	for i := 0; i < int(numOptions); i++ {
		optTypeBuf := make([]byte, 1)
		if _, err := io.ReadFull(c.reader, optTypeBuf); err != nil {
			return err
		}

		optType := optTypeBuf[0]
		c.logger.Tracef("Processing client option type: %d", optType)

		switch optType {
		case ClientOptionWindowSize:
			sizeBuf := make([]byte, 4)
			if _, err := io.ReadFull(c.reader, sizeBuf); err != nil {
				return err
			}

			// NOTE: For now, always use display screen even if client did not request it.
			// width := binary.BigEndian.Uint16(sizeBuf[0:2])
			// height := binary.BigEndian.Uint16(sizeBuf[2:4])
			// c.options.Width = width
			// c.options.Height = height

			c.logger.Infof("Client requested window size: %dx%d", c.options.Width, c.options.Height)
		case ClientOptionFPS:
			fpsBuf := make([]byte, 2)
			if _, err := io.ReadFull(c.reader, fpsBuf); err != nil {
				return err
			}

			fps := binary.BigEndian.Uint16(fpsBuf)
			c.logger.Tracef("Client requested FPS: %d", fps)

			if fps > MaxFPS {
				fps = MaxFPS
			} else if fps < MinFPS {
				fps = MinFPS
			}

			c.options.FPS = fps

			c.logger.Infof("Client requested FPS set to: %d", c.options.FPS)
		case ClientOptionEncoder:
			encodersBuf := make([]byte, 1)
			if _, err := io.ReadFull(c.reader, encodersBuf); err != nil {
				return err
			}

			encoderNums := encodersBuf[0]
			if encoderNums <= 0 {
				c.logger.Warn("Client sent invalid number of encoders; using default encoder (H.264)")

				break
			}

			encodersList := make([]byte, encoderNums)
			if _, err := io.ReadFull(c.reader, encodersList); err != nil {
				return err
			}

			if encoderNums == 0 {
				c.logger.Info("No encoder preference sent by client; using default encoder (H.264)")

				break
			}

			c.logger.Tracef("Client supported encoders: %v", encodersList)

			c.options.Encoders = encodersList
		default:
			c.logger.Warnf("Unknown client option type: %d", optType)
		}
	}

	return c.writer.Flush()
}

func (c *Connection) serverInit() error {
	buf := make([]byte, 7)

	if c.options.Width == 0 || c.options.Height == 0 {
		width, height := c.display.GetScreenSize()
		c.logger.Infof("Using display screen size: %dx%d", width, height)

		c.options.Width = width
		c.options.Height = height
	}

	if c.options.FPS == 0 {
		c.options.FPS = DefaultFPS
	}

	if len(c.options.Encoders) == 0 {
		c.logger.Info("No encoder preference sent by client; using default encoder (H.264)")

		c.options.Encoders = []byte{EncoderH264}
	}

	for _, encoderType := range c.options.Encoders {
		switch encoderType {
		case EncoderH264:
			var err error

			c.encoder, err = encoders.NewH264(int(c.options.Width), int(c.options.Height), c.options.FPS)
			if err != nil {
				return err
			}

			c.logger.Infof("Using encoder: H.264")
		}
	}

	c.logger.WithFields(log.Fields{
		"width":   c.options.Width,
		"height":  c.options.Height,
		"fps":     c.options.FPS,
		"encoder": c.encoder.Code(),
	}).Info("Server initialized with options")

	binary.BigEndian.PutUint16(buf[0:2], c.options.Width)
	binary.BigEndian.PutUint16(buf[2:4], c.options.Height)
	binary.BigEndian.PutUint16(buf[4:6], c.options.FPS)
	buf[6] = byte(c.encoder.Code()) // Encoder type
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

	encoded, err := c.encoder.Encode(c.options.Width, c.options.Height, data)
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
	ticker := time.NewTicker(1000 / time.Duration(c.options.FPS))
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
