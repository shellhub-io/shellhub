package models

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
)

type Member struct {
	ID      string          `json:"id,omitempty"`
	AddedAt time.Time       `json:"added_at"`
	Email   string          `json:"email" validate:"email"`
	Role    authorizer.Role `json:"role" validate:"required,oneof=administrator operator observer"`
	// AccountStatus is the member's underlying user account status (confirmed or
	// not-confirmed). A not-confirmed member was provisioned inline and still has to
	// complete their account through an activation link. It is the account status, not the
	// cloud membership-invitation status (accepted/pending), which is a separate concept.
	AccountStatus UserStatus `json:"account_status,omitempty"`
	// AwaitingApproval mirrors the member's user account flag: true while a namespace admin
	// provisioned them but a system admin has not approved the account yet. The activation
	// link cannot be minted for them until an admin approves.
	AwaitingApproval bool `json:"awaiting_approval,omitempty"`
}
