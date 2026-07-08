package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

// defaultSSHIdentityName derives a human-scannable name for an enrolled key when
// none is given (JIT enrollment from the browser): the key algorithm plus a
// short fingerprint slice, e.g. "ed25519 (jJhkjTqB)". The user can rename it.
func defaultSSHIdentityName(fingerprint string, data []byte) string {
	algo := "ssh key"
	if pubKey, _, _, _, err := ssh.ParseAuthorizedKey(data); err == nil { //nolint:dogsled
		// Type() is e.g. "ssh-ed25519", "ssh-rsa", "ecdsa-sha2-nistp256".
		algo = strings.TrimPrefix(pubKey.Type(), "ssh-")
		if i := strings.IndexByte(algo, '-'); i > 0 {
			algo = algo[:i] // "ecdsa-sha2-nistp256" -> "ecdsa"
		}
	}

	short := strings.TrimPrefix(fingerprint, "SHA256:")
	if len(short) > 8 {
		short = short[:8]
	}

	return fmt.Sprintf("%s (%s)", algo, short)
}

type SSHIdentityService interface {
	// ResolveSSHIdentity looks up an enrolled identity by namespace and
	// fingerprint. found is false (with a nil identity and nil error) when the
	// key is not enrolled. A recognized lookup stamps the identity's last-used
	// time.
	ResolveSSHIdentity(ctx context.Context, tenantID, fingerprint string) (identity *models.SSHIdentity, found bool, err error)

	// EnrollSSHIdentity binds a public key to userID within the namespace. It is
	// idempotent for the same account and returns a duplicated error when the
	// fingerprint is already bound to a different identity in the namespace. Used
	// by the approval accept (JIT) and manual add.
	EnrollSSHIdentity(ctx context.Context, userID, tenantID, fingerprint string, data []byte, name string) error

	// ListSSHIdentities returns the caller's enrolled identities in the namespace.
	// When all is true it returns every member's (the caller must hold
	// SSHIdentityManage, enforced at the handler).
	ListSSHIdentities(ctx context.Context, req *requests.SSHIdentityList) ([]models.SSHIdentity, error)

	// CreateSSHIdentity manually enrolls a pasted OpenSSH public key for the
	// caller and returns the stored identity.
	CreateSSHIdentity(ctx context.Context, req *requests.SSHIdentityCreate) (*models.SSHIdentity, error)

	// RenameSSHIdentity renames one of the caller's own identities.
	RenameSSHIdentity(ctx context.Context, req *requests.SSHIdentityUpdate) (*models.SSHIdentity, error)

	// DeleteSSHIdentity revokes an identity. Revoking the caller's own needs
	// SSHIdentityEnroll; revoking another member's needs SSHIdentityManage
	// (signalled by req.Manage, resolved at the handler).
	DeleteSSHIdentity(ctx context.Context, req *requests.SSHIdentityDelete) error
}

func (s *service) ResolveSSHIdentity(ctx context.Context, tenantID, fingerprint string) (*models.SSHIdentity, bool, error) {
	identity, err := s.store.SSHIdentityResolve(ctx, store.SSHIdentityFingerprintResolver, fingerprint, s.store.Options().InNamespace(tenantID))
	if err != nil {
		if errors.Is(err, store.ErrNoDocuments) {
			return nil, false, nil
		}

		return nil, false, err
	}

	// A recognized connect stamps last-used, feeding the management screen and
	// later stale-key cleanup. A failure here must not fail the connection.
	if err := s.store.SSHIdentityTouchLastUsed(ctx, tenantID, fingerprint); err != nil {
		log.WithError(err).WithField("fingerprint", fingerprint).
			Warn("failed to stamp ssh identity last-used; connection proceeds")
	}

	return identity, true, nil
}

func (s *service) EnrollSSHIdentity(ctx context.Context, userID, tenantID, fingerprint string, data []byte, name string) error {
	_, err := s.enrollSSHIdentity(ctx, userID, tenantID, fingerprint, data, name)

	return err
}

// enrollSSHIdentity creates the binding, returning the existing one unchanged
// when the same account already holds it (idempotent) and a duplicated error
// when the fingerprint is taken by another identity in the namespace.
func (s *service) enrollSSHIdentity(ctx context.Context, userID, tenantID, fingerprint string, data []byte, name string) (*models.SSHIdentity, error) {
	existing, err := s.store.SSHIdentityResolve(ctx, store.SSHIdentityFingerprintResolver, fingerprint, s.store.Options().InNamespace(tenantID))
	if err != nil && !errors.Is(err, store.ErrNoDocuments) {
		return nil, err
	}

	if existing != nil {
		if existing.UserID == userID {
			return existing, nil
		}

		return nil, NewErrSSHIdentityDuplicated(fingerprint, nil)
	}

	if name == "" {
		name = defaultSSHIdentityName(fingerprint, data)
	}

	identity := &models.SSHIdentity{
		TenantID:    tenantID,
		UserID:      userID,
		Fingerprint: fingerprint,
		Data:        data,
		Name:        name,
	}

	id, err := s.store.SSHIdentityCreate(ctx, identity)
	if err != nil {
		if errors.Is(err, store.ErrDuplicate) {
			return nil, NewErrSSHIdentityDuplicated(fingerprint, err)
		}

		return nil, err
	}

	identity.ID = id

	return identity, nil
}

func (s *service) ListSSHIdentities(ctx context.Context, req *requests.SSHIdentityList) ([]models.SSHIdentity, error) {
	opts := []store.QueryOption{s.store.Options().InNamespace(req.TenantID)}
	if !req.All {
		opts = append(opts, s.store.Options().WithUserID(req.UserID))
	}

	identities, _, err := s.store.SSHIdentityList(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return identities, nil
}

func (s *service) CreateSSHIdentity(ctx context.Context, req *requests.SSHIdentityCreate) (*models.SSHIdentity, error) {
	// data is a raw OpenSSH authorized_keys line; derive the fingerprint the same
	// way the gateway does at connect (SHA256) so a manually-added key resolves.
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(req.Data)) //nolint:dogsled
	if err != nil {
		return nil, NewErrSSHIdentityInvalid(req.Data, err)
	}

	fingerprint := ssh.FingerprintSHA256(pubKey)
	data := ssh.MarshalAuthorizedKey(pubKey)

	return s.enrollSSHIdentity(ctx, req.UserID, req.TenantID, fingerprint, data, req.Name)
}

func (s *service) RenameSSHIdentity(ctx context.Context, req *requests.SSHIdentityUpdate) (*models.SSHIdentity, error) {
	identity, err := s.store.SSHIdentityResolve(ctx, store.SSHIdentityIDResolver, req.ID, s.store.Options().InNamespace(req.TenantID))
	if err != nil {
		return nil, NewErrSSHIdentityNotFound(req.ID, err)
	}

	// Renaming is own-key only (SSHIdentityEnroll); managing others' keys is
	// limited to revocation.
	if identity.UserID != req.UserID {
		return nil, NewErrForbidden(ErrForbidden, nil)
	}

	identity.Name = req.Name

	if err := s.store.SSHIdentityUpdate(ctx, identity); err != nil {
		return nil, err
	}

	return identity, nil
}

func (s *service) DeleteSSHIdentity(ctx context.Context, req *requests.SSHIdentityDelete) error {
	identity, err := s.store.SSHIdentityResolve(ctx, store.SSHIdentityIDResolver, req.ID, s.store.Options().InNamespace(req.TenantID))
	if err != nil {
		return NewErrSSHIdentityNotFound(req.ID, err)
	}

	// A member revokes only their own keys; revoking another member's requires
	// the manage permission (offboarding), signalled by req.Manage.
	if identity.UserID != req.UserID && !req.Manage {
		return NewErrForbidden(ErrForbidden, nil)
	}

	return s.store.SSHIdentityDelete(ctx, &models.SSHIdentity{ID: req.ID, TenantID: req.TenantID})
}
