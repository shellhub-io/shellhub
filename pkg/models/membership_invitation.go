package models

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/clock"
)

type MembershipInvitationStatus string

const (
	MembershipInvitationStatusPending   MembershipInvitationStatus = "pending"
	MembershipInvitationStatusAccepted  MembershipInvitationStatus = "accepted"
	MembershipInvitationStatusRejected  MembershipInvitationStatus = "rejected"
	MembershipInvitationStatusCancelled MembershipInvitationStatus = "cancelled"
)

type MembershipInvitation struct {
	ID              string                     `json:"-"`
	TenantID        string                     `json:"-"`
	UserID          string                     `json:"-"`
	InvitedBy       string                     `json:"invited_by"`
	CreatedAt       time.Time                  `json:"created_at"`
	UpdatedAt       time.Time                  `json:"updated_at"`
	ExpiresAt       *time.Time                 `json:"expires_at"`
	Status          MembershipInvitationStatus `json:"status"`
	StatusUpdatedAt time.Time                  `json:"status_updated_at"`
	Role            authorizer.Role            `json:"role"`
	Invitations     int                        `json:"-"`
	// Sig is the one-time signature that ties the invitation link to this row. It
	// replaces the former Redis "invite={sig}" token; validity is the row's ExpiresAt.
	Sig string `json:"-"`

	// NamespaceName isn't saved on the database
	NamespaceName string `json:"-"`
	// UserEmail isn't saved on the database
	UserEmail string `json:"-"`
}

func (m MembershipInvitation) IsExpired() bool {
	return m.ExpiresAt != nil && m.ExpiresAt.Before(clock.Now())
}

func (m MembershipInvitation) IsPending() bool {
	return m.Status == MembershipInvitationStatusPending
}

// MembershipInvitationNotification is the typed, email-relevant snapshot of a membership
// invitation event. It is assembled once by the membership-intake flow and carried — via the
// OnMembershipInvited hook and the internal client — to the worker that renders and sends the
// invitation email, which reads it without a single store round-trip.
//
// It is the single contract across the shellhub↔cloud seam: JSON-encoded over the worker's
// []byte transport, replacing the former positional colon-delimited string. It deliberately
// carries only what the email template consumes — not the role or namespace name, which the
// template uses neither of.
type MembershipInvitationNotification struct {
	// Signature is the invitation's one-time signature; the accept-invite link is keyed by it.
	Signature string `json:"signature"`
	// ExpiresAt is when the invitation stops resolving, shown to the recipient as the link expiry.
	ExpiresAt time.Time `json:"expires_at"`
	// RecipientEmail is the invited address, already lowercased.
	RecipientEmail string `json:"recipient_email"`
	// RecipientName is the invitee's display name, empty for a not-yet-registered invitee (exactly
	// as before).
	RecipientName string `json:"recipient_name"`
	// ForwardedProto and ForwardedHost come from the originating request's X-Forwarded-* headers and
	// build the accept-invite link in the email.
	ForwardedProto string `json:"forwarded_proto"`
	ForwardedHost  string `json:"forwarded_host"`
}
