package srdp

const (
	// Protocol Version: SRDP 000.001
	ProtocolVersion = "SRDP 000.001\n"

	// Security Types - negotiated during handshake
	SecurityTypeNone     = 1 // No authentication
	SecurityTypeSRDPAuth = 2 // SRDP authentication (not yet implemented)

	// Client to Server Messages - sent during message loop
	ClientFramebufferEvent = 0 // Framebuffer event (reserved)
	ClientKeyEvent         = 1 // Keyboard event: [down_flag(1), padding(2), key(4)]
	ClientPointerEvent     = 2 // Pointer event: [button_mask(1), x(2), y(2)]

	// Server to Client Messages - sent during streaming
	ServerFrameUpdate = 0 // H.264 encoded video frame: [length(8), data(N)]

	// Encoders - negotiated during ClientInit/ServerInit
	EncoderInvalid = 0 // Invalid encoder (error state)
	EncoderH264    = 1 // H.264 video codec (only supported encoder)

	// Client Options - sent during ClientInit phase
	// Format: [option_type(1), option_value...]
	ClientOptionWindowSize = 1 // Request display dimensions: [width(2), height(2)]
	ClientOptionFPS        = 2 // Request frames per second: [fps(2)]
	ClientOptionEncoder    = 3 // Request encoder preference: [num_encoders(1), encoder_ids...]
)
