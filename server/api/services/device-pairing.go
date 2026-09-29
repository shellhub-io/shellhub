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

const devicePairingTTL = 10 * time.Minute

type devicePairing struct {
	Hostname  string                 `json:"hostname"`
	Identity  *models.DeviceIdentity `json:"identity"`
	Info      *models.DeviceInfo     `json:"info"`
	PublicKey string                 `json:"public_key"`

	Status   models.DeviceStatus `json:"status"`
	TenantID string              `json:"tenant_id"`
	UID      string              `json:"uid"`
}

// DevicePairingService issues pairing codes to tenant-less agents and lets a logged-in member
// accept the paired device into a namespace.
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
	//
	// The response carries the surviving device's owner, which is empty only when
	// the device merged into a team device. It returns an error when the accepted
	// device cannot be read back, rather than a response that would read as one.
	AcceptDevicePairing(ctx context.Context, userID string, req *requests.DevicePairingAccept) (*models.DevicePairingAccepted, error)
}

func (s *service) CreateDevicePairing(ctx context.Context, req *requests.DevicePairingCreate) (*models.DevicePairing, error) {
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

func hashPublicKey(publicKey string) string {
	sum := sha256.Sum256([]byte(publicKey))

	return hex.EncodeToString(sum[:])
}

func (s *service) GetDevicePairingStatus(ctx context.Context, code string) (*models.DevicePairingStatus, error) {
	code = pairingcode.Normalize(code)

	pairing := new(devicePairing)
	if err := s.cache.Get(ctx, "pairing_code/"+code, pairing); err != nil || pairing.PublicKey == "" {
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

	auth, err := s.acceptPairingDevice(ctx, pairing, namespace.TenantID, userID)
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

	accepted := &models.DevicePairingAccepted{
		UID:       auth.UID,
		TenantID:  namespace.TenantID,
		Namespace: namespace.Name,
	}

	device, err := s.store.DeviceResolve(ctx, scope.MustBounded(namespace.TenantID), store.DeviceUIDResolver, auth.UID)
	if err != nil {
		return nil, err
	}

	accepted.OwnerID = device.OwnerID

	return accepted, nil
}

func (s *service) acceptPairingDevice(ctx context.Context, pairing *devicePairing, tenantID, ownerID string) (*models.DeviceAuthResponse, error) {
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

	auth, err := s.authDevice(ctx, authReq, enrollmentOptions{paired: true, ownerID: ownerID})
	if err != nil {
		return nil, err
	}

	accept := &requests.DeviceUpdateStatus{
		TenantID: tenantID,
		UID:      auth.UID,
		Status:   string(models.DeviceStatusAccepted),
	}
	if err := s.updateDeviceStatusOwnedBy(ctx, accept, ownerID); err != nil && !errors.Is(err, ErrDeviceStatusAccepted) {
		return nil, err
	}

	return auth, nil
}

func pairingPreviewName(pairing *devicePairing) string {
	var mac string
	if pairing.Identity != nil {
		mac = pairing.Identity.MAC
	}

	return strings.ToLower(deviceHostname(pairing.Hostname, mac))
}
