package web

type MessageKind uint8

const (
	// MessageKindInput is the identifier to a input message. This kind of message can be directly send to [web.Conn].
	MessageKindInput MessageKind = iota + 1
	// MessageKindResize is the identifier to a resize request message. This kind of message contains the number of
	// columns and rows what the terminal should have.
	MessageKindResize
)

type Message struct {
	Kind MessageKind `json:"kind"`
	Data any         `json:"data"`
}
