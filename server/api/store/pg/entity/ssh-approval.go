package entity

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

type SSHApproval struct {
	bun.BaseModel `bun:"table:ssh_approvals"`

	Code         string    `bun:"code,pk"`
	NamespaceID  string    `bun:"namespace_id,type:uuid"`
	Kind         string    `bun:"kind"`
	SessionUID   string    `bun:"session_uid"`
	SSHID        string    `bun:"sshid"`
	DeviceUID    string    `bun:"device_uid"`
	DeviceName   string    `bun:"device_name"`
	Username     string    `bun:"username"`
	IPAddress    string    `bun:"ip_address"`
	Fingerprint  string    `bun:"fingerprint"`
	Data         []byte    `bun:"data,type:bytea"`
	ReauthPeriod *int      `bun:"reauth_period"`
	State        string    `bun:"state"`
	DecidedBy    *string   `bun:"decided_by,type:uuid"`
	RequestedAt  time.Time `bun:"requested_at"`
	ExpiresAt    time.Time `bun:"expires_at"`
}

func SSHApprovalFromModel(model *models.SSHApproval) *SSHApproval {
	approval := &SSHApproval{
		Code:         model.Code,
		NamespaceID:  model.TenantID,
		Kind:         string(model.Kind),
		SessionUID:   model.SessionUID,
		SSHID:        model.SSHID,
		DeviceUID:    model.DeviceUID,
		DeviceName:   model.DeviceName,
		Username:     model.Username,
		IPAddress:    model.IPAddress,
		Fingerprint:  model.Fingerprint,
		Data:         model.Data,
		ReauthPeriod: model.ReauthPeriod,
		State:        string(model.State),
		RequestedAt:  model.RequestedAt,
		ExpiresAt:    model.ExpiresAt,
	}

	if model.DecidedBy != "" {
		approval.DecidedBy = &model.DecidedBy
	}

	return approval
}

func SSHApprovalToModel(e *SSHApproval) *models.SSHApproval {
	approval := &models.SSHApproval{
		Code:         e.Code,
		TenantID:     e.NamespaceID,
		Kind:         models.SSHApprovalKind(e.Kind),
		SessionUID:   e.SessionUID,
		SSHID:        e.SSHID,
		DeviceUID:    e.DeviceUID,
		DeviceName:   e.DeviceName,
		Username:     e.Username,
		IPAddress:    e.IPAddress,
		Fingerprint:  e.Fingerprint,
		Data:         e.Data,
		ReauthPeriod: e.ReauthPeriod,
		State:        models.SSHApprovalState(e.State),
		RequestedAt:  e.RequestedAt,
		ExpiresAt:    e.ExpiresAt,
	}

	if e.DecidedBy != nil {
		approval.DecidedBy = *e.DecidedBy
	}

	return approval
}
