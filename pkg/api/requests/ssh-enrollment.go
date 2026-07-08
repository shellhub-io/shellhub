package requests

// SSHEnrollmentCreate is the payload the SSH gateway posts to mint a JIT login
// approval. The gateway has already resolved the target device and namespace,
// so it passes them through; the API does not re-parse the SSHID.
type SSHEnrollmentCreate struct {
	SessionUID string `json:"session_uid" validate:"required"`
	SSHID      string `json:"sshid" validate:"required"`
	TenantID   string `json:"tenant_id" validate:"required,uuid"`
	DeviceUID  string `json:"device_uid" validate:"required"`
	DeviceName string `json:"device_name" validate:""`
	Username   string `json:"username" validate:"required"`
	IPAddress  string `json:"ip_address" validate:"required"`
	// Fingerprint and Data carry the presented SSH public key when the approval
	// is an enrollment (identity mode). They are empty for a plain session
	// approval; the gateway attaches them once the key is known (see
	// SSHEnrollmentKey), since the SSH banner is minted before any key is
	// offered.
	Fingerprint string `json:"fingerprint" validate:""`
	Data        []byte `json:"data" validate:""`
}

// SSHEnrollmentKey attaches the presented key to a pending approval and marks
// whether accepting it should enroll the key as an identity. The gateway calls
// it from the public-key handler, after resolving the fingerprint, since the
// approval code (and its terminal banner) is minted before the key is offered.
type SSHEnrollmentKey struct {
	// Code comes from the path; json:"-" keeps the client from serializing an
	// empty Code into the body, which BindBody would otherwise use to overwrite
	// the path-bound value back to empty (failing validation).
	Code        string `param:"code" json:"-" validate:"required"`
	Fingerprint string `json:"fingerprint" validate:"required"`
	Data        []byte `json:"data" validate:"required"`
	// Enroll is true for JIT enrollment (unknown key) and false for a step-up
	// approval of an already-enrolled key.
	Enroll bool `json:"enroll" validate:""`
}

// SSHEnrollmentStatus is the request data for the status endpoint the gateway
// polls. The code is validated (normalized + checked) in the service.
type SSHEnrollmentStatus struct {
	Code string `param:"code" validate:"required"`
}

// SSHEnrollmentGet is the request data for the endpoint the console page uses to
// render the approval request details.
type SSHEnrollmentGet struct {
	Code string `param:"code" validate:"required"`
}

// SSHEnrollmentConfirm is the request data for the accept endpoint.
type SSHEnrollmentConfirm struct {
	Code string `param:"code" validate:"required"`
}

// SSHEnrollmentReject is the request data for the deny endpoint.
type SSHEnrollmentReject struct {
	Code string `param:"code" validate:"required"`
}
