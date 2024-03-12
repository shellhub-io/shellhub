package web

type messageKind uint8

const (
	// messageKindInput is the identifier to a input message. This kind of message can be directly send to [web.Conn].
	messageKindInput messageKind = iota + 1
	// messageKindResize is the identifier to a resize request message. This kind of message contains the number of
	// columns and rows what the terminal should have.
	messageKindResize
)

type Message struct {
	Kind messageKind `json:"kind"`
	Data any         `json:"data"`
}
