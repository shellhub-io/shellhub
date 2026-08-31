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
	// KindReauthRequired is sent when the identity is allowed but a policy's
	// require_reauth window has lapsed, so a fresh step-up is needed before the
	// session proceeds. On the web path the bridge maps it to a control frame that
	// prompts the browser step-up; access is not denied, just gated.
	KindReauthRequired
)

//go:embed messages/invalid_ssh_id.txt
var rawInvalidSSHID string

//go:embed messages/connection_failed.txt
var rawConnectionFailed string

//go:embed messages/access_denied.txt
var rawAccessDenied string

//go:embed messages/reauth_required.txt
var rawReauthRequired string

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
	normalize(rawReauthRequired):   KindReauthRequired,
}

const codeTrailer = "shellhub-approval: "

func raw(k Kind) string {
	switch k {
	case KindInvalidSSHID:
		return rawInvalidSSHID
	case KindConnectionFailed:
		return rawConnectionFailed
	case KindAccessDenied:
		return rawAccessDenied
	case KindReauthRequired:
		return rawReauthRequired
	default:
		return ""
	}
}

// Message returns the banner text for the given Kind with CRLF line endings.
// It returns an empty string for KindNone.
func Message(k Kind) string {
	return render(raw(k))
}

// MessageWithCode returns the banner for the given Kind with an approval code
// appended. Classify strips the trailer back off, so the message still matches
// its Kind.
func MessageWithCode(k Kind, code string) string {
	message := raw(k)
	if message == "" {
		return ""
	}

	return render(message + "\n" + codeTrailer + code + "\n")
}

// Classify returns the Kind matching the given banner message and the approval
// code it carries, if any. It returns KindNone when the message matches no known
// banner. The comparison is insensitive to line ending style and surrounding
// whitespace.
func Classify(message string) (Kind, string) {
	normalized := normalize(message)

	code := ""
	if at := strings.LastIndex(normalized, codeTrailer); at >= 0 {
		code = strings.TrimSpace(normalized[at+len(codeTrailer):])
		normalized = strings.TrimSpace(normalized[:at])
	}

	if k, ok := index[normalized]; ok {
		return k, code
	}

	return KindNone, ""
}
