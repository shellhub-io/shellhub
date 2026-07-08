package models

import "time"

// SSHEnrollmentState is the lifecycle of a JIT SSH key enrollment. There is no
// stored "expired" state: a code simply disappears from the cache when its TTL
// elapses, and a lookup miss is treated as expired.
type SSHEnrollmentState string

const (
	SSHEnrollmentPending   SSHEnrollmentState = "pending"
	SSHEnrollmentConfirmed SSHEnrollmentState = "confirmed"
	SSHEnrollmentRejected  SSHEnrollmentState = "rejected"
)

// SSHEnrollment is the response to an enrollment creation request: the short
// code the SSH gateway embeds in the terminal banner and the window the user has
// to complete it before it expires.
type SSHEnrollment struct {
	Code      string `json:"code"`
	ExpiresIn int    `json:"expires_in_seconds"`
}

// SSHEnrollmentStatus is what the SSH gateway polls while it holds the login
// open. UserID carries the enrolling account once the decision is made, so the
// gateway can bind it to the session.
type SSHEnrollmentStatus struct {
	State  SSHEnrollmentState `json:"state"`
	UserID string             `json:"user_id,omitempty"`
}

// SSHEnrollmentRequest is the detail the console renders on the enrollment page
// so the user sees which key and login they are enrolling before deciding.
type SSHEnrollmentRequest struct {
	SSHID       string             `json:"sshid"`
	DeviceName  string             `json:"device_name"`
	Username    string             `json:"username"`
	IPAddress   string             `json:"ip_address"`
	RequestedAt time.Time          `json:"requested_at"`
	State       SSHEnrollmentState `json:"state"`
	// Code echoes the correlation code so the enrollment page can display it for
	// the user to visually match against their terminal banner (anti-phishing).
	Code string `json:"code"`
	// Fingerprint is the presented key's fingerprint, shown front-and-center on
	// the enrollment page. Empty for a step-up of an already-enrolled key.
	Fingerprint string `json:"fingerprint"`
	// Enroll reports whether confirming this binds the presented key as a new
	// identity (JIT enrollment) versus a step-up confirmation of an already
	// enrolled key.
	Enroll bool `json:"enroll"`
}
