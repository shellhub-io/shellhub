package web

type messageKind uint8

const (
	// messageKindInput is the identifier to a input message. This kind of message can be directly send to [web.Conn].
	messageKindInput messageKind = iota + 1
	// messageKindResize is the identifier to a resize request message. This kind of message contains the number of
	// columns and rows what the terminal should have.
	messageKindResize
	// messageKindSignature is the identifier to a signature message. This kind of message contains the data to be
	// signed by the user's private key.
	messageKindSignature
	// messageKindError is the identifier to output an erro rmessage. This kind of message contains data to be show
	// in terminal for information propose.
	messageKindError
	// messageKindShare is sent by the client to ask the server to expose this live console session as
	// a public shareable terminal, and sent back by the server carrying the generated share token.
	messageKindShare
)

// ShareRequest is the payload of a client -> server [messageKindShare] message: it asks the server
// to share the current console session. The server replies with a [messageKindShare] message whose
// data is the generated share token (the client builds the public URL from it).
type ShareRequest struct {
	Name     string `json:"name"`
	Writable bool   `json:"writable"`
	// TTL is the lifetime in seconds: 0 = server default, negative = no expiry, positive = custom.
	TTL int `json:"ttl"`
}

// MessageMinSize is the minimum size of a message in bytes. This is used to validate if the message is valid.
const MessageMinSize = 20

// Message is the structure used to send and receive messages through the [web.Conn].
//
// A message min size could match with [MessageMinSize] constant, which is the size of the JSON object without data.
type Message struct {
	Kind messageKind `json:"kind"`
	Data any         `json:"data"`
}
