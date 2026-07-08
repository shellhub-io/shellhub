package internalclient

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
)

// sshEnrollmentAPI defines methods the SSH gateway uses to hold a login open
// while the connecting user enrolls the presented key in the console.
type sshEnrollmentAPI interface {
	// CreateSSHEnrollment mints a JIT enrollment code for a held-open SSH login.
	CreateSSHEnrollment(ctx context.Context, req requests.SSHEnrollmentCreate) (*models.SSHEnrollment, error)

	// GetSSHEnrollmentStatus polls the decision; it returns the state and, once
	// confirmed, the enrolling user id.
	GetSSHEnrollmentStatus(ctx context.Context, code string) (*models.SSHEnrollmentStatus, error)

	// AttachSSHEnrollmentKey attaches the presented key to a pending enrollment
	// so the console page can show its fingerprint and the confirm can bind it.
	// It is a separate call because the enrollment (and its terminal banner) is
	// minted before the public key is offered in the handshake. Enroll marks a
	// JIT enrollment (unknown key); false is a step-up of an already-enrolled key.
	AttachSSHEnrollmentKey(ctx context.Context, code, fingerprint string, data []byte, enroll bool) error
}

func (c *client) CreateSSHEnrollment(ctx context.Context, req requests.SSHEnrollmentCreate) (*models.SSHEnrollment, error) {
	enrollment := new(models.SSHEnrollment)

	resp, err := c.http.
		R().
		SetContext(ctx).
		SetBody(req).
		SetResult(enrollment).
		Post(c.config.APIBaseURL + "/internal/sshid/enrollment")
	if err := HasError(resp, err); err != nil {
		return nil, err
	}

	return enrollment, nil
}

func (c *client) GetSSHEnrollmentStatus(ctx context.Context, code string) (*models.SSHEnrollmentStatus, error) {
	status := new(models.SSHEnrollmentStatus)

	resp, err := c.http.
		R().
		SetContext(ctx).
		SetPathParam("code", code).
		SetResult(status).
		Get(c.config.APIBaseURL + "/internal/sshid/enrollment/{code}/status")
	if err := HasError(resp, err); err != nil {
		return nil, err
	}

	return status, nil
}

func (c *client) AttachSSHEnrollmentKey(ctx context.Context, code, fingerprint string, data []byte, enroll bool) error {
	resp, err := c.http.
		R().
		SetContext(ctx).
		SetPathParam("code", code).
		SetBody(requests.SSHEnrollmentKey{
			Fingerprint: fingerprint,
			Data:        data,
			Enroll:      enroll,
		}).
		Patch(c.config.APIBaseURL + "/internal/sshid/enrollment/{code}/key")

	return HasError(resp, err)
}
