package web

type messageKind uint8

const (
	messageKindInput messageKind = iota + 1
	messageKindResize
	messageKindSignature
	messageKindError
	messageKindSession
	messageKindReauth
	// messageKindReauthDone carries the browser's answer to an approval: the
	// confirmation code the console showed, or empty when the person dismissed
	// the dialog. It is the only inbound kind the bridge waits on mid-handshake.
	messageKindReauthDone
)

// MessageMinSize is the minimum size of a message in bytes. This is used to validate if the message is valid.
const MessageMinSize = 20

// Message is the structure used to send and receive messages through the [web.Conn].
//
// A message min size could match with [MessageMinSize] constant, which is the size of the JSON object without data.
type Message struct {
	Kind messageKind `json:"kind"`
	Data any         `json:"data"`
}
