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

	// SSHApprovalDecide moves a pending, unexpired approval to the decision and
	// reports whether this caller moved it. The transition is the single-use
	// claim: a double submit, or an approval already decided, expired or missing,
	// returns false and a nil error. A decision with a durable effect of its own,
	// such as a refreshed re-auth window, belongs in one transaction with it, so
	// the gateway never polls a confirmation whose effect has not landed. It
	// returns ctx's error when ctx is cancelled or its deadline passes and an
	// error wrapping [ErrInternal] for any other failure.
	SSHApprovalDecide(ctx context.Context, code string, decision models.SSHApprovalDecision, now time.Time) (bool, error)

	// SSHApprovalRedeem deletes a confirmed, unexpired approval whose
	// confirmation code matches and returns it. The delete is the single-use
	// claim. It returns [ErrNoDocuments] for any other approval, one already
	// redeemed included, ctx's error when ctx is cancelled or its deadline
	// passes, and an error wrapping [ErrInternal] for any other failure.
	SSHApprovalRedeem(ctx context.Context, code, confirmationCode string, now time.Time) (*models.SSHApproval, error)

	// SSHApprovalCleanup deletes approvals that expired before the given time.
	SSHApprovalCleanup(ctx context.Context, before time.Time) (int64, error)
}
