package vnc

import (
	"encoding/binary"
	"fmt"
)

// Encoder interface for different VNC encodings
type Encoder interface {
	Encode(data []uint8, width, height uint16, pixelFormat PixelFormat) ([]byte, error)
	EncodingType() int32
}

// RawEncoder implements raw pixel encoding (no compression)
type RawEncoder struct{}

// NewRawEncoder creates a new raw encoder
func NewRawEncoder() *RawEncoder {
	return &RawEncoder{}
}

// EncodingType returns the encoding type identifier
func (e *RawEncoder) EncodingType() int32 {
	return EncodingRaw
}

// Encode encodes pixel data using raw encoding
func (e *RawEncoder) Encode(data []uint8, width, height uint16, pixelFormat PixelFormat) ([]byte, error) {
	// Raw encoding just returns the pixel data as-is
	// Convert from RGBA to the client's pixel format if needed
	return e.convertPixelFormat(data, width, height, pixelFormat)
}

// convertPixelFormat converts pixel data to match the client's pixel format
func (e *RawEncoder) convertPixelFormat(data []uint8, width, height uint16, pixelFormat PixelFormat) ([]byte, error) {
	bytesPerPixel := pixelFormat.BitsPerPixel / 8
	pixelCount := int(width) * int(height)
	output := make([]byte, pixelCount*int(bytesPerPixel))

	for i := 0; i < pixelCount; i++ {
		srcOffset := i * 4 // RGBA input
		dstOffset := i * int(bytesPerPixel)

		if srcOffset+3 >= len(data) {
			break
		}

		// Extract RGBA values
		r := uint32(data[srcOffset])
		g := uint32(data[srcOffset+1])
		b := uint32(data[srcOffset+2])
		// a := uint32(data[srcOffset+3]) // Alpha not used in most VNC implementations

		// Convert to client pixel format
		var pixel uint32
		if pixelFormat.TrueColourFlag != 0 {
			// True color
			rScaled := (r * uint32(pixelFormat.RedMax)) / 255
			gScaled := (g * uint32(pixelFormat.GreenMax)) / 255
			bScaled := (b * uint32(pixelFormat.BlueMax)) / 255

			pixel = (rScaled << pixelFormat.RedShift) |
				(gScaled << pixelFormat.GreenShift) |
				(bScaled << pixelFormat.BlueShift)
		} else {
			// Color map mode - simplified, just use grayscale
			gray := (r + g + b) / 3
			pixel = gray
		}

		// Write pixel in correct byte order
		switch bytesPerPixel {
		case 1:
			output[dstOffset] = uint8(pixel)
		case 2:
			if pixelFormat.BigEndianFlag != 0 {
				binary.BigEndian.PutUint16(output[dstOffset:dstOffset+2], uint16(pixel))
			} else {
				binary.LittleEndian.PutUint16(output[dstOffset:dstOffset+2], uint16(pixel))
			}
		case 4:
			if pixelFormat.BigEndianFlag != 0 {
				binary.BigEndian.PutUint32(output[dstOffset:dstOffset+4], pixel)
			} else {
				binary.LittleEndian.PutUint32(output[dstOffset:dstOffset+4], pixel)
			}
		default:
			return nil, fmt.Errorf("unsupported bytes per pixel: %d", bytesPerPixel)
		}
	}

	return output, nil
}

// RREEncoder implements Run-Length Encoding
type RREEncoder struct{}

// NewRREEncoder creates a new RRE encoder
func NewRREEncoder() *RREEncoder {
	return &RREEncoder{}
}

// EncodingType returns the encoding type identifier
func (e *RREEncoder) EncodingType() int32 {
	return EncodingRRE
}

// Encode encodes pixel data using RRE encoding
func (e *RREEncoder) Encode(data []uint8, width, height uint16, pixelFormat PixelFormat) ([]byte, error) {
	// For simplicity, fall back to raw encoding
	// A full RRE implementation would find rectangular regions of the same color
	rawEncoder := NewRawEncoder()
	return rawEncoder.Encode(data, width, height, pixelFormat)
}

// HextileEncoder implements Hextile encoding
type HextileEncoder struct{}

// NewHextileEncoder creates a new Hextile encoder
func NewHextileEncoder() *HextileEncoder {
	return &HextileEncoder{}
}

// EncodingType returns the encoding type identifier
func (e *HextileEncoder) EncodingType() int32 {
	return EncodingHextile
}

// Encode encodes pixel data using Hextile encoding
func (e *HextileEncoder) Encode(data []uint8, width, height uint16, pixelFormat PixelFormat) ([]byte, error) {
	// For simplicity, fall back to raw encoding
	// A full Hextile implementation would divide the image into 16x16 tiles
	rawEncoder := NewRawEncoder()
	return rawEncoder.Encode(data, width, height, pixelFormat)
}

// EncodingManager manages different encoders
type EncodingManager struct {
	encoders map[int32]Encoder
}

// NewEncodingManager creates a new encoding manager
func NewEncodingManager() *EncodingManager {
	em := &EncodingManager{
		encoders: make(map[int32]Encoder),
	}

	// Register available encoders
	em.RegisterEncoder(NewRawEncoder())
	// TODO: Implement proper RRE and Hextile encoders
	// em.RegisterEncoder(NewRREEncoder())
	// em.RegisterEncoder(NewHextileEncoder())

	return em
}

// RegisterEncoder registers an encoder
func (em *EncodingManager) RegisterEncoder(encoder Encoder) {
	em.encoders[encoder.EncodingType()] = encoder
}

// GetEncoder returns an encoder for the given encoding type
func (em *EncodingManager) GetEncoder(encodingType int32) Encoder {
	if encoder, exists := em.encoders[encodingType]; exists {
		return encoder
	}
	// Fall back to raw encoding
	return em.encoders[EncodingRaw]
}

// GetBestEncoder returns the best encoder for the given client preferences
func (em *EncodingManager) GetBestEncoder(clientEncodings []int32) Encoder {
	// Return the first encoding that we support, in client preference order
	for _, encoding := range clientEncodings {
		if encoder, exists := em.encoders[encoding]; exists {
			return encoder
		}
	}
	
	// Default to raw encoding
	return em.encoders[EncodingRaw]
}

// EncodeRectangle encodes a rectangle using the best available encoding
func (em *EncodingManager) EncodeRectangle(data []uint8, width, height uint16, 
	pixelFormat PixelFormat, clientEncodings []int32) ([]byte, int32, error) {
	
	encoder := em.GetBestEncoder(clientEncodings)
	encodedData, err := encoder.Encode(data, width, height, pixelFormat)
	if err != nil {
		return nil, 0, err
	}
	
	return encodedData, encoder.EncodingType(), nil
}