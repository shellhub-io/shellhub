package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/pairingcode"
	"github.com/shellhub-io/shellhub/server/api/store"
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
	if req.Code != "" {
		return s.claimDevicePairing(ctx, req)
	}

	sc := scope.NewUnbounded("pairing by public key: the device has not been placed in a namespace yet, and possession of the matching private key is still required")
	if device, err := s.store.DeviceResolve(ctx, sc, store.DevicePublicKeyResolver, req.PublicKey, s.store.Options().WithDeviceStatus(models.DeviceStatusAccepted)); err == nil && device != nil {
		return &models.DevicePairing{Status: models.DeviceStatusAccepted, TenantID: device.TenantID}, nil
	}

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
	if err := s.cache.Get(ctx, "pairing_code/"+code, pairing); err != nil || pairing.PreauthTenantID == "" {
		return nil, NewErrDevicePairingCodeNotFound(code, err)
	}

	if pairing.PublicKey != "" {
		if pairing.PublicKey == req.PublicKey {
			return &models.DevicePairing{Status: pairing.Status, TenantID: pairing.TenantID}, nil
		}

		return nil, NewErrDevicePairingCodeNotFound(code, nil)
	}

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
		_ = s.cache.Delete(ctx, claimRef)

		return nil, err
	}

	pairing.Status = models.DeviceStatusAccepted
	pairing.TenantID = namespace.TenantID
	pairing.UID = auth.UID

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

	if pairing.PreauthTenantID != "" {
		return nil, NewErrDevicePairingCodeNotFound(code, nil)
	}

	namespace, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID)
	if err != nil {
		return nil, NewErrNamespaceNotFound(req.TenantID, err)
	}

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
