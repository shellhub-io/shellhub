package srdp

const (
	// Protocol Version
	ProtocolVersion = "SRDP 000.001\n"

	// Security Types
	SecurityTypeNone     = 1
	SecurityTypeSRDPAuth = 2

	// Client to Server Messages
	ClientFramebufferEvent = 0
	ClientKeyEvent         = 1
	ClientPointerEvent     = 2

	// Server to Client Messages
	ServerFrameUpdate = 0

	// Encoders
	EncoderInvalid = 0
	EncoderH264    = 1
)
