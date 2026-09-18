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

func render(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")

	return strings.ReplaceAll(s, "\n", "\r\n")
}

func normalize(s string) string {
	s = strings.ReplaceAll(s, "\r", "")

	return strings.TrimSpace(s)
}

var index = map[string]Kind{
	normalize(rawInvalidSSHID):     KindInvalidSSHID,
	normalize(rawConnectionFailed): KindConnectionFailed,
	normalize(rawAccessDenied):     KindAccessDenied,
}

func raw(k Kind) string {
	switch k {
	case KindInvalidSSHID:
		return rawInvalidSSHID
	case KindConnectionFailed:
		return rawConnectionFailed
	case KindAccessDenied:
		return rawAccessDenied
	default:
		return ""
	}
}

// Message returns the banner text for the given Kind with CRLF line endings.
// It returns an empty string for KindNone.
func Message(k Kind) string {
	return render(raw(k))
}

// Classify returns the Kind matching the given banner message. It returns
// KindNone when the message matches no known banner. The comparison is
// insensitive to line ending style and surrounding whitespace.
func Classify(message string) Kind {
	if k, ok := index[normalize(message)]; ok {
		return k
	}

	return KindNone
}
