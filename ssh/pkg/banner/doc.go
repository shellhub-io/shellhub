// Package banner provides SSH banner messages for well-known error conditions.
//
// It defines a Kind enum for each message type, renders messages with CRLF
// line endings (required by the SSH protocol), and classifies an incoming
// banner string back to its Kind.
package banner
