package rfb

import (
	"encoding/binary"
	"fmt"
)

// RawEncoder implements raw pixel encoding (no compression)
type RawEncoder struct{}

// NewRawEncoder creates a new raw encoder
func NewRawEncoder() *RawEncoder {
	return &RawEncoder{}
}

var _ Encoder = (*RawEncoder)(nil)

// EncodingType returns the encoding type identifier
func (e *RawEncoder) EncodingType() int32 {
	return EncodingRaw
}

// Reset resets the encoder state (no-op for raw encoder)
func (e *RawEncoder) Reset() error {
	return nil
}

// SetCompressionLevel sets compression level (no-op for raw encoder)
func (e *RawEncoder) SetCompressionLevel(level int) error {
	return nil
}

// Encode encodes pixel data using raw encoding
func (e *RawEncoder) Encode(ctx *EncoderContext, data []uint8, width, height uint16, pixelFormat PixelFormat) ([]byte, error) {
	// Raw encoding just returns the pixel data as-is
	// Convert from RGBA to the client's pixel format if needed
	return e.convertPixelFormat(data, width, height, pixelFormat)
}

// convertPixelFormat converts pixel data to match the client's pixel format
func (e *RawEncoder) convertPixelFormat(data []uint8, width, height uint16, pixelFormat PixelFormat) ([]byte, error) {
	bytesPerPixel := pixelFormat.BitsPerPixel / 8
	pixelCount := int(width) * int(height)
	output := make([]byte, pixelCount*int(bytesPerPixel))

	for i := range pixelCount {
		srcOffset := i * 4 // RGBA input
		dstOffset := i * int(bytesPerPixel)

		if srcOffset+3 >= len(data) {
			break
		}

		// Extract RGBA values
		r := uint32(data[srcOffset])
		g := uint32(data[srcOffset+1])
		b := uint32(data[srcOffset+2])
		// a := uint32(data[srcOffset+3]) // Alpha not used in most RFB implementations

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
