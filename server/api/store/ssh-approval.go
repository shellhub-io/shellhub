package store

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
)

// SSHApprovalStore persists pending SSH approvals, which are short-lived by design: an
// approval that nobody acts on expires rather than waiting indefinitely.
type SSHApprovalStore interface {
	// SSHApprovalCreate stores a pending SSH login approval.
	SSHApprovalCreate(ctx context.Context, approval *models.SSHApproval) error

	// SSHApprovalGet retrieves an approval by its code. An approval past its
	// expiry is reported as not found: the row may still be there until the cron
	// prunes it, but it can no longer be read or decided.
	SSHApprovalGet(ctx context.Context, code string, now time.Time) (*models.SSHApproval, error)

	// SSHApprovalDecide moves an approval out of pending and reports whether this
	// caller is the one that moved it. The transition itself is the single-use
	// claim, so a double submit cannot write two decisions onto one code. Call it
	// inside a transaction with the decision's durable effect, so the gateway
	// never polls a confirmation whose effect has not landed.
	SSHApprovalDecide(ctx context.Context, code string, state models.SSHApprovalState, userID string, now time.Time) (bool, error)

	// SSHApprovalCleanup deletes approvals that expired before the given time.
	SSHApprovalCleanup(ctx context.Context, before time.Time) (int64, error)
}
