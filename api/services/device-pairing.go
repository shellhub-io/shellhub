package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/pairingcode"
	log "github.com/sirupsen/logrus"
)

// devicePairingTTL is how long a pairing code remains valid, and also how long
// the accepted result stays available for the agent to poll after the user
// accepts.
const devicePairingTTL = 10 * time.Minute

// devicePairing is the payload cached under `pairing_code/<code>`. It carries
// the identity a tenant-less agent submitted and, once accepted, the outcome.
//
// A code minted by PrepareDevicePairing starts with an empty PublicKey and a
// PreauthTenantID set: it is a pre-authorization for whichever device claims it
// first, at which point the identity fields are filled and the device is
// accepted into PreauthTenantID without a user in the loop.
type devicePairing struct {
	Hostname  string                 `json:"hostname"`
	Identity  *models.DeviceIdentity `json:"identity"`
	Info      *models.DeviceInfo     `json:"info"`
	PublicKey string                 `json:"public_key"`

	Status   models.DeviceStatus `json:"status"`
	TenantID string              `json:"tenant_id"`
	UID      string              `json:"uid"`

	// PreauthTenantID and PreauthBy are set only for codes minted by a logged-in
	// user via PrepareDevicePairing; they carry the namespace the code enrolls
	// into and the user who authorized it.
	PreauthTenantID string `json:"preauth_tenant_id,omitempty"`
	PreauthBy       string `json:"preauth_by,omitempty"`
}

type DevicePairingService interface {
	// PrepareDevicePairing mints a short-lived, single-use pre-authorized pairing
	// code for a namespace. A logged-in member with the device-accept permission
	// calls it (from the Add Device page); the code is then embedded in the
	// install command so the device that claims it is accepted automatically,
	// with no trip through the pending list.
	PrepareDevicePairing(ctx context.Context, userID, tenantID string) (*models.DevicePairing, error)

	// CreateDevicePairing stores the identity payload of a tenant-less agent and
	// returns a short-lived code that deep-links it into the console's accept
	// page. No device exists until a user accepts the pairing into a namespace.
	//
	// When the request carries a pre-authorized code (req.Code), it takes the
	// claim path instead: the device is accepted into the code's namespace
	// straight away and the code is consumed.
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
	// Claim path: the agent was handed a pre-authorized code at install time.
	// Accept it into the pre-authorized namespace right away, no user in the loop.
	if req.Code != "" {
		return s.claimDevicePairing(ctx, req)
	}

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

	code, err := pairingcode.New(pairingcode.DeviceCodeLength)
	if err != nil {
		return nil, err
	}

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

func (s *service) PrepareDevicePairing(ctx context.Context, userID, tenantID string) (*models.DevicePairing, error) {
	namespace, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, tenantID)
	if err != nil {
		return nil, NewErrNamespaceNotFound(tenantID, err)
	}

	// The route middleware already checked the permission for the session's
	// namespace, but re-derive it from membership here so the pre-authorization
	// can never outrank the user who minted it.
	member, ok := namespace.FindMember(userID)
	if !ok {
		return nil, NewErrNamespaceMemberNotFound(userID, nil)
	}

	if !member.Role.HasPermission(authorizer.DeviceAccept) {
		return nil, NewErrRoleForbidden()
	}

	code, err := pairingcode.New(pairingcode.DeviceCodeLength)
	if err != nil {
		return nil, err
	}

	pairing := &devicePairing{
		Status:          models.DeviceStatusPending,
		PreauthTenantID: namespace.TenantID,
		PreauthBy:       userID,
	}

	if err := s.cache.Set(ctx, "pairing_code/"+code, pairing, devicePairingTTL); err != nil {
		return nil, err
	}

	return &models.DevicePairing{
		Code:      code,
		ExpiresIn: int(devicePairingTTL.Seconds()),
		Status:    models.DeviceStatusPending,
	}, nil
}

// claimDevicePairing accepts a device that presented a pre-authorized code. The
// code itself is the authorization (a member with the accept permission minted
// it), so there is no user session to check here.
func (s *service) claimDevicePairing(ctx context.Context, req *requests.DevicePairingCreate) (*models.DevicePairing, error) {
	code := pairingcode.Normalize(req.Code)
	if !pairingcode.IsValid(code, pairingcode.DeviceCodeLength) {
		return nil, NewErrDevicePairingCodeNotFound(code, nil)
	}

	pairing := new(devicePairing)
	// Only codes minted by PrepareDevicePairing (PreauthTenantID set) can be
	// claimed; an agent-minted code has no pre-authorization and must go through
	// a user accept.
	if err := s.cache.Get(ctx, "pairing_code/"+code, pairing); err != nil || pairing.PreauthTenantID == "" {
		return nil, NewErrDevicePairingCodeNotFound(code, err)
	}

	// Single-use: once a device has claimed the code (PublicKey filled), only that
	// same device may re-read the outcome (idempotent retry); any other key is
	// rejected so a leaked code can enroll at most one device.
	if pairing.PublicKey != "" {
		if pairing.PublicKey == req.PublicKey {
			return &models.DevicePairing{Status: pairing.Status, TenantID: pairing.TenantID}, nil
		}

		return nil, NewErrDevicePairingCodeNotFound(code, nil)
	}

	// Reserve the code atomically before materializing anything. The Get/check/Set
	// above is not race-safe on its own: two devices with different keys could both
	// pass the PublicKey=="" gate and both get accepted. SetNX lets exactly one
	// concurrent claim win, keeping the code truly single-use. The reservation is
	// released below if materialization fails, so a legitimate retry can proceed.
	claimRef := "pairing_claim/" + code

	reserved, err := s.cache.SetNX(ctx, claimRef, hashPublicKey(req.PublicKey), devicePairingTTL)
	if err != nil {
		return nil, err
	}

	if !reserved {
		return nil, NewErrDevicePairingCodeNotFound(code, nil)
	}

	namespace, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, pairing.PreauthTenantID)
	if err != nil {
		_ = s.cache.Delete(ctx, claimRef)

		return nil, NewErrNamespaceNotFound(pairing.PreauthTenantID, err)
	}

	// Fill the payload so the cached outcome is complete, then accept via the shared path.
	pairing.Hostname = req.Hostname
	pairing.PublicKey = req.PublicKey

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

	auth, err := s.acceptPairingDevice(ctx, pairing, namespace.TenantID)
	if err != nil {
		// Release the reservation so a retry can proceed. We deliberately do NOT
		// delete the device on failure: AuthDevice is create-or-resolve, so it may
		// have returned a device that already existed (e.g. a pending one blocked
		// by the device limit) — deleting by UID could destroy the user's own
		// device. A leftover pending device is harmless and can be accepted or
		// removed from the console.
		_ = s.cache.Delete(ctx, claimRef)

		return nil, err
	}

	pairing.Status = models.DeviceStatusAccepted
	pairing.TenantID = namespace.TenantID
	pairing.UID = auth.UID

	// Renew the TTL so the console has the full window to poll the outcome.
	if err := s.cache.Set(ctx, "pairing_code/"+code, pairing, devicePairingTTL); err != nil {
		log.WithError(err).WithField("device_uid", auth.UID).
			Warn("device accepted but failed to store the pairing outcome; the console will not see it via this code")
	}

	return &models.DevicePairing{Status: models.DeviceStatusAccepted, TenantID: namespace.TenantID}, nil
}

// hashPublicKey derives a cache-key-safe identifier for a device public key
// (the PEM contains newlines and is long), used to map a key to its live code.
func hashPublicKey(publicKey string) string {
	sum := sha256.Sum256([]byte(publicKey))

	return hex.EncodeToString(sum[:])
}

func (s *service) GetDevicePairingStatus(ctx context.Context, code string) (*models.DevicePairingStatus, error) {
	code = pairingcode.Normalize(code)

	pairing := new(devicePairing)
	// NOTE: A cache miss is not an error; it leaves the value untouched. A code
	// exists if it was minted by an agent (PublicKey set once it submits) or
	// pre-authorized by a user (PreauthTenantID set before any device claims it).
	if err := s.cache.Get(ctx, "pairing_code/"+code, pairing); err != nil ||
		(pairing.PublicKey == "" && pairing.PreauthTenantID == "") {

		return nil, NewErrDevicePairingCodeNotFound(code, err)
	}

	return &models.DevicePairingStatus{
		Status:   pairing.Status,
		TenantID: pairing.TenantID,
		UID:      pairing.UID,
		Name:     pairingPreviewName(pairing),
	}, nil
}

func (s *service) AcceptDevicePairing(ctx context.Context, userID string, req *requests.DevicePairingAccept) (*models.DevicePairingAccepted, error) {
	code := pairingcode.Normalize(req.Code)
	if !pairingcode.IsValid(code, pairingcode.DeviceCodeLength) {
		return nil, NewErrDevicePairingCodeNotFound(code, nil)
	}

	pairing := new(devicePairing)
	if err := s.cache.Get(ctx, "pairing_code/"+code, pairing); err != nil || pairing.PublicKey == "" {
		return nil, NewErrDevicePairingCodeNotFound(code, err)
	}

	// Pre-authorized codes are auto-accept-only: they carry their own namespace
	// and are claimed by the agent, never routed through a user-chosen accept.
	// Refuse them here so a code holder can't re-materialize the device into an
	// arbitrary namespace.
	if pairing.PreauthTenantID != "" {
		return nil, NewErrDevicePairingCodeNotFound(code, nil)
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
		return nil, NewErrRoleForbidden()
	}

	auth, err := s.acceptPairingDevice(ctx, pairing, namespace.TenantID)
	if err != nil {
		return nil, err
	}

	pairing.Status = models.DeviceStatusAccepted
	pairing.TenantID = namespace.TenantID
	pairing.UID = auth.UID

	// Renew the TTL so the agent has the full window to poll the outcome.
	if err := s.cache.Set(ctx, "pairing_code/"+code, pairing, devicePairingTTL); err != nil {
		log.WithError(err).WithField("device_uid", auth.UID).
			Warn("device accepted but failed to store the pairing outcome; the agent will not learn its tenant from this code")
	}

	return &models.DevicePairingAccepted{
		UID:       auth.UID,
		TenantID:  namespace.TenantID,
		Namespace: namespace.Name,
	}, nil
}

// acceptPairingDevice materializes the pairing payload as a device and accepts it.
// The device auth uses the same fields the agent later sends on its own, so the UID
// hash matches; an already-accepted device is tolerated for idempotency.
func (s *service) acceptPairingDevice(ctx context.Context, pairing *devicePairing, tenantID string) (*models.DeviceAuthResponse, error) {
	authReq := requests.DeviceAuth{
		Hostname:  pairing.Hostname,
		PublicKey: pairing.PublicKey,
		TenantID:  tenantID,
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

	auth, err := s.authDevice(ctx, authReq, true)
	if err != nil {
		return nil, err
	}

	accept := &requests.DeviceUpdateStatus{
		TenantID: tenantID,
		UID:      auth.UID,
		Status:   string(models.DeviceStatusAccepted),
	}
	if err := s.UpdateDeviceStatus(ctx, accept); err != nil && !errors.Is(err, ErrDeviceStatusAccepted) {
		return nil, err
	}

	return auth, nil
}

// pairingPreviewName derives the display hostname the same way AuthDevice does
// when materializing the device, so the preview matches the device that will
// be created.
func pairingPreviewName(pairing *devicePairing) string {
	var mac string
	if pairing.Identity != nil {
		mac = pairing.Identity.MAC
	}

	return strings.ToLower(deviceHostname(pairing.Hostname, mac))
}
