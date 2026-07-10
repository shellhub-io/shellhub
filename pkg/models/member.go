package models

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
)

// Member status values used by MemberView.Status. They flatten a member's account state into
// a single field the members list renders. Both statuses derive from core-only concepts (the
// users.awaiting_approval flag and the login gate). Cloud/enterprise may extend the view with
// additional statuses (e.g. pending invitations) in its own response type — kept out of core.
const (
	// MemberStatusActive is a completed, confirmed account that is a full member.
	MemberStatusActive = "active"
	// MemberStatusAwaitingApproval is a completed account still waiting for a system admin to
	// approve it; it is already a member but cannot sign in yet (see the auth login gate).
	MemberStatusAwaitingApproval = "awaiting_approval"
	// MemberStatusNotConfirmed is a member whose account was provisioned by the invite but not
	// yet completed; the invitee still has to finish setting it up.
	MemberStatusNotConfirmed = "not-confirmed"
)

// MemberView is the enriched, list-friendly member representation returned by
// GET /api/namespaces/members. Unlike Member it carries the user's name/username and a
// flattened account Status, joining the users table.
type MemberView struct {
	ID       string          `json:"id,omitempty"`
	Name     string          `json:"name,omitempty"`
	Username string          `json:"username,omitempty"`
	Email    string          `json:"email"`
	Role     authorizer.Role `json:"role"`
	// Status is MemberStatusActive or MemberStatusAwaitingApproval.
	Status  string    `json:"status"`
	AddedAt time.Time `json:"added_at,omitempty"`
}

type Member struct {
	ID      string          `json:"id,omitempty"`
	AddedAt time.Time       `json:"added_at"`
	Email   string          `json:"email" validate:"email"`
	Role    authorizer.Role `json:"role" validate:"required,oneof=administrator operator observer"`
	// AccountStatus is the member's underlying user account status (confirmed or
	// not-confirmed). A not-confirmed member still has to finish setting up their account. It
	// is the account status, not the membership-invitation status (accepted/pending), which is
	// a separate concept.
	AccountStatus UserStatus `json:"account_status,omitempty"`
	// AwaitingApproval mirrors the member's user account flag: true while a namespace admin
	// provisioned them but a system admin has not approved the account yet. The account cannot
	// sign in until an admin approves it.
	AwaitingApproval bool `json:"awaiting_approval,omitempty"`
}
