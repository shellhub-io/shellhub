package encoders

// EncoderType represents the type of encoder.
type EncoderType byte

// EncoderTypeH264 is the encoder type for H.264.
// TODO: Create a table to map AVCodecID to encoder types.
const EncoderTypeH264 EncoderType = 1

type Encoder interface {
	Encode(uint16, uint16, []byte) ([]byte, error)
	Code() EncoderType
	Close()
}
