package models

import "time"

// UserInvitationStatus is where an invitation to create an account sits. Unlike a membership
// invitation it cannot be rejected: it is either open or used.
type UserInvitationStatus string

const (
	// UserInvitationStatusPending is an invitation whose account has not been created yet.
	UserInvitationStatusPending UserInvitationStatus = "pending"
	// UserInvitationStatusAccepted is an invitation that has been used to create an account, and
	// cannot be used again.
	UserInvitationStatusAccepted UserInvitationStatus = "accepted"
)

// UserInvitation is an offer to create an account on the instance, addressed to an email that has
// none. It is distinct from MembershipInvitation, which offers an existing account a place in a
// namespace.
type UserInvitation struct {
	ID          string               `json:"id"`
	Email       string               `json:"email"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
	Invitations int                  `json:"invitations"`
	Status      UserInvitationStatus `json:"status"`
}
