package banner

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageKindNone(t *testing.T) {
	assert.Empty(t, Message(KindNone))
}

func TestMessageHasCRLF(t *testing.T) {
	cases := []struct {
		description string
		kind        Kind
	}{
		{
			description: "KindInvalidSSHID message uses CRLF line endings",
			kind:        KindInvalidSSHID,
		},
		{
			description: "KindConnectionFailed message uses CRLF line endings",
			kind:        KindConnectionFailed,
		},
		{
			description: "KindAccessDenied message uses CRLF line endings",
			kind:        KindAccessDenied,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			msg := Message(tc.kind)
			assert.NotEmpty(t, msg)
			assert.Contains(t, msg, "\r\n", "message must contain CRLF line endings")
		})
	}
}

func TestClassifyRoundTrip(t *testing.T) {
	cases := []struct {
		description string
		kind        Kind
	}{
		{
			description: "round-trip for KindInvalidSSHID",
			kind:        KindInvalidSSHID,
		},
		{
			description: "round-trip for KindConnectionFailed",
			kind:        KindConnectionFailed,
		},
		{
			description: "round-trip for KindAccessDenied",
			kind:        KindAccessDenied,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			kind, code := Classify(Message(tc.kind))
			assert.Equal(t, tc.kind, kind)
			assert.Empty(t, code)
		})
	}
}

func TestClassifyLFOnlyInput(t *testing.T) {
	cases := []struct {
		description string
		kind        Kind
	}{
		{
			description: "LF-only input classifies KindInvalidSSHID",
			kind:        KindInvalidSSHID,
		},
		{
			description: "LF-only input classifies KindConnectionFailed",
			kind:        KindConnectionFailed,
		},
		{
			description: "LF-only input classifies KindAccessDenied",
			kind:        KindAccessDenied,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			// Replace all CRLF with LF to simulate LF-only input.
			lfOnly := strings.ReplaceAll(Message(tc.kind), "\r\n", "\n")
			kind, _ := Classify(lfOnly)
			assert.Equal(t, tc.kind, kind)
		})
	}
}

func TestClassifyStrippedTrailingNewline(t *testing.T) {
	cases := []struct {
		description string
		kind        Kind
	}{
		{
			description: "trailing-newline-stripped input classifies KindInvalidSSHID",
			kind:        KindInvalidSSHID,
		},
		{
			description: "trailing-newline-stripped input classifies KindConnectionFailed",
			kind:        KindConnectionFailed,
		},
		{
			description: "trailing-newline-stripped input classifies KindAccessDenied",
			kind:        KindAccessDenied,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			stripped := strings.TrimRight(Message(tc.kind), "\r\n")
			kind, _ := Classify(stripped)
			assert.Equal(t, tc.kind, kind)
		})
	}
}

func TestClassifyUnknownInputsReturnKindNone(t *testing.T) {
	cases := []struct {
		description string
		input       string
	}{
		{
			description: "empty string returns KindNone",
			input:       "",
		},
		{
			description: "unrelated string returns KindNone",
			input:       "some unrelated message",
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			kind, _ := Classify(tc.input)
			assert.Equal(t, KindNone, kind)
		})
	}
}

// A banner carrying an approval code still classifies as its Kind, and hands the
// code back: it is how the web bridge learns which held login to open a screen
// for.
func TestClassifyRecoversTheApprovalCode(t *testing.T) {
	kind, code := Classify(MessageWithCode(KindReauthRequired, "WXYZ2K7Q"))

	assert.Equal(t, KindReauthRequired, kind)
	assert.Equal(t, "WXYZ2K7Q", code)
}

func TestMessageWithCodeIsEmptyForKindNone(t *testing.T) {
	assert.Empty(t, MessageWithCode(KindNone, "WXYZ2K7Q"))
}
