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
