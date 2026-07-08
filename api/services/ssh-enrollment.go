package services

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/pairingcode"
	log "github.com/sirupsen/logrus"
)

// sshEnrollmentTTL is how long an enrollment code remains valid. It is a
// just-in-time gate: the SSH gateway holds the login open for this long while
// the user opens the console and enrolls the key. It must stay under the
// gateway's wait timeout so the code never outlives the connection waiting on it.
const sshEnrollmentTTL = 90 * time.Second

// sshEnrollment is the payload cached under `ssh_enrollment/<code>`. The SSH
// gateway mints it and polls it; a logged-in console user resolves it and writes
// the decision back. UserID is the account bound to the session once the
// enrollment is confirmed.
type sshEnrollment struct {
	SessionUID  string    `json:"session_uid"`
	SSHID       string    `json:"sshid"`
	TenantID    string    `json:"tenant_id"`
	DeviceUID   string    `json:"device_uid"`
	DeviceName  string    `json:"device_name"`
	Username    string    `json:"username"`
	IPAddress   string    `json:"ip_address"`
	RequestedAt time.Time `json:"requested_at"`

	// Fingerprint and Data carry the presented SSH public key. They are attached
	// after the enrollment is minted, once the gateway sees the key (see
	// AttachSSHEnrollmentKey).
	Fingerprint string `json:"fingerprint,omitempty"`
	Data        []byte `json:"data,omitempty"`
	// Enroll marks an enrollment whose confirm must bind the presented key as a
	// new identity (JIT enrollment), versus a step-up confirmation.
	Enroll bool `json:"enroll,omitempty"`

	State  models.SSHEnrollmentState `json:"state"`
	UserID string                    `json:"user_id,omitempty"`
}

type SSHEnrollmentService interface {
	// CreateSSHEnrollment stores a pending JIT enrollment and returns a
	// short-lived code the SSH gateway embeds in the terminal banner. The code
	// itself is the secret; it expires with a short TTL.
	CreateSSHEnrollment(ctx context.Context, req *requests.SSHEnrollmentCreate) (*models.SSHEnrollment, error)

	// GetSSHEnrollmentStatus reports the decision to the gateway polling it. An
	// unknown or expired code returns not found.
	GetSSHEnrollmentStatus(ctx context.Context, code string) (*models.SSHEnrollmentStatus, error)

	// GetSSHEnrollment returns the request details the console renders on the
	// enrollment page.
	GetSSHEnrollment(ctx context.Context, code string) (*models.SSHEnrollmentRequest, error)

	// AttachSSHEnrollmentKey attaches the presented key to a pending enrollment
	// so the console page can show its fingerprint and the confirm can bind it.
	// It is called by the gateway after resolving the fingerprint, since the
	// code is minted before any key is offered in the handshake.
	AttachSSHEnrollmentKey(ctx context.Context, req *requests.SSHEnrollmentKey) error

	// ConfirmSSHEnrollment confirms a pending enrollment and binds the enrolling
	// user to it. The user must be a member of the target's namespace with the
	// session approve permission.
	ConfirmSSHEnrollment(ctx context.Context, userID string, req *requests.SSHEnrollmentConfirm) error

	// RejectSSHEnrollment rejects a pending enrollment. Same authorization as
	// confirm.
	RejectSSHEnrollment(ctx context.Context, userID string, req *requests.SSHEnrollmentReject) error
}

func (s *service) CreateSSHEnrollment(ctx context.Context, req *requests.SSHEnrollmentCreate) (*models.SSHEnrollment, error) {
	code, err := pairingcode.New(pairingcode.DeviceCodeLength)
	if err != nil {
		return nil, err
	}

	enrollment := &sshEnrollment{
		SessionUID:  req.SessionUID,
		SSHID:       req.SSHID,
		TenantID:    req.TenantID,
		DeviceUID:   req.DeviceUID,
		DeviceName:  req.DeviceName,
		Username:    req.Username,
		IPAddress:   req.IPAddress,
		Fingerprint: req.Fingerprint,
		Data:        req.Data,
		RequestedAt: clock.Now(),
		State:       models.SSHEnrollmentPending,
	}

	if err := s.cache.Set(ctx, "ssh_enrollment/"+code, enrollment, sshEnrollmentTTL); err != nil {
		return nil, err
	}

	return &models.SSHEnrollment{
		Code:      code,
		ExpiresIn: int(sshEnrollmentTTL.Seconds()),
	}, nil
}

func (s *service) GetSSHEnrollmentStatus(ctx context.Context, code string) (*models.SSHEnrollmentStatus, error) {
	code = pairingcode.Normalize(code)

	enrollment := new(sshEnrollment)
	if err := s.cache.Get(ctx, "ssh_enrollment/"+code, enrollment); err != nil || enrollment.State == "" {
		return nil, NewErrSSHEnrollmentCodeNotFound(code, err)
	}

	return &models.SSHEnrollmentStatus{
		State:  enrollment.State,
		UserID: enrollment.UserID,
	}, nil
}

func (s *service) GetSSHEnrollment(ctx context.Context, code string) (*models.SSHEnrollmentRequest, error) {
	code = pairingcode.Normalize(code)

	enrollment := new(sshEnrollment)
	if err := s.cache.Get(ctx, "ssh_enrollment/"+code, enrollment); err != nil || enrollment.State == "" {
		return nil, NewErrSSHEnrollmentCodeNotFound(code, err)
	}

	return &models.SSHEnrollmentRequest{
		SSHID:       enrollment.SSHID,
		DeviceName:  enrollment.DeviceName,
		Username:    enrollment.Username,
		IPAddress:   enrollment.IPAddress,
		RequestedAt: enrollment.RequestedAt,
		State:       enrollment.State,
		Code:        code,
		Fingerprint: enrollment.Fingerprint,
		Enroll:      enrollment.Enroll,
	}, nil
}

// AttachSSHEnrollmentKey records the presented key on a pending enrollment. It
// preserves the existing fields and refreshes the TTL. A missing or non-pending
// code is treated as not found.
func (s *service) AttachSSHEnrollmentKey(ctx context.Context, req *requests.SSHEnrollmentKey) error {
	code := pairingcode.Normalize(req.Code)
	if !pairingcode.IsValid(code, pairingcode.DeviceCodeLength) {
		return NewErrSSHEnrollmentCodeNotFound(code, nil)
	}

	enrollment := new(sshEnrollment)
	if err := s.cache.Get(ctx, "ssh_enrollment/"+code, enrollment); err != nil || enrollment.State != models.SSHEnrollmentPending {
		return NewErrSSHEnrollmentCodeNotFound(code, err)
	}

	enrollment.Fingerprint = req.Fingerprint
	enrollment.Data = req.Data
	enrollment.Enroll = req.Enroll

	return s.cache.Set(ctx, "ssh_enrollment/"+code, enrollment, sshEnrollmentTTL)
}

func (s *service) ConfirmSSHEnrollment(ctx context.Context, userID string, req *requests.SSHEnrollmentConfirm) error {
	return s.decideSSHEnrollment(ctx, userID, req.Code, models.SSHEnrollmentConfirmed)
}

func (s *service) RejectSSHEnrollment(ctx context.Context, userID string, req *requests.SSHEnrollmentReject) error {
	return s.decideSSHEnrollment(ctx, userID, req.Code, models.SSHEnrollmentRejected)
}

// decideSSHEnrollment writes the confirm/reject decision after checking that the
// user is a member of the target's namespace with the approve permission. A
// SetNX reservation makes the decision single-use so a double submit can't race
// two writes onto the same code.
func (s *service) decideSSHEnrollment(ctx context.Context, userID, code string, decision models.SSHEnrollmentState) error {
	code = pairingcode.Normalize(code)
	if !pairingcode.IsValid(code, pairingcode.DeviceCodeLength) {
		return NewErrSSHEnrollmentCodeNotFound(code, nil)
	}

	enrollment := new(sshEnrollment)
	if err := s.cache.Get(ctx, "ssh_enrollment/"+code, enrollment); err != nil || enrollment.State != models.SSHEnrollmentPending {
		return NewErrSSHEnrollmentCodeNotFound(code, err)
	}

	// The console session may be scoped to another namespace, so the gateway's
	// permission middleware cannot cover this route; check the user's role in the
	// target namespace explicitly.
	namespace, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, enrollment.TenantID)
	if err != nil {
		return NewErrNamespaceNotFound(enrollment.TenantID, err)
	}

	member, ok := namespace.FindMember(userID)
	if !ok {
		return NewErrNamespaceMemberNotFound(userID, nil)
	}

	if !member.Role.HasPermission(authorizer.SessionApprove) {
		return NewErrRoleForbidden()
	}

	// Enrolling a key binds it to the user's account; require the enroll
	// permission on top of the approve permission (observers cannot enroll).
	if decision == models.SSHEnrollmentConfirmed && enrollment.Enroll && !member.Role.HasPermission(authorizer.SSHIdentityEnroll) {
		return NewErrRoleForbidden()
	}

	// Reserve the code before writing so exactly one concurrent decision wins.
	reserved, err := s.cache.SetNX(ctx, "ssh_enrollment_decision/"+code, userID, sshEnrollmentTTL)
	if err != nil {
		return err
	}

	if !reserved {
		return NewErrSSHEnrollmentCodeNotFound(code, nil)
	}

	// JIT enrollment: a confirmed enrollment binds the presented key to the
	// account before the outcome is published, so the gateway's poll sees a
	// confirmed state only once the identity exists. EnrollSSHIdentity is
	// idempotent for the same account, so a re-confirm does not fail. An empty
	// name lets the identity service generate a default from the key.
	if decision == models.SSHEnrollmentConfirmed && enrollment.Enroll && enrollment.Fingerprint != "" {
		if err := s.EnrollSSHIdentity(ctx, userID, enrollment.TenantID, enrollment.Fingerprint, enrollment.Data, ""); err != nil {
			return err
		}
	}

	enrollment.State = decision
	enrollment.UserID = userID

	if err := s.cache.Set(ctx, "ssh_enrollment/"+code, enrollment, sshEnrollmentTTL); err != nil {
		log.WithError(err).WithField("session_uid", enrollment.SessionUID).
			Warn("ssh enrollment decided but failed to store the outcome; the gateway will not see it")

		return err
	}

	return nil
}
