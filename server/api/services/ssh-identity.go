package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shellhub-io/shellhub/pkg/clock"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

// defaultSSHIdentityName derives a human-scannable name for an enrolled key when
// none is given (a key accepted at login): the key algorithm plus a short
// fingerprint slice, e.g. "ed25519 (jJhkjTqB)". The user can rename it.
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

// sshIdentityExpiry turns a TTL in days into the absolute instant the key stops
// working. A nil TTL means it never expires, which is what a durable key wants.
func sshIdentityExpiry(days *int) *time.Time {
	if days == nil {
		return nil
	}

	at := clock.Now().AddDate(0, 0, *days)

	return &at
}

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

	// A recognized connect stamps last-used, feeding the management screen and
	// later stale-key cleanup. A failure here must not fail the connection.
	if err := s.store.SSHIdentityTouchLastUsed(ctx, tenantID, fingerprint); err != nil {
		log.WithError(err).WithField("fingerprint", fingerprint).
			Warn("failed to stamp ssh identity last-used; connection proceeds")
	}

	return identity, true, nil
}

func (s *service) ConsumeSSHIdentity(ctx context.Context, tenantID, fingerprint string) (bool, error) {
	return s.store.SSHIdentityConsume(ctx, tenantID, fingerprint)
}

// enrollSSHIdentity creates the binding, failing with a duplicated error when
// the fingerprint is already taken in the namespace — by this principal or
// another. A caller re-adding a key it already holds is a mistake worth
// reporting, not a no-op to hide; the one path that must tolerate a replay uses
// reenrollSSHIdentity. A zero ExpiresAt and SingleUse give the durable,
// reusable key most callers want.
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

// reenrollSSHIdentity is enrollSSHIdentity for the paths that can legitimately
// run twice for the same key — a confirmation replayed by the person or the
// gateway — returning the existing binding unchanged instead of a conflict. The
// fingerprint held by a different principal still conflicts.
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

// resolveEnrolledSSHIdentity reports a free fingerprint as a nil identity rather
// than an error, so the enroll paths branch on the holder instead of on
// ErrNoDocuments.
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
	// data is a raw OpenSSH authorized_keys line; derive the fingerprint the same
	// way the gateway does at connect (SHA256) so a manually-added key resolves.
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(req.Data)) //nolint:dogsled
	if err != nil {
		return nil, NewErrSSHIdentityInvalid(req.Data, err)
	}

	fingerprint := ssh.FingerprintSHA256(pubKey)
	data := ssh.MarshalAuthorizedKey(pubKey)

	// The source is the caller's claim here, so it may only ever label the key.
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

	// Renaming is own-key only (SSHIdentityAdd); managing others' keys is
	// limited to revocation.
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

	// A member revokes only their own keys; revoking another member's requires
	// the manage permission (offboarding), signalled by req.Manage.
	if identity.PrincipalID != req.UserID && !req.Manage {
		return NewErrForbidden(ErrForbidden, nil)
	}

	return s.store.SSHIdentityDelete(ctx, &models.SSHIdentity{ID: req.ID, TenantID: req.TenantID})
}
