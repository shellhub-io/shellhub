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
	ID              string                     `json:"-" bson:"_id"`
	TenantID        string                     `json:"-" bson:"tenant_id"`
	UserID          string                     `json:"-" bson:"user_id"`
	InvitedBy       string                     `json:"invited_by" bson:"invited_by"`
	CreatedAt       time.Time                  `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time                  `json:"updated_at" bson:"updated_at"`
	ExpiresAt       *time.Time                 `json:"expires_at" bson:"expires_at"`
	Status          MembershipInvitationStatus `json:"status" bson:"status"`
	StatusUpdatedAt time.Time                  `json:"status_updated_at" bson:"status_updated_at"`
	Role            authorizer.Role            `json:"role" bson:"role"`
	Invitations     int                        `json:"-" bson:"invitations"`

	// NamespaceName isn't saved on the database
	NamespaceName string `json:"-" bson:"namespace_name,omitempty"`
	// UserEmail isn't saved on the database
	UserEmail string `json:"-" bson:"user_email,omitempty"`
}

func (m MembershipInvitation) IsExpired() bool {
	return m.ExpiresAt != nil && m.ExpiresAt.Before(clock.Now())
}

func (m MembershipInvitation) IsPending() bool {
	return m.Status == MembershipInvitationStatusPending
}
