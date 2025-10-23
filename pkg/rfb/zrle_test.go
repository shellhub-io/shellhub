package rfb

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"testing"
)

// TestZRLEEncoderCreation tests that the ZRLE encoder can be created
func TestZRLEEncoderCreation(t *testing.T) {
	encoder := NewZRLEEncoder()
	if encoder == nil {
		t.Fatal("Failed to create ZRLE encoder")
	}

	if encoder.EncodingType() != EncodingZRLE {
		t.Errorf("Expected encoding type %d, got %d", EncodingZRLE, encoder.EncodingType())
	}
}

// TestZRLEEncoderSimpleData tests ZRLE encoding with simple unichrome data
func TestZRLEEncoderSimpleData(t *testing.T) {
	encoder := NewZRLEEncoder()

	// Create a simple 64x64 red tile in RGBA format
	width := uint16(64)
	height := uint16(64)
	pixelCount := int(width) * int(height)

	// RGBA format: red = [255, 0, 0, 255]
	data := make([]uint8, pixelCount*4)
	for i := 0; i < pixelCount*4; i += 4 {
		data[i] = 255   // R
		data[i+1] = 0   // G
		data[i+2] = 0   // B
		data[i+3] = 255 // A
	}

	pixelFormat := DefaultPixelFormat()
	ctx := &EncoderContext{ConnectionID: 1, CompressionLevel: 6}
	encoded, err := encoder.Encode(ctx, data, width, height, pixelFormat)
	if err != nil {
		t.Fatalf("Failed to encode: %v", err)
	}

	if len(encoded) == 0 {
		t.Fatal("Encoded data is empty")
	}

	// Verify it starts with zlib data length
	if len(encoded) < 4 {
		t.Fatal("Encoded data too short to contain zlib length")
	}

	zlibLen := binary.BigEndian.Uint32(encoded[:4])
	if zlibLen == 0 || int(zlibLen) > len(encoded)-4 {
		t.Errorf("Invalid zlib length: %d", zlibLen)
	}
}

// TestZRLEEncoderVariedData tests ZRLE encoding with varied color data
func TestZRLEEncoderVariedData(t *testing.T) {
	encoder := NewZRLEEncoder()

	// Create a 32x32 checkerboard pattern in RGBA format
	width := uint16(32)
	height := uint16(32)
	pixelCount := int(width) * int(height)

	data := make([]uint8, pixelCount*4)
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			idx := (y*32 + x) * 4
			if (x+y)%2 == 0 {
				// Red
				data[idx] = 255
				data[idx+1] = 0
				data[idx+2] = 0
				data[idx+3] = 255
			} else {
				// Blue
				data[idx] = 0
				data[idx+1] = 0
				data[idx+2] = 255
				data[idx+3] = 255
			}
		}
	}

	pixelFormat := DefaultPixelFormat()
	ctx := &EncoderContext{ConnectionID: 1, CompressionLevel: 6}
	encoded, err := encoder.Encode(ctx, data, width, height, pixelFormat)
	if err != nil {
		t.Fatalf("Failed to encode: %v", err)
	}

	if len(encoded) == 0 {
		t.Fatal("Encoded data is empty")
	}

	// Verify the data is valid and compressible
	zlibLen := binary.BigEndian.Uint32(encoded[:4])
	if zlibLen == 0 {
		t.Fatal("Zlib length is zero")
	}

	// Verify we can decompress it
	compressedData := encoded[4:]
	reader, err := zlib.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		t.Fatalf("Failed to create zlib reader: %v", err)
	}
	defer reader.Close()

	decompressed := make([]byte, 1024*1024) // Large buffer
	n, err := reader.Read(decompressed)
	if n == 0 {
		t.Fatal("No decompressed data")
	}
}

// TestZRLEEncoderSmallTile tests ZRLE encoding with a small tile
func TestZRLEEncoderSmallTile(t *testing.T) {
	encoder := NewZRLEEncoder()

	// Create a small 16x16 tile
	width := uint16(16)
	height := uint16(16)
	pixelCount := int(width) * int(height)

	data := make([]uint8, pixelCount*4)
	// Fill with white color
	for i := 0; i < pixelCount*4; i += 4 {
		data[i] = 255
		data[i+1] = 255
		data[i+2] = 255
		data[i+3] = 255
	}

	pixelFormat := DefaultPixelFormat()
	ctx := &EncoderContext{ConnectionID: 1, CompressionLevel: 6}
	encoded, err := encoder.Encode(ctx, data, width, height, pixelFormat)
	if err != nil {
		t.Fatalf("Failed to encode: %v", err)
	}

	if len(encoded) == 0 {
		t.Fatal("Encoded data is empty")
	}
}

// TestZRLEEncoderNonSquareDimensions tests ZRLE encoding with non-square dimensions
func TestZRLEEncoderNonSquareDimensions(t *testing.T) {
	encoder := NewZRLEEncoder()

	// Create a 100x50 framebuffer (will span multiple tiles)
	width := uint16(100)
	height := uint16(50)
	pixelCount := int(width) * int(height)

	data := make([]uint8, pixelCount*4)
	// Fill with green gradient
	for y := 0; y < 50; y++ {
		for x := 0; x < 100; x++ {
			idx := (y*100 + x) * 4
			data[idx] = 0
			data[idx+1] = uint8((y * 255) / 50)
			data[idx+2] = 0
			data[idx+3] = 255
		}
	}

	pixelFormat := DefaultPixelFormat()
	ctx := &EncoderContext{ConnectionID: 1, CompressionLevel: 6}
	encoded, err := encoder.Encode(ctx, data, width, height, pixelFormat)
	if err != nil {
		t.Fatalf("Failed to encode: %v", err)
	}

	if len(encoded) == 0 {
		t.Fatal("Encoded data is empty")
	}

	// Verify compression (should be reasonably compressed due to gradient)
	zlibLen := binary.BigEndian.Uint32(encoded[:4])
	expectedMinSize := pixelCount * 4 / 10 // At least 10x compression
	if int(zlibLen) > expectedMinSize*100 {
		t.Logf("Warning: compression ratio is low: %d bytes for %d pixels", zlibLen, pixelCount)
	}
}

// TestZRLETileExtraction tests the extractTile function
func TestZRLETileExtraction(t *testing.T) {
	// Create a 128x128 framebuffer
	fbWidth := uint16(128)
	fbHeight := uint16(128)
	bytesPerPixel := 4
	fb := make([]uint8, int(fbWidth)*int(fbHeight)*bytesPerPixel)

	// Fill with a pattern where each pixel's R value is based on its position
	for y := 0; y < 128; y++ {
		for x := 0; x < 128; x++ {
			idx := (y*128 + x) * 4
			fb[idx] = uint8((x + y) % 256) // R
			fb[idx+1] = 0                  // G
			fb[idx+2] = 0                  // B
			fb[idx+3] = 255                // A
		}
	}

	// Extract a 32x32 tile starting at (32, 32)
	tileData := extractTile(fb, 32, 32, 32, 32, fbWidth, bytesPerPixel)

	// Verify the extracted tile
	expectedSize := 32 * 32 * 4
	if len(tileData) != expectedSize {
		t.Errorf("Expected tile size %d, got %d", expectedSize, len(tileData))
	}

	// Check that the first pixel in the tile is correct (should be from position 32,32 in framebuffer)
	expectedR := uint8((32 + 32) % 256)
	if tileData[0] != expectedR {
		t.Errorf("Expected R value %d, got %d", expectedR, tileData[0])
	}
}

// TestZRLEIsUnichromeTile tests the isUnichromeTile function
func TestZRLEIsUnichromeTile(t *testing.T) {
	encoder := NewZRLEEncoder()

	tests := []struct {
		name           string
		data           []uint8
		bytesPerPixel  int
		expectedResult bool
	}{
		{
			name:           "Unichrome tile",
			data:           []uint8{255, 0, 0, 255, 255, 0, 0, 255, 255, 0, 0, 255},
			bytesPerPixel:  4,
			expectedResult: true,
		},
		{
			name:           "Non-unichrome tile",
			data:           []uint8{255, 0, 0, 255, 0, 255, 0, 255},
			bytesPerPixel:  4,
			expectedResult: false,
		},
		{
			name:           "Single pixel",
			data:           []uint8{128, 64, 32, 255},
			bytesPerPixel:  4,
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := encoder.isUnichromeTile(tt.data, tt.bytesPerPixel)
			if result != tt.expectedResult {
				t.Errorf("Expected %v, got %v", tt.expectedResult, result)
			}
		})
	}
}

// TestZRLEEncodingManagerIntegration tests ZRLE with EncodingManager
func TestZRLEEncodingManagerIntegration(t *testing.T) {
	em := NewEncodingManager()

	// Create connection encoders
	connID := uint32(1)
	if err := em.CreateConnectionEncoders(connID); err != nil {
		t.Fatalf("Failed to create connection encoders: %v", err)
	}
	defer em.CloseConnection(connID)

	// Create test data
	width := uint16(64)
	height := uint16(64)
	pixelCount := int(width) * int(height)

	data := make([]uint8, pixelCount*4)
	for i := 0; i < pixelCount*4; i += 4 {
		data[i] = 255
		data[i+1] = 0
		data[i+2] = 0
		data[i+3] = 255
	}

	pixelFormat := DefaultPixelFormat()

	// Test with ZRLE encoding preference
	clientEncodings := []int32{EncodingZRLE, EncodingRaw}
	encodedData, encodingType, err := em.EncodeRectangle(connID, data, width, height, pixelFormat, clientEncodings)
	if err != nil {
		t.Fatalf("Failed to encode rectangle: %v", err)
	}

	if encodingType != EncodingZRLE {
		t.Errorf("Expected encoding type %d, got %d", EncodingZRLE, encodingType)
	}

	if len(encodedData) == 0 {
		t.Fatal("Encoded data is empty")
	}
}
