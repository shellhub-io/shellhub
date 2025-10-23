package rfb

import (
	"fmt"
)

// EncoderContext provides connection-specific context to encoders
type EncoderContext struct {
	ConnectionID      uint32
	CompressionLevel  int // 0-9 for zlib, 0=none, 9=max
	ClientPixelFormat *PixelFormat
}

// Encoder interface for different RFB encodings
type Encoder interface {
	// Encode encodes pixel data using the specific encoding
	Encode(ctx *EncoderContext, data []uint8, width, height uint16, pixelFormat PixelFormat) ([]byte, error)

	// EncodingType returns the encoding type identifier
	EncodingType() int32

	// Reset resets the encoder state (e.g., zlib streams for ZRLE)
	// Should be called at the start of a new client connection
	Reset() error

	// SetCompressionLevel sets the compression level (if applicable)
	// For zlib-based encodings, level should be 0-9 (0=none, 9=max)
	SetCompressionLevel(level int) error
}

// EncodingManager manages different encoders
type EncodingManager struct {
	encoders     map[int32]Encoder
	connections  map[uint32]map[int32]Encoder // Per-connection encoder instances
	defaultLevel int
}

// NewEncodingManager creates a new encoding manager
func NewEncodingManager() *EncodingManager {
	em := &EncodingManager{
		encoders:     make(map[int32]Encoder),
		connections:  make(map[uint32]map[int32]Encoder),
		defaultLevel: 6, // Default zlib compression level (0-9)
	}

	// Register available encoders
	em.RegisterEncoder(NewRawEncoder())
	em.RegisterEncoder(NewZRLEEncoder())

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

// CreateConnectionEncoders creates per-connection encoder instances for a new client
func (em *EncodingManager) CreateConnectionEncoders(connID uint32) error {
	connEncoders := make(map[int32]Encoder)

	// Create fresh instances of each encoder for this connection
	for encType, baseEncoder := range em.encoders {
		var encoder Encoder
		switch baseEncoder.EncodingType() {
		case EncodingRaw:
			encoder = NewRawEncoder()
		case EncodingZRLE:
			encoder = NewZRLEEncoder()
		default:
			encoder = baseEncoder
		}

		// Set default compression level
		if err := encoder.SetCompressionLevel(em.defaultLevel); err != nil {
			return fmt.Errorf("failed to set compression level: %v", err)
		}

		connEncoders[encType] = encoder
	}

	em.connections[connID] = connEncoders

	return nil
}

// GetConnectionEncoder returns the encoder for a specific connection and encoding type
func (em *EncodingManager) GetConnectionEncoder(connID uint32, encodingType int32) Encoder {
	if connEncoders, exists := em.connections[connID]; exists {
		if encoder, exists := connEncoders[encodingType]; exists {
			return encoder
		}

		// Fall back to raw for this connection
		if encoder, exists := connEncoders[EncodingRaw]; exists {
			return encoder
		}
	}
	// Fall back to default encoder
	return em.GetEncoder(encodingType)
}

// ResetConnection resets all encoders for a connection
func (em *EncodingManager) ResetConnection(connID uint32) error {
	if connEncoders, exists := em.connections[connID]; exists {
		for _, encoder := range connEncoders {
			if err := encoder.Reset(); err != nil {
				return fmt.Errorf("failed to reset encoder: %v", err)
			}
		}
	}

	return nil
}

// CloseConnection closes and cleans up all encoders for a connection
func (em *EncodingManager) CloseConnection(connID uint32) {
	delete(em.connections, connID)
}

// EncodeRectangle encodes a rectangle using the best available encoding
func (em *EncodingManager) EncodeRectangle(connID uint32, data []uint8, width, height uint16,
	pixelFormat PixelFormat, clientEncodings []int32,
) ([]byte, int32, error) {
	// Get the best encoder for client preferences
	baseEncoder := em.GetBestEncoder(clientEncodings)
	encoder := em.GetConnectionEncoder(connID, baseEncoder.EncodingType())

	// Create context for the encoder
	ctx := &EncoderContext{
		ConnectionID:      connID,
		CompressionLevel:  em.defaultLevel,
		ClientPixelFormat: &pixelFormat,
	}

	encodedData, err := encoder.Encode(ctx, data, width, height, pixelFormat)
	if err != nil {
		return nil, 0, err
	}

	return encodedData, encoder.EncodingType(), nil
}
