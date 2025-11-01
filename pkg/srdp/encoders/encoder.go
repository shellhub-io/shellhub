package encoders

type Encoder interface {
	Encode(uint16, uint16, []byte) ([]byte, error)
	Code() uint8
	Close()
}
