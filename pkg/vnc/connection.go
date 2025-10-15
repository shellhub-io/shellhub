package vnc

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"time"
)

// Connection represents a VNC client connection
type Connection struct {
	conn        io.ReadWriteCloser
	reader      *bufio.Reader
	writer      *bufio.Writer
	pixelFormat PixelFormat
	encodings   []int32
	width       uint16
	height      uint16
	framebuffer []uint8
	lastUpdate  time.Time
	display     Display
	encodingMgr *EncodingManager
}

// Display interface for screen capture
type Display interface {
	GetScreenSize() (width, height uint16)
	CaptureScreen() ([]uint8, error)
	SendKeyEvent(key uint32, down bool) error
	SendPointerEvent(x, y uint16, buttonMask uint8) error
}

// NewConnection creates a new VNC connection handler
func NewConnection(conn io.ReadWriteCloser, display Display) *Connection {
	width, height := display.GetScreenSize()
	return &Connection{
		conn:        conn,
		reader:      bufio.NewReader(conn),
		writer:      bufio.NewWriter(conn),
		pixelFormat: DefaultPixelFormat(),
		width:       width,
		height:      height,
		display:     display,
		encodingMgr: NewEncodingManager(),
	}
}

// protocolVersionHandshake handles the initial protocol version exchange
func (c *Connection) protocolVersionHandshake(version string) error {
	// Send protocol version
	if _, err := c.writer.WriteString(version); err != nil {
		return err
	}

	if err := c.writer.Flush(); err != nil {
		return err
	}

	// Read client protocol version
	clientVersion := make([]byte, 12)
	if _, err := io.ReadFull(c.reader, clientVersion); err != nil {
		return err
	}

	// For simplicity, accept any 3.x version
	if string(clientVersion[:7]) != "RFB 003" {
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

	log.Printf("Supported security types: %d", n)

	var types []byte
	for _, auth := range auths {
		types = append(types, auth.SecurityType(c))
	}

	// Send security types
	if _, err := c.writer.Write(types); err != nil {
		return err
	}

	log.Printf("Sent security types: %v", types)

	if err := c.writer.Flush(); err != nil {
		return err
	}

	log.Printf("Waiting for client to select security type")

	// Read client's chosen security type
	var clientSecurityType [1]byte
	if _, err := io.ReadFull(c.reader, clientSecurityType[:]); err != nil {
		return err
	}

	fmt.Println("Client selected security type:", clientSecurityType[0])

	securityResult := make([]byte, 4)

	switch clientSecurityType[0] {
	case SecurityTypeNone:
	case SecurityTypeVNCAuth:
		// Send security result (failed)
		binary.BigEndian.PutUint32(securityResult, SecurityTypeResponseFailed) // 1 = failed
		if _, err := c.writer.Write(securityResult); err != nil {
			return err
		}

		return fmt.Errorf("VNC authentication not implemented")
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

// clientInitialization reads the client initialization message
func (c *Connection) clientInitialization() error {
	var shared [1]byte
	if _, err := io.ReadFull(c.reader, shared[:]); err != nil {
		return err
	}

	// We ignore the shared flag for simplicity
	// TODO: Handle shared flag if needed

	return nil
}

// serverInitialization sends the server initialization message
func (c *Connection) serverInitialization(name string) error {
	serverName := name

	// Prepare server init message
	buf := make([]byte, 24+len(serverName))

	// Framebuffer width and height
	binary.BigEndian.PutUint16(buf[0:2], c.width)
	binary.BigEndian.PutUint16(buf[2:4], c.height)

	// Server pixel format
	pixelFormatBytes := c.pixelFormat.WriteBytes()
	copy(buf[4:20], pixelFormatBytes)

	// Name length and name
	binary.BigEndian.PutUint32(buf[20:24], uint32(len(serverName)))
	copy(buf[24:], []byte(serverName))

	if _, err := c.writer.Write(buf); err != nil {
		return err
	}

	return c.writer.Flush()
}

// messageLoop handles client messages
func (c *Connection) messageLoop() error {
	for {
		var msgType [1]byte
		if _, err := io.ReadFull(c.reader, msgType[:]); err != nil {
			if err == io.EOF {
				return nil // Client disconnected
			}

			return err
		}

		switch msgType[0] {
		case ClientSetPixelFormat:
			if err := c.handleSetPixelFormat(); err != nil {
				return err
			}
		case ClientSetEncodings:
			if err := c.handleSetEncodings(); err != nil {
				return err
			}
		case ClientFramebufferUpdateRequest:
			if err := c.handleFramebufferUpdateRequest(); err != nil {
				return err
			}
		case ClientKeyEvent:
			if err := c.handleKeyEvent(); err != nil {
				return err
			}
		case ClientPointerEvent:
			if err := c.handlePointerEvent(); err != nil {
				return err
			}
		case ClientCutText:
			if err := c.handleCutText(); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown client message type: %d", msgType[0])
		}
	}
}

// handleSetPixelFormat processes pixel format changes
func (c *Connection) handleSetPixelFormat() error {
	// Skip padding
	padding := make([]byte, 3)
	if _, err := io.ReadFull(c.reader, padding); err != nil {
		return err
	}

	// Read pixel format
	pfBytes := make([]byte, 16)
	if _, err := io.ReadFull(c.reader, pfBytes); err != nil {
		return err
	}

	pf, err := ReadPixelFormat(pfBytes)
	if err != nil {
		return err
	}

	c.pixelFormat = *pf

	return nil
}

// handleSetEncodings processes encoding preferences
func (c *Connection) handleSetEncodings() error {
	// Skip padding
	var padding [1]byte
	if _, err := io.ReadFull(c.reader, padding[:]); err != nil {
		return err
	}

	// Read number of encodings
	var numEncodings [2]byte
	if _, err := io.ReadFull(c.reader, numEncodings[:]); err != nil {
		return err
	}

	count := binary.BigEndian.Uint16(numEncodings[:])

	// Read encodings
	c.encodings = make([]int32, count)
	for i := range count {
		var encoding [4]byte
		if _, err := io.ReadFull(c.reader, encoding[:]); err != nil {
			return err
		}

		c.encodings[i] = int32(binary.BigEndian.Uint32(encoding[:]))
	}

	return nil
}

// handleFramebufferUpdateRequest processes screen update requests
func (c *Connection) handleFramebufferUpdateRequest() error {
	var req [9]byte
	if _, err := io.ReadFull(c.reader, req[:]); err != nil {
		return err
	}

	incremental := req[0]
	x := binary.BigEndian.Uint16(req[1:3])
	y := binary.BigEndian.Uint16(req[3:5])
	width := binary.BigEndian.Uint16(req[5:7])
	height := binary.BigEndian.Uint16(req[7:9])

	// Send framebuffer update
	return c.sendFramebufferUpdate(incremental != 0, x, y, width, height)
}

// sendFramebufferUpdate sends screen content to client
func (c *Connection) sendFramebufferUpdate(incremental bool, x, y, width, height uint16) error {
	// Capture screen content
	screenData, err := c.display.CaptureScreen()
	if err != nil {
		return err
	}

	// For simplicity, always send the full screen
	actualWidth := c.width
	actualHeight := c.height
	if width > 0 && height > 0 {
		actualWidth = width
		actualHeight = height
	}

	// Prepare framebuffer update header
	header := make([]byte, 4)
	header[0] = ServerFramebufferUpdate
	header[1] = 0                              // padding
	binary.BigEndian.PutUint16(header[2:4], 1) // number of rectangles

	if _, err := c.writer.Write(header); err != nil {
		return err
	}

	// Encode pixel data using the best available encoding
	encodedData, encodingType, err := c.encodingMgr.EncodeRectangle(
		screenData, actualWidth, actualHeight, c.pixelFormat, c.encodings)
	if err != nil {
		return fmt.Errorf("encoding failed: %v", err)
	}

	// Rectangle header
	rectHeader := make([]byte, 12)
	binary.BigEndian.PutUint16(rectHeader[0:2], x)
	binary.BigEndian.PutUint16(rectHeader[2:4], y)
	binary.BigEndian.PutUint16(rectHeader[4:6], actualWidth)
	binary.BigEndian.PutUint16(rectHeader[6:8], actualHeight)
	binary.BigEndian.PutUint32(rectHeader[8:12], uint32(encodingType))

	if _, err := c.writer.Write(rectHeader); err != nil {
		return err
	}

	// Send encoded pixel data
	if _, err := c.writer.Write(encodedData); err != nil {
		return err
	}

	return c.writer.Flush()
}

// handleKeyEvent processes keyboard events
func (c *Connection) handleKeyEvent() error {
	var event [7]byte
	if _, err := io.ReadFull(c.reader, event[:]); err != nil {
		return err
	}

	downFlag := event[0]
	key := binary.BigEndian.Uint32(event[3:7])

	return c.display.SendKeyEvent(key, downFlag != 0)
}

// handlePointerEvent processes mouse events
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

// handleCutText processes clipboard events
func (c *Connection) handleCutText() error {
	// Skip padding
	padding := make([]byte, 3)
	if _, err := io.ReadFull(c.reader, padding); err != nil {
		return err
	}

	// Read text length
	var length [4]byte
	if _, err := io.ReadFull(c.reader, length[:]); err != nil {
		return err
	}
	textLength := binary.BigEndian.Uint32(length[:])

	// Read and ignore the text for now
	text := make([]byte, textLength)
	if _, err := io.ReadFull(c.reader, text); err != nil {
		return err
	}

	// TODO: Handle clipboard integration
	return nil
}
