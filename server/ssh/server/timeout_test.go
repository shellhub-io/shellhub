package server

import (
	"testing"

	"github.com/shellhub-io/shellhub/server/ssh/session"
	"github.com/stretchr/testify/assert"
)

func TestHandshakeBudgetOutlastsApproval(t *testing.T) {
	assert.Greater(t, handshakeBudget, session.ApprovalWaitTimeout,
		"a login waiting on a browser approval would be cut off mid-handshake")
}
