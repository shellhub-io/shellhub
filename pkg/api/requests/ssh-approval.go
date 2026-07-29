package requests

import "github.com/shellhub-io/shellhub/pkg/models"

// SSHApprovalCreate is the payload the SSH gateway posts to open a JIT login
// approval. The gateway only posts it once it knows a browser step is actually
// needed, so the presented key and the kind come with it. It has already
// resolved the target device and namespace, so it passes them through; the API
// does not re-parse the SSHID.
type SSHApprovalCreate struct {
	SessionUID string `json:"session_uid" validate:"required"`
	SSHID      string `json:"sshid" validate:"required"`
	TenantID   string `json:"tenant_id" validate:"required,uuid"`
	DeviceUID  string `json:"device_uid" validate:"required"`
	DeviceName string `json:"device_name" validate:""`
	Username   string `json:"username" validate:"required"`
	IPAddress  string `json:"ip_address" validate:"required"`
	// Kind is what confirming will do: bind the key as an identity, or refresh an
	// existing identity's re-auth window.
	Kind        models.SSHApprovalKind `json:"kind" validate:"required,oneof=identity reauth"`
	Fingerprint string                 `json:"fingerprint" validate:"required"`
	Data        []byte                 `json:"data" validate:"required"`
	// ReauthPeriod carries the policy's window so the console can say how long
	// confirming lasts. Only meaningful for the reauth kind.
	ReauthPeriod *int `json:"reauth_period" validate:""`
}

// SSHApprovalStatus is the request data for the status endpoint the gateway
// polls. The code is validated (normalized + checked) in the service.
type SSHApprovalStatus struct {
	Code string `param:"code" validate:"required"`
	// Wait asks the API to hold the request open until the login is decided
	// instead of answering "pending" right away. The gateway sets it so the
	// person is not left staring at a frozen terminal for a poll interval after
	// they already decided.
	Wait bool `query:"wait" validate:""`
}

// SSHApprovalGet is the request data for the endpoint the console page uses to
// render the approval request details.
type SSHApprovalGet struct {
	Code string `param:"code" validate:"required"`
}

// SSHApprovalConfirm is the request data for the accept endpoint.
type SSHApprovalConfirm struct {
	Code string `param:"code" validate:"required"`
	// ExpiresIn is how many days the key being bound should keep working,
	// omitted for a key that never expires. Only meaningful for the identity
	// kind, since a re-auth binds nothing.
	ExpiresIn *int `json:"expires_in" validate:"omitempty,min=1"`
}

// SSHApprovalReject is the request data for the deny endpoint.
type SSHApprovalReject struct {
	Code string `param:"code" validate:"required"`
}
