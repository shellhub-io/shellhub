package encoders

import (
	"errors"

	log "github.com/sirupsen/logrus"
)

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

var ErrNoEncoderAvailable = errors.New("no H.264 encoder available")

// NewH264 creates a new H.264 encoder, trying NVENC, VAAPI, and libx264 in that order.
func NewH264(width, height int, fps uint16) (Encoder, error) {
	if e, err := NewH264NVENCEncoder(width, height, fps); err == nil {
		log.Info("Using NVIDIA NVENC to H.264 encoder")

		return e, nil
	}

	if e, err := NewH264VAAPIEncoder(width, height, fps); err == nil {
		log.Info("Using VAAPI to H.264 encoder")

		return e, nil
	}

	if e, err := NewH264LibX264Encoder(width, height, fps); err == nil {
		log.Warn("Using libx264 to H.264 encoder, performance may be suboptimal")

		return e, nil
	}

	return nil, ErrNoEncoderAvailable
}
