package rfb

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"time"

	log "github.com/sirupsen/logrus"
)

// Connection represents a RFB client connection
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
	connID      uint32 // Unique connection ID for encoder state
	logger      *log.Entry
}

// Display interface for screen capture
type Display interface {
	GetScreenSize() (width, height uint16)
	CaptureScreen() ([]uint8, error)
	SendKeyEvent(key uint32, down bool) error
	SendPointerEvent(x, y uint16, buttonMask uint8) error
}

// NewConnection creates a new RFB connection handler
func NewConnection(conn io.ReadWriteCloser, display Display, logger *log.Entry) *Connection {
	width, height := display.GetScreenSize()

	if logger == nil {
		logger = log.NewEntry(log.StandardLogger())
	}

	encodingMgr := NewEncodingManager()
	connID := uint32(time.Now().UnixNano() % 0xFFFFFFFF) // Generate unique connection ID

	return &Connection{
		conn:        conn,
		reader:      bufio.NewReader(conn),
		writer:      bufio.NewWriter(conn),
		pixelFormat: DefaultPixelFormat(),
		width:       width,
		height:      height,
		display:     display,
		encodingMgr: encodingMgr,
		connID:      connID,
		logger:      logger,
	}
}

// Close cleans up connection resources, including encoder state
func (c *Connection) Close() error {
	if c.encodingMgr != nil {
		c.encodingMgr.CloseConnection(c.connID)
	}

	return nil
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
	case SecurityTypeRFBAuth:
		// Send security result (failed)
		binary.BigEndian.PutUint32(securityResult, SecurityTypeResponseFailed) // 1 = failed
		if _, err := c.writer.Write(securityResult); err != nil {
			return err
		}

		return fmt.Errorf("RFB authentication not implemented")
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
	defer c.logger.Trace("Exiting message loop")

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

	// Create per-connection encoder instances on first SetEncodings message
	if err := c.encodingMgr.CreateConnectionEncoders(c.connID); err != nil {
		c.logger.Errorf("Failed to create connection encoders: %v", err)

		return err
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

	// Initialize framebuffer on first update
	if c.framebuffer == nil {
		c.framebuffer = make([]uint8, len(screenData))
		copy(c.framebuffer, screenData)
		incremental = false // Force full update on first request
	}

	// Determine regions to update
	var rects []Rectangle
	if !incremental {
		// Full update: send the entire screen
		rects = append(rects, Rectangle{
			XPosition: 0,
			YPosition: 0,
			Width:     c.width,
			Height:    c.height,
		})
	} else {
		// Incremental update: detect changed regions
		changedRects := c.detectChangedRegions(screenData, x, y, width, height)
		rects = changedRects
		if len(rects) == 0 {
			// No changes detected, send empty framebuffer update
			header := make([]byte, 4)
			header[0] = ServerFramebufferUpdate        // message type
			header[1] = 0                              // padding
			binary.BigEndian.PutUint16(header[2:4], 0) // number of rectangles

			if _, err := c.writer.Write(header); err != nil {
				return err
			}

			return c.writer.Flush()
		}
	}

	// Prepare framebuffer update header with actual number of rectangles
	header := make([]byte, 4)
	header[0] = ServerFramebufferUpdate                         // message type
	header[1] = 0                                               // padding
	binary.BigEndian.PutUint16(header[2:4], uint16(len(rects))) // number of rectangles

	if _, err := c.writer.Write(header); err != nil {
		return err
	}

	// Send each rectangle
	for _, rect := range rects {
		// Extract the rectangle data from the full screen
		rectData := c.extractRectangleData(screenData, rect.XPosition, rect.YPosition, rect.Width, rect.Height)

		// Encode pixel data for this rectangle using the best available encoding
		encodedData, encodingType, err := c.encodingMgr.EncodeRectangle(
			c.connID,
			rectData,
			rect.Width,
			rect.Height,
			c.pixelFormat,
			c.encodings,
		)
		if err != nil {
			return fmt.Errorf("encoding failed: %v", err)
		}

		// Rectangle header
		rectHeader := make([]byte, 12)
		binary.BigEndian.PutUint16(rectHeader[0:2], rect.XPosition)
		binary.BigEndian.PutUint16(rectHeader[2:4], rect.YPosition)
		binary.BigEndian.PutUint16(rectHeader[4:6], rect.Width)
		binary.BigEndian.PutUint16(rectHeader[6:8], rect.Height)
		binary.BigEndian.PutUint32(rectHeader[8:12], uint32(encodingType))

		if _, err := c.writer.Write(rectHeader); err != nil {
			return err
		}

		// Send encoded pixel data
		if _, err := c.writer.Write(encodedData); err != nil {
			return err
		}
	}

	// Update cached framebuffer
	if len(c.framebuffer) == len(screenData) {
		copy(c.framebuffer, screenData)
	}

	c.lastUpdate = time.Now()

	return c.writer.Flush()
}

// extractRectangleData extracts pixel data for a specific rectangle from the full framebuffer
func (c *Connection) extractRectangleData(fullData []uint8, x, y, width, height uint16) []uint8 {
	bytesPerPixel := (c.pixelFormat.BitsPerPixel + 7) / 8
	stride := int(c.width) * int(bytesPerPixel)

	// Allocate buffer for the rectangle
	rectSize := int(width) * int(height) * int(bytesPerPixel)
	rectData := make([]uint8, rectSize)

	// Copy the rectangle data row by row
	for row := uint16(0); row < height; row++ {
		srcOffset := int(y+row)*stride + int(x)*int(bytesPerPixel)
		dstOffset := int(row) * int(width) * int(bytesPerPixel)
		copyLen := int(width) * int(bytesPerPixel)

		// Validate source offset and length
		if srcOffset < 0 || srcOffset+copyLen > len(fullData) {
			// If we can't copy this row, fill with zeros (or we could skip)
			continue
		}

		copy(rectData[dstOffset:dstOffset+copyLen], fullData[srcOffset:srcOffset+copyLen])
	}

	return rectData
}

// detectChangedRegions detects regions that have changed between old and new framebuffer
func (c *Connection) detectChangedRegions(newData []uint8, reqX, reqY, reqWidth, reqHeight uint16) []Rectangle {
	if len(c.framebuffer) != len(newData) {
		// Size mismatch, force full screen update
		return []Rectangle{{XPosition: 0, YPosition: 0, Width: c.width, Height: c.height}}
	}

	// Bytes per pixel based on pixel format
	bytesPerPixel := (c.pixelFormat.BitsPerPixel + 7) / 8
	stride := int(c.width) * int(bytesPerPixel)

	var rects []Rectangle
	const TILE_SIZE uint16 = 16 // Tile size for change detection

	// Scan entire screen in tiles to find changed regions
	for tileY := uint16(0); tileY < c.height; tileY += TILE_SIZE {
		for tileX := uint16(0); tileX < c.width; tileX += TILE_SIZE {
			x := tileX
			y := tileY
			w := TILE_SIZE
			h := TILE_SIZE

			// Clamp tile to screen boundaries
			if x+w > c.width {
				w = c.width - x
			}
			if y+h > c.height {
				h = c.height - y
			}

			// Check if this tile has changed
			if c.isTileChanged(newData, x, y, w, h, stride, bytesPerPixel) {
				rects = append(rects, Rectangle{XPosition: x, YPosition: y, Width: w, Height: h})
			}
		}
	}

	return rects
}

// isTileChanged checks if a tile region has changed between old and new framebuffer
func (c *Connection) isTileChanged(newData []uint8, x, y, width, height uint16, stride int, bytesPerPixel uint8) bool {
	for row := uint16(0); row < height; row++ {
		// Calculate byte offset for this row in both old and new framebuffers
		offset := int(y+row)*stride + int(x)*int(bytesPerPixel)
		rowSize := int(width) * int(bytesPerPixel)

		// Validate offset is non-negative
		if offset < 0 {
			return true // Invalid offset, consider changed
		}

		// Check bounds before comparing
		if offset+rowSize > len(c.framebuffer) || offset+rowSize > len(newData) {
			// If we can't fully compare, consider it changed to be safe
			return true
		}

		// Compare this row of pixels
		for i := 0; i < rowSize; i++ {
			if c.framebuffer[offset+i] != newData[offset+i] {
				return true
			}
		}
	}

	return false
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
