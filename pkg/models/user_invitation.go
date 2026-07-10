package models

import "time"

type UserInvitationStatus string

const (
	UserInvitationStatusPending  UserInvitationStatus = "pending"
	UserInvitationStatusAccepted UserInvitationStatus = "accepted"
)

type UserInvitation struct {
	ID          string               `json:"id"`
	Email       string               `json:"email"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
	Invitations int                  `json:"invitations"`
	Status      UserInvitationStatus `json:"status"`
}
