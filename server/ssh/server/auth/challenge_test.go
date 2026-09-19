package auth

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/ssh/session"
	"github.com/stretchr/testify/assert"
)

// TestWebChallengeCarriesTheCodeInTheQuestion pins the half of the
// server-to-bridge contract that lives here. The web terminal's bridge reads the
// challenge's question, so a code placed anywhere else reaches it as an empty
// string and opens a dialog on nothing.
//
// Its counterpart is TestApprovalChallengeCarriesTheCodeToTheBrowser, in the web
// package. Neither catches a mismatch alone: one asserts what is sent, the other
// what is read.
func TestWebChallengeCarriesTheCodeInTheQuestion(t *testing.T) {
	web := &session.Session{Data: session.Data{Web: true, ApprovalCode: "WXYZ2K7Q"}} //nolint:exhaustruct // only the fields the challenge reads

	assert.Equal(t, "WXYZ2K7Q", question(web), "the bridge reads the question, so the code travels there")
	assert.Empty(t, instruction(web, models.SSHApprovalReauth), "a program has no use for prose written for a person")
}

// TestTerminalChallengeAsksAPersonForTheCode covers the other side of the same
// branch: a terminal gets a label to type under and the text explaining where
// the code comes from.
func TestTerminalChallengeAsksAPersonForTheCode(t *testing.T) {
	terminal := &session.Session{Data: session.Data{Web: false, ApprovalCode: "WXYZ2K7Q"}} //nolint:exhaustruct // only the fields the challenge reads

	assert.Equal(t, "Confirmation code: ", question(terminal))
	assert.Contains(t, instruction(terminal, models.SSHApprovalIdentity), "WXYZ2K7Q",
		"the person needs the code to find the right screen")
}
