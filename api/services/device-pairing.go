package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
)

// devicePairingTTL is how long a pairing code remains valid, and also how long
// the accepted result stays available for the agent to poll after the user
// accepts.
const devicePairingTTL = 10 * time.Minute

// devicePairing is the payload cached under `pairing_code/<code>`. It carries
// the identity a tenant-less agent submitted and, once accepted, the outcome.
type devicePairing struct {
	Hostname  string                 `json:"hostname"`
	Identity  *models.DeviceIdentity `json:"identity"`
	Info      *models.DeviceInfo     `json:"info"`
	PublicKey string                 `json:"public_key"`

	Status   models.DeviceStatus `json:"status"`
	TenantID string              `json:"tenant_id"`
	UID      string              `json:"uid"`
}

type DevicePairingService interface {
	// CreateDevicePairing stores the identity payload of a tenant-less agent and
	// returns a short-lived code that deep-links it into the console's accept
	// page. No device exists until a user accepts the pairing into a namespace.
	CreateDevicePairing(ctx context.Context, req *requests.DevicePairingCreate) (*models.DevicePairing, error)

	// GetDevicePairingStatus reports the pairing outcome to the agent. The code
	// itself is the secret; unknown or expired codes return not found.
	GetDevicePairingStatus(ctx context.Context, code string) (*models.DevicePairingStatus, error)

	// AcceptDevicePairing materializes the pairing payload as a device in the
	// chosen namespace and accepts it. The user must be a member of the chosen
	// namespace with the device accept permission.
	AcceptDevicePairing(ctx context.Context, userID string, req *requests.DevicePairingAccept) (*models.DevicePairingAccepted, error)
}

func (s *service) CreateDevicePairing(ctx context.Context, req *requests.DevicePairingCreate) (*models.DevicePairing, error) {
	// Resume: if this public key already belongs to an accepted device (the
	// agent crashed after acceptance but before persisting its tenant, or the
	// previous accept's cache writeback was lost), hand the tenant back right
	// away instead of starting a new pairing the user would have to accept
	// again. The public key is not exposed by any read API, so disclosing the
	// tenant to whoever holds it is acceptable; possession of the matching
	// private key is still required to actually operate as the device.
	if device, err := s.store.DeviceResolve(ctx, store.DevicePublicKeyResolver, req.PublicKey, s.store.Options().WithDeviceStatus(models.DeviceStatusAccepted)); err == nil && device != nil {
		return &models.DevicePairing{Status: models.DeviceStatusAccepted, TenantID: device.TenantID}, nil
	}

	// Dedup by public key: the daemon and the `login` command both request a
	// pairing for the same device (same key), so the server is the single point
	// that coordinates them. If a live code already exists for this key, return
	// the same one instead of minting a second — one code, no agent-side IPC.
	// The server owns the code lifecycle (it issues it; there is no external IdP
	// in the loop), so dedup belongs here, not on the agent.
	pubKeyRef := "pairing_code_pubkey/" + hashPublicKey(req.PublicKey)

	var existingCode string
	if err := s.cache.Get(ctx, pubKeyRef, &existingCode); err == nil && existingCode != "" {
		existing := new(devicePairing)
		if err := s.cache.Get(ctx, "pairing_code/"+existingCode, existing); err == nil && existing.PublicKey != "" {
			return &models.DevicePairing{
				Code:      existingCode,
				ExpiresIn: int(devicePairingTTL.Seconds()),
				Status:    existing.Status,
				TenantID:  existing.TenantID,
			}, nil
		}
	}

	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}

	code := hex.EncodeToString(buf)

	pairing := &devicePairing{
		Hostname:  req.Hostname,
		PublicKey: req.PublicKey,
		Status:    models.DeviceStatusPending,
	}

	if req.Identity != nil {
		pairing.Identity = &models.DeviceIdentity{MAC: req.Identity.MAC}
	}

	if req.Info != nil {
		pairing.Info = &models.DeviceInfo{
			ID:         req.Info.ID,
			PrettyName: req.Info.PrettyName,
			Version:    req.Info.Version,
			Arch:       req.Info.Arch,
			Platform:   req.Info.Platform,
		}
	}

	if err := s.cache.Set(ctx, "pairing_code/"+code, pairing, devicePairingTTL); err != nil {
		return nil, err
	}

	// Map the public key to its current code so a concurrent request for the
	// same device reuses it. Same TTL, so both expire together.
	if err := s.cache.Set(ctx, pubKeyRef, code, devicePairingTTL); err != nil {
		log.WithError(err).Warn("failed to store the pairing dedup reference; a duplicate code may be minted")
	}

	return &models.DevicePairing{
		Code:      code,
		ExpiresIn: int(devicePairingTTL.Seconds()),
		Status:    models.DeviceStatusPending,
	}, nil
}

// hashPublicKey derives a cache-key-safe identifier for a device public key
// (the PEM contains newlines and is long), used to map a key to its live code.
func hashPublicKey(publicKey string) string {
	sum := sha256.Sum256([]byte(publicKey))

	return hex.EncodeToString(sum[:])
}

func (s *service) GetDevicePairingStatus(ctx context.Context, code string) (*models.DevicePairingStatus, error) {
	pairing := new(devicePairing)
	// NOTE: A cache miss is not an error; it leaves the value untouched.
	if err := s.cache.Get(ctx, "pairing_code/"+code, pairing); err != nil || pairing.PublicKey == "" {
		return nil, NewErrDeviceLoginCodeNotFound(code, err)
	}

	return &models.DevicePairingStatus{Status: pairing.Status, TenantID: pairing.TenantID}, nil
}

func (s *service) AcceptDevicePairing(ctx context.Context, userID string, req *requests.DevicePairingAccept) (*models.DevicePairingAccepted, error) {
	pairing := new(devicePairing)
	if err := s.cache.Get(ctx, "pairing_code/"+req.Code, pairing); err != nil || pairing.PublicKey == "" {
		return nil, NewErrDeviceLoginCodeNotFound(req.Code, err)
	}

	namespace, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID)
	if err != nil {
		return nil, NewErrNamespaceNotFound(req.TenantID, err)
	}

	// The session may be scoped to another tenant, so the gateway's permission
	// middleware cannot cover this route; check the user's role in the chosen
	// namespace explicitly.
	member, ok := namespace.FindMember(userID)
	if !ok {
		return nil, NewErrNamespaceMemberNotFound(userID, nil)
	}

	if !member.Role.HasPermission(authorizer.DeviceAccept) {
		return nil, NewErrRoleInvalid()
	}

	// Materialize the device with the same fields the agent will later send on
	// its own device auth, so the UID hash matches and the agent resolves the
	// accepted device. This mirrors the auto-accept path in AuthDevice: create
	// pending, then accept; "already accepted" is tolerated for idempotency.
	authReq := requests.DeviceAuth{
		Hostname:  pairing.Hostname,
		PublicKey: pairing.PublicKey,
		TenantID:  namespace.TenantID,
	}

	if pairing.Identity != nil {
		authReq.Identity = &requests.DeviceIdentity{MAC: pairing.Identity.MAC}
	}

	if pairing.Info != nil {
		authReq.Info = &requests.DeviceInfo{
			ID:         pairing.Info.ID,
			PrettyName: pairing.Info.PrettyName,
			Version:    pairing.Info.Version,
			Arch:       pairing.Info.Arch,
			Platform:   pairing.Info.Platform,
		}
	}

	auth, err := s.AuthDevice(ctx, authReq)
	if err != nil {
		return nil, err
	}

	accept := &requests.DeviceUpdateStatus{
		TenantID: namespace.TenantID,
		UID:      auth.UID,
		Status:   string(models.DeviceStatusAccepted),
	}
	if err := s.UpdateDeviceStatus(ctx, accept); err != nil && !errors.Is(err, ErrDeviceStatusAccepted) {
		return nil, err
	}

	pairing.Status = models.DeviceStatusAccepted
	pairing.TenantID = namespace.TenantID
	pairing.UID = auth.UID

	// Renew the TTL so the agent has the full window to poll the outcome.
	if err := s.cache.Set(ctx, "pairing_code/"+req.Code, pairing, devicePairingTTL); err != nil {
		log.WithError(err).WithField("device_uid", auth.UID).
			Warn("device accepted but failed to store the pairing outcome; the agent will not learn its tenant from this code")
	}

	return &models.DevicePairingAccepted{
		UID:       auth.UID,
		TenantID:  namespace.TenantID,
		Namespace: namespace.Name,
	}, nil
}

// pairingPreviewName derives the display hostname the same way AuthDevice does
// when materializing the device, so the preview matches the device that will
// be created.
func pairingPreviewName(pairing *devicePairing) string {
	if pairing.Hostname != "" {
		return strings.ToLower(pairing.Hostname)
	}

	if pairing.Identity != nil && pairing.Identity.MAC != "" {
		return strings.ToLower(strings.ReplaceAll(pairing.Identity.MAC, ":", "-"))
	}

	return ""
}
