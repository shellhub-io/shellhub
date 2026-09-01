package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

func defaultSSHIdentityName(fingerprint string, data []byte) string {
	algo := "ssh key"
	if pubKey, _, _, _, err := ssh.ParseAuthorizedKey(data); err == nil { //nolint:dogsled
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

func sshIdentityExpiry(days *int) *time.Time {
	if days == nil {
		return nil
	}

	at := clock.Now().AddDate(0, 0, *days)

	return &at
}

// SSHIdentityService owns the enrolled public keys that identify a person to a device, as
// opposed to the namespace-wide keys in [SSHKeysService].
type SSHIdentityService interface {
	// ResolveSSHIdentity looks up an enrolled identity by namespace and
	// fingerprint. found is false (with a nil identity and nil error) when the
	// key is not enrolled. A recognized lookup stamps the identity's last-used
	// time.
	ResolveSSHIdentity(ctx context.Context, tenantID, fingerprint string) (identity *models.SSHIdentity, found bool, err error)

	// ConsumeSSHIdentity atomically burns a single-use identity once its session
	// is established, returning whether this call won the burn. A false means a
	// concurrent session already consumed the key and the caller must be denied.
	ConsumeSSHIdentity(ctx context.Context, tenantID, fingerprint string) (bool, error)

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
	// SSHIdentityAdd; revoking another member's needs SSHIdentityManage
	// (signalled by req.Manage, resolved at the handler).
	DeleteSSHIdentity(ctx context.Context, req *requests.SSHIdentityDelete) error
}

func (s *service) ResolveSSHIdentity(ctx context.Context, tenantID, fingerprint string) (*models.SSHIdentity, bool, error) {
	sc, err := BoundTo(tenantID)
	if err != nil {
		return nil, false, err
	}

	identity, err := s.store.SSHIdentityResolve(ctx, sc, store.SSHIdentityFingerprintResolver, fingerprint)
	if err != nil {
		if errors.Is(err, store.ErrNoDocuments) {
			return nil, false, nil
		}

		return nil, false, err
	}

	if err := s.store.SSHIdentityTouchLastUsed(ctx, tenantID, fingerprint); err != nil {
		log.WithError(err).WithField("fingerprint", fingerprint).
			Warn("failed to stamp ssh identity last-used; connection proceeds")
	}

	return identity, true, nil
}

func (s *service) ConsumeSSHIdentity(ctx context.Context, tenantID, fingerprint string) (bool, error) {
	return s.store.SSHIdentityConsume(ctx, tenantID, fingerprint)
}

func (s *service) enrollSSHIdentity(ctx context.Context, identity *models.SSHIdentity) (*models.SSHIdentity, error) {
	existing, err := s.resolveEnrolledSSHIdentity(ctx, identity.TenantID, identity.Fingerprint)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		return nil, NewErrSSHIdentityDuplicated(identity.Fingerprint, nil)
	}

	return s.persistSSHIdentity(ctx, identity)
}

func (s *service) reenrollSSHIdentity(ctx context.Context, identity *models.SSHIdentity) (*models.SSHIdentity, error) {
	existing, err := s.resolveEnrolledSSHIdentity(ctx, identity.TenantID, identity.Fingerprint)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		if existing.PrincipalID == identity.PrincipalID {
			return existing, nil
		}

		return nil, NewErrSSHIdentityDuplicated(identity.Fingerprint, nil)
	}

	return s.persistSSHIdentity(ctx, identity)
}

func (s *service) resolveEnrolledSSHIdentity(ctx context.Context, tenantID, fingerprint string) (*models.SSHIdentity, error) {
	sc, err := BoundTo(tenantID)
	if err != nil {
		return nil, err
	}

	existing, err := s.store.SSHIdentityResolve(ctx, sc, store.SSHIdentityFingerprintResolver, fingerprint)
	if err != nil && !errors.Is(err, store.ErrNoDocuments) {
		return nil, err
	}

	return existing, nil
}

func (s *service) persistSSHIdentity(ctx context.Context, identity *models.SSHIdentity) (*models.SSHIdentity, error) {
	if identity.Name == "" {
		identity.Name = defaultSSHIdentityName(identity.Fingerprint, identity.Data)
	}

	if identity.Source == "" {
		identity.Source = models.SSHIdentitySourceManual
	}

	id, err := s.store.SSHIdentityCreate(ctx, identity)
	if err != nil {
		if errors.Is(err, store.ErrDuplicate) {
			return nil, NewErrSSHIdentityDuplicated(identity.Fingerprint, err)
		}

		return nil, err
	}

	identity.ID = id

	return identity, nil
}

func (s *service) ListSSHIdentities(ctx context.Context, req *requests.SSHIdentityList) ([]models.SSHIdentity, error) {
	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return nil, err
	}

	var opts []store.QueryOption
	if !req.All {
		opts = append(opts, s.store.Options().WithUserID(req.UserID))
	}

	identities, _, err := s.store.SSHIdentityList(ctx, sc, opts...)
	if err != nil {
		return nil, err
	}

	return identities, nil
}

func (s *service) CreateSSHIdentity(ctx context.Context, req *requests.SSHIdentityCreate) (*models.SSHIdentity, error) {
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(req.Data)) //nolint:dogsled
	if err != nil {
		return nil, NewErrSSHIdentityInvalid(req.Data, err)
	}

	fingerprint := ssh.FingerprintSHA256(pubKey)
	data := ssh.MarshalAuthorizedKey(pubKey)

	return s.enrollSSHIdentity(ctx, &models.SSHIdentity{
		TenantID:    req.TenantID,
		PrincipalID: req.UserID,
		Fingerprint: fingerprint,
		Data:        data,
		Name:        req.Name,
		Source:      req.Source,
		ExpiresAt:   sshIdentityExpiry(req.ExpiresIn),
	})
}

func (s *service) RenameSSHIdentity(ctx context.Context, req *requests.SSHIdentityUpdate) (*models.SSHIdentity, error) {
	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return nil, err
	}

	identity, err := s.store.SSHIdentityResolve(ctx, sc, store.SSHIdentityIDResolver, req.ID)
	if err != nil {
		return nil, NewErrSSHIdentityNotFound(req.ID, err)
	}

	if identity.PrincipalID != req.UserID {
		return nil, NewErrForbidden(ErrForbidden, nil)
	}

	identity.Name = req.Name

	if err := s.store.SSHIdentityUpdate(ctx, identity); err != nil {
		return nil, err
	}

	return identity, nil
}

func (s *service) DeleteSSHIdentity(ctx context.Context, req *requests.SSHIdentityDelete) error {
	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return err
	}

	identity, err := s.store.SSHIdentityResolve(ctx, sc, store.SSHIdentityIDResolver, req.ID)
	if err != nil {
		return NewErrSSHIdentityNotFound(req.ID, err)
	}

	if identity.PrincipalID != req.UserID && !req.Manage {
		return NewErrForbidden(ErrForbidden, nil)
	}

	return s.store.SSHIdentityDelete(ctx, &models.SSHIdentity{ID: req.ID, TenantID: req.TenantID})
}
