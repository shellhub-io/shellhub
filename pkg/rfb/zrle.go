package rfb

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
)

const (
	// ZRLE tile size - standard 64x64 pixels
	ZRLETileSize = 64

	// ZRLE tile encoding types
	ZRLETileRaw        = 0   // Raw tile data
	ZRLETileUnichrome  = 1   // Single color tile
	ZRLETilePaletteMin = 128 // Palette encoding base type
	ZRLETilePaletteMax = 143 // Palette can be 2-16 colors
)

// ZRLEEncoder implements ZRLE (Zlib Run-Length Encoding) for RFB
type ZRLEEncoder struct {
	rawEncoder     *RawEncoder
	zlibWriter     *zlib.Writer
	compressionBuf bytes.Buffer
	compressionLvl int
}

var _ Encoder = (*ZRLEEncoder)(nil)

// NewZRLEEncoder creates a new ZRLE encoder
func NewZRLEEncoder() *ZRLEEncoder {
	return &ZRLEEncoder{
		rawEncoder:     NewRawEncoder(),
		compressionLvl: 6, // Default zlib compression level
	}
}

// EncodingType returns the encoding type identifier
func (e *ZRLEEncoder) EncodingType() int32 {
	return EncodingZRLE
}

// Reset resets the encoder state for a new connection
func (e *ZRLEEncoder) Reset() error {
	// Close and reset any existing zlib writer
	if e.zlibWriter != nil {
		e.zlibWriter.Close()
		e.zlibWriter = nil
	}
	e.compressionBuf.Reset()
	return nil
}

// SetCompressionLevel sets the zlib compression level (0-9)
func (e *ZRLEEncoder) SetCompressionLevel(level int) error {
	if level < 0 || level > 9 {
		return fmt.Errorf("invalid compression level: %d (must be 0-9)", level)
	}
	e.compressionLvl = level
	return nil
}

// Encode encodes pixel data using ZRLE encoding
func (e *ZRLEEncoder) Encode(ctx *EncoderContext, data []uint8, width, height uint16, pixelFormat PixelFormat) ([]byte, error) {
	// First convert pixel format to match client requirements
	pixelData, err := e.rawEncoder.convertPixelFormat(data, width, height, pixelFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to convert pixel format: %v", err)
	}

	bytesPerPixel := int(pixelFormat.BitsPerPixel / 8)

	// Prepare buffer for encoded tiles
	var encodedBuffer bytes.Buffer

	// Process framebuffer into tiles
	for ty := uint16(0); ty < height; ty += ZRLETileSize {
		for tx := uint16(0); tx < width; tx += ZRLETileSize {
			// Calculate actual tile dimensions (may be smaller at edges)
			tileWidth := ZRLETileSize
			tileHeight := ZRLETileSize

			if int(tx)+tileWidth > int(width) {
				tileWidth = int(width) - int(tx)
			}
			if int(ty)+tileHeight > int(height) {
				tileHeight = int(height) - int(ty)
			}

			// Extract tile data
			tileData := extractTile(pixelData, tx, ty, uint16(tileWidth), uint16(tileHeight),
				width, bytesPerPixel)

			// Encode the tile
			tileEncoded, err := e.encodeTile(tileData, uint16(tileWidth), uint16(tileHeight),
				bytesPerPixel, pixelFormat)
			if err != nil {
				return nil, fmt.Errorf("failed to encode tile at (%d, %d): %v", tx, ty, err)
			}

			if _, err := encodedBuffer.Write(tileEncoded); err != nil {
				return nil, fmt.Errorf("failed to write encoded tile: %v", err)
			}
		}
	}

	// Get uncompressed data length (for the length field in ZRLE header)
	uncompressedLen := encodedBuffer.Len()

	// Compress all tile data with zlib using configured compression level
	var compressedBuffer bytes.Buffer
	zlibWriter, err := zlib.NewWriterLevel(&compressedBuffer, e.compressionLvl)
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib writer: %v", err)
	}
	if _, err := zlibWriter.Write(encodedBuffer.Bytes()); err != nil {
		return nil, fmt.Errorf("failed to write to zlib: %v", err)
	}
	if err := zlibWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close zlib writer: %v", err)
	}

	// Write uncompressed data length followed by compressed data
	// RFC 6143 Section 7.7.6: length field contains the number of uncompressed bytes
	var result bytes.Buffer
	lengthBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBuf, uint32(uncompressedLen))
	result.Write(lengthBuf)
	result.Write(compressedBuffer.Bytes())

	return result.Bytes(), nil
}

// encodeTile encodes a single tile using the best strategy
func (e *ZRLEEncoder) encodeTile(tileData []uint8, width, height uint16,
	bytesPerPixel int, pixelFormat PixelFormat,
) ([]byte, error) {
	pixelCount := int(width) * int(height)

	// Check if all pixels are the same color (unichrome)
	if pixelCount > 0 && e.isUnichromeTile(tileData, bytesPerPixel) {
		return e.encodeUnichroneTile(tileData, bytesPerPixel)
	}

	// Try packed palette encoding for small number of colors
	paletteData, rawData := e.tryPaletteEncoding(tileData, width, height, bytesPerPixel)
	if paletteData != nil {
		return paletteData, nil
	}
	if rawData != nil {
		return rawData, nil
	}

	// Fall back to raw tile encoding
	return e.encodeRawTile(tileData, bytesPerPixel)
}

// isUnichromeTile checks if all pixels in the tile are the same color
func (e *ZRLEEncoder) isUnichromeTile(tileData []uint8, bytesPerPixel int) bool {
	if len(tileData) < bytesPerPixel {
		return false
	}

	firstPixel := tileData[:bytesPerPixel]
	for i := bytesPerPixel; i+bytesPerPixel <= len(tileData); i += bytesPerPixel {
		if !bytes.Equal(tileData[i:i+bytesPerPixel], firstPixel) {
			return false
		}
	}

	return true
}

// encodeUnichroneTile encodes a tile with a single color
func (e *ZRLEEncoder) encodeUnichroneTile(tileData []uint8, bytesPerPixel int) ([]byte, error) {
	// Type 1: unichrome tile
	// Format: [1 byte type=1][pixel data]
	result := make([]byte, 1+bytesPerPixel)
	result[0] = ZRLETileUnichrome
	copy(result[1:], tileData[:bytesPerPixel])

	return result, nil
}

// tryPaletteEncoding tries to encode the tile using a palette
// Returns (encodedData, nil) if palette encoding is effective, (nil, encodedData) if raw is better
func (e *ZRLEEncoder) tryPaletteEncoding(tileData []uint8, width, height uint16,
	bytesPerPixel int,
) ([]byte, []byte) {
	// Count unique colors
	colorSet := make(map[string]uint32)
	for i := 0; i+bytesPerPixel <= len(tileData); i += bytesPerPixel {
		pixel := tileData[i : i+bytesPerPixel]
		pixelKey := string(pixel)
		colorSet[pixelKey]++
	}

	numColors := len(colorSet)

	// Palette encoding is only useful for 2-16 colors
	if numColors < 2 || numColors > 16 {
		return nil, nil
	}

	// Build palette
	palette := make([][]byte, 0, numColors)
	colorToIndex := make(map[string]byte)

	for pixelKey := range colorSet {
		colorToIndex[pixelKey] = byte(len(palette))
		palette = append(palette, []byte(pixelKey))
	}

	// Encode with palette
	var encodedBuffer bytes.Buffer

	// Type: 128 + (numColors - 1)
	paletteType := byte(ZRLETilePaletteMin + byte(numColors-1))
	encodedBuffer.WriteByte(paletteType)

	// Write palette colors
	for _, color := range palette {
		encodedBuffer.Write(color)
	}

	// Encode pixels using palette indices and run-length encoding
	pixelIndex := 0
	for pixelIndex < len(tileData) {
		pixelKey := string(tileData[pixelIndex : pixelIndex+bytesPerPixel])
		colorIndex := colorToIndex[pixelKey]

		// Count run length
		runLength := 1
		nextIndex := pixelIndex + bytesPerPixel
		for nextIndex+bytesPerPixel <= len(tileData) && runLength < 255 {
			nextPixelKey := string(tileData[nextIndex : nextIndex+bytesPerPixel])
			if nextPixelKey != pixelKey {
				break
			}
			runLength++
			nextIndex += bytesPerPixel
		}

		// Write color index (with run-length encoding if run > 1)
		if runLength == 1 {
			encodedBuffer.WriteByte(colorIndex)
		} else {
			encodedBuffer.WriteByte(colorIndex | 0x80) // Set high bit for run-length
			encodedBuffer.WriteByte(byte(runLength - 1))
		}

		pixelIndex = nextIndex
	}

	paletteEncoded := encodedBuffer.Bytes()

	// Compare with raw encoding - use palette if it's smaller
	rawEncoded, _ := e.encodeRawTile(tileData, bytesPerPixel)

	if len(paletteEncoded) < len(rawEncoded) {
		return paletteEncoded, nil
	}

	return nil, rawEncoded
}

// encodeRawTile encodes a tile with raw pixel data
func (e *ZRLEEncoder) encodeRawTile(tileData []uint8, bytesPerPixel int) ([]byte, error) {
	// Type 0: raw tile
	// Format: [1 byte type=0][raw pixel data]
	result := make([]byte, 1+len(tileData))
	result[0] = ZRLETileRaw
	copy(result[1:], tileData)

	return result, nil
}

// extractTile extracts a tile region from the full framebuffer
func extractTile(framebuffer []uint8, startX, startY, tileWidth, tileHeight, fbWidth uint16,
	bytesPerPixel int,
) []uint8 {
	tileData := make([]uint8, int(tileWidth)*int(tileHeight)*bytesPerPixel)
	dstIdx := 0

	for y := startY; y < startY+tileHeight; y++ {
		fbRowStart := int(y)*int(fbWidth)*bytesPerPixel + int(startX)*bytesPerPixel
		fbRowEnd := fbRowStart + int(tileWidth)*bytesPerPixel

		// Handle partial rows at framebuffer boundaries
		if fbRowStart < len(framebuffer) {
			// Ensure we don't read past the end of the framebuffer
			actualEnd := fbRowEnd
			if actualEnd > len(framebuffer) {
				actualEnd = len(framebuffer)
			}
			// Copy only the available bytes
			bytesToCopy := actualEnd - fbRowStart
			copy(tileData[dstIdx:], framebuffer[fbRowStart:actualEnd])
			dstIdx += bytesToCopy
			// Pad the rest of the row with zeros if we hit the boundary
			dstIdx += int(tileWidth)*bytesPerPixel - bytesToCopy
		} else {
			// Entire row is out of bounds - fill with zeros (black)
			dstIdx += int(tileWidth) * bytesPerPixel
		}
	}

	return tileData
}
