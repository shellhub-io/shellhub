package banner

import (
	_ "embed"
	"strings"
)

// Kind identifies which banner message to use.
type Kind int

const (
	// KindNone means no banner should be sent.
	KindNone Kind = iota
	// KindInvalidSSHID is sent when the SSH ID format is incorrect.
	KindInvalidSSHID
	// KindConnectionFailed is sent when the target device cannot be reached.
	KindConnectionFailed
	// KindAccessDenied is sent when access to the device is denied.
	KindAccessDenied
)

//go:embed messages/invalid_ssh_id.txt
var rawInvalidSSHID string

//go:embed messages/connection_failed.txt
var rawConnectionFailed string

//go:embed messages/access_denied.txt
var rawAccessDenied string

// render converts LF line endings to the CRLF required by the SSH protocol.
func render(s string) string {
	// Normalise first so we never double-add \r.
	s = strings.ReplaceAll(s, "\r\n", "\n")

	return strings.ReplaceAll(s, "\n", "\r\n")
}

// normalize strips \r and trims surrounding whitespace so that messages with
// different line endings or trailing newlines compare equal.
func normalize(s string) string {
	s = strings.ReplaceAll(s, "\r", "")

	return strings.TrimSpace(s)
}

// index maps the normalised form of each embedded message to its Kind.
var index = map[string]Kind{
	normalize(rawInvalidSSHID):     KindInvalidSSHID,
	normalize(rawConnectionFailed): KindConnectionFailed,
	normalize(rawAccessDenied):     KindAccessDenied,
}

// Message returns the banner text for the given Kind with CRLF line endings.
// It returns an empty string for KindNone.
func Message(k Kind) string {
	switch k {
	case KindInvalidSSHID:
		return render(rawInvalidSSHID)
	case KindConnectionFailed:
		return render(rawConnectionFailed)
	case KindAccessDenied:
		return render(rawAccessDenied)
	default:
		return ""
	}
}

// Classify returns the Kind that matches the given banner message, or KindNone
// if the message does not match any known banner. The comparison is insensitive
// to line ending style and surrounding whitespace.
func Classify(message string) Kind {
	if k, ok := index[normalize(message)]; ok {
		return k
	}

	return KindNone
}
