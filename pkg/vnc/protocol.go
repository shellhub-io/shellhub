package vnc

import (
	"encoding/binary"
	"fmt"
)

// VNC Protocol Constants
const (
	// Protocol Version
	ProtocolVersionMajor = 3
	ProtocolVersionMinor = 8
	ProtocolVersion      = "RFB 003.008\n"

	// Security Types
	SecurityTypeNone     = 1
	SecurityTypeVNCAuth  = 2
	SecurityTypeTight    = 16
	SecurityTypeUltra    = 17
	SecurityTypeTLS      = 18
	SecurityTypeVeNCrypt = 19

	// Client to Server Messages
	ClientSetPixelFormat           = 0
	ClientSetEncodings             = 2
	ClientFramebufferUpdateRequest = 3
	ClientKeyEvent                 = 4
	ClientPointerEvent             = 5
	ClientCutText                  = 6

	// Server to Client Messages
	ServerFramebufferUpdate   = 0
	ServerSetColourMapEntries = 1
	ServerBell                = 2
	ServerCutText             = 3

	// Encoding Types
	EncodingRaw      = 0
	EncodingCopyRect = 1
	EncodingRRE      = 2
	EncodingHextile  = 5
	EncodingTRLE     = 15
	EncodingZRLE     = 16

	// Pseudo-encodings
	EncodingDesktopSize = -223
	EncodingCursor      = -239
	EncodingXCursor     = -240
)

// PixelFormat represents the VNC pixel format
type PixelFormat struct {
	BitsPerPixel   uint8
	Depth          uint8
	BigEndianFlag  uint8
	TrueColourFlag uint8
	RedMax         uint16
	GreenMax       uint16
	BlueMax        uint16
	RedShift       uint8
	GreenShift     uint8
	BlueShift      uint8
	Padding        [3]uint8
}

// ServerInit message sent to client after handshake
type ServerInit struct {
	FramebufferWidth  uint16
	FramebufferHeight uint16
	ServerPixelFormat PixelFormat
	NameLength        uint32
	Name              string
}

// Rectangle represents a screen rectangle
type Rectangle struct {
	XPosition uint16
	YPosition uint16
	Width     uint16
	Height    uint16
	Encoding  int32
}

// KeyEvent represents a keyboard event
type KeyEvent struct {
	DownFlag uint8
	Padding  [2]uint8
	Key      uint32
}

// PointerEvent represents a mouse event
type PointerEvent struct {
	ButtonMask uint8
	XPosition  uint16
	YPosition  uint16
}

// FramebufferUpdateRequest from client
type FramebufferUpdateRequest struct {
	Incremental uint8
	XPosition   uint16
	YPosition   uint16
	Width       uint16
	Height      uint16
}

// DefaultPixelFormat returns a standard 32-bit RGBA pixel format
func DefaultPixelFormat() PixelFormat {
	return PixelFormat{
		BitsPerPixel:   32,
		Depth:          24,
		BigEndianFlag:  0,
		TrueColourFlag: 1,
		RedMax:         255,
		GreenMax:       255,
		BlueMax:        255,
		RedShift:       16,
		GreenShift:     8,
		BlueShift:      0,
		Padding:        [3]uint8{0, 0, 0},
	}
}

// WritePixelFormat writes a pixel format to a byte slice
func (pf *PixelFormat) WriteBytes() []byte {
	buf := make([]byte, 16)
	buf[0] = pf.BitsPerPixel
	buf[1] = pf.Depth
	buf[2] = pf.BigEndianFlag
	buf[3] = pf.TrueColourFlag
	binary.BigEndian.PutUint16(buf[4:6], pf.RedMax)
	binary.BigEndian.PutUint16(buf[6:8], pf.GreenMax)
	binary.BigEndian.PutUint16(buf[8:10], pf.BlueMax)
	buf[10] = pf.RedShift
	buf[11] = pf.GreenShift
	buf[12] = pf.BlueShift
	copy(buf[13:16], pf.Padding[:])
	return buf
}

// ReadPixelFormat reads a pixel format from a byte slice
func ReadPixelFormat(buf []byte) (*PixelFormat, error) {
	if len(buf) < 16 {
		return nil, fmt.Errorf("buffer too short for pixel format")
	}

	pf := &PixelFormat{
		BitsPerPixel:   buf[0],
		Depth:          buf[1],
		BigEndianFlag:  buf[2],
		TrueColourFlag: buf[3],
		RedMax:         binary.BigEndian.Uint16(buf[4:6]),
		GreenMax:       binary.BigEndian.Uint16(buf[6:8]),
		BlueMax:        binary.BigEndian.Uint16(buf[8:10]),
		RedShift:       buf[10],
		GreenShift:     buf[11],
		BlueShift:      buf[12],
	}
	copy(pf.Padding[:], buf[13:16])
	return pf, nil
}
