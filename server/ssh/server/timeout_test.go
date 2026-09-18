package server

import (
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/server/api/services"
	"github.com/stretchr/testify/assert"
)

func TestHandshakeBudgetOutlastsAnApproval(t *testing.T) {
	assert.Greater(t, handshakeBudget, services.SSHApprovalTTL,
		"a login whose approval is still valid would be cut off mid-handshake")

	assert.GreaterOrEqual(t, handshakeBudget-services.SSHApprovalTTL, 30*time.Second,
		"the budget must also cover the handshake itself and the person typing the code back")
}
