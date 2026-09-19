package models

import "time"

// SSHApprovalKind is what confirming an approval actually does. A native SSH
// login can wait on either, and both act on the identity, not on the session:
// the session is only registered once the auth pipeline clears.
type SSHApprovalKind string

const (
	// SSHApprovalIdentity binds the presented key as a new identity.
	SSHApprovalIdentity SSHApprovalKind = "identity"
	// SSHApprovalReauth refreshes the re-auth window of an identity that already
	// exists, because a policy demands a fresh one. It creates nothing.
	SSHApprovalReauth SSHApprovalKind = "reauth"
)

// SSHApprovalState is the lifecycle of an approval. There is no stored "expired"
// state: a row past ExpiresAt reads as unknown, and a cron prunes it later.
type SSHApprovalState string

const (
	// SSHApprovalPending is an approval nobody has answered yet. A pending row past its ExpiresAt is
	// no longer pending — it is unknown.
	SSHApprovalPending SSHApprovalState = "pending"
	// SSHApprovalConfirmed is an approval a member granted; the parked login proceeds.
	SSHApprovalConfirmed SSHApprovalState = "confirmed"
	// SSHApprovalRejected is an approval a member denied; the parked login is refused.
	SSHApprovalRejected SSHApprovalState = "rejected"
)

// SSHApproval is a decision the SSH gateway parked while it holds a pure-OpenSSH
// login open, for a member to resolve in the console. The code is its identity
// and its secret: the gateway shows it on its approval prompt.
type SSHApproval struct {
	Code        string          `json:"code"`
	TenantID    string          `json:"tenant_id"`
	Kind        SSHApprovalKind `json:"kind"`
	SessionUID  string          `json:"session_uid"`
	SSHID       string          `json:"sshid"`
	DeviceUID   string          `json:"device_uid"`
	DeviceName  string          `json:"device_name"`
	Username    string          `json:"username"`
	IPAddress   string          `json:"ip_address"`
	Fingerprint string          `json:"fingerprint"`
	Data        []byte          `json:"data"`
	// ReauthPeriod is the policy's window in seconds, on a reauth approval. Nil or
	// zero means the policy asks every time.
	ReauthPeriod *int             `json:"reauth_period"`
	State        SSHApprovalState `json:"state"`
	// DecidedBy is the account that resolved the approval. On an identity
	// approval it is the account the key binds to, and the gateway adopts it as
	// the session's identity.
	DecidedBy string `json:"decided_by"`
	// ConfirmationCode is minted when the approval is confirmed and shown only in
	// the console, to the person who confirmed it. They type it at the terminal,
	// which is what proves they reached the console rather than merely following
	// a link. Empty until confirmed.
	ConfirmationCode string    `json:"confirmation_code"`
	RequestedAt      time.Time `json:"requested_at"`
	ExpiresAt        time.Time `json:"expires_at"`
}

// SSHApprovalCreated is the response to creating an approval: the short code the
// gateway prints, and the window the user has to decide.
type SSHApprovalCreated struct {
	Code      string `json:"code"`
	ExpiresIn int    `json:"expires_in_seconds"`
}

// SSHApprovalStatus is what the SSH gateway reads once, after the client answers its approval prompt.
// UserID carries the approving account once the decision is made, so the gateway
// can bind it to the session.
type SSHApprovalStatus struct {
	State  SSHApprovalState `json:"state"`
	UserID string           `json:"user_id,omitempty"`
	// ConfirmationCode is what the person was shown in the console and has to
	// type at the terminal. Empty unless State is confirmed.
	ConfirmationCode string `json:"confirmation_code,omitempty"`
}

// SSHApprovalConfirmation is the answer to confirming an approval: the code the
// person carries from the console to the terminal their login is waiting at.
type SSHApprovalConfirmation struct {
	ConfirmationCode string `json:"confirmation_code"`
}

// SSHApprovalRequest is the detail the console renders so the user sees which
// key and login they are deciding on.
type SSHApprovalRequest struct {
	SSHID       string           `json:"sshid"`
	DeviceName  string           `json:"device_name"`
	Username    string           `json:"username"`
	IPAddress   string           `json:"ip_address"`
	RequestedAt time.Time        `json:"requested_at"`
	State       SSHApprovalState `json:"state"`
	// Code echoes the correlation code so the page can display it for the user to
	// visually match against their approval prompt (anti-phishing).
	Code string `json:"code"`
	// Fingerprint is the presented key's fingerprint, shown front-and-center when
	// the key is becoming an identity.
	Fingerprint string `json:"fingerprint"`
	// Kind is what confirming does, and it is what the console branches the whole
	// screen on.
	Kind SSHApprovalKind `json:"kind"`
	// ReauthPeriod lets the console say how long confirming lasts, which is not
	// this login: the window is per identity, so other logins with the same key
	// skip the browser step until it lapses.
	ReauthPeriod *int `json:"reauth_period,omitempty"`
	// ExpiresIn is how much of the approval window is left, in seconds.
	ExpiresIn int `json:"expires_in_seconds"`
	// Namespace names where the key lands. The login carries it in the SSHID, so
	// it is not the console's current namespace: a member can approve a key into
	// a namespace they are not currently browsing.
	Namespace string `json:"namespace"`
}
