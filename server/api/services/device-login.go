package services

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/pairingcode"
	"github.com/shellhub-io/shellhub/server/api/store"
	log "github.com/sirupsen/logrus"
)

const deviceLoginCodeTTL = 10 * time.Minute

type deviceLoginCode struct {
	UID      string `json:"uid"`
	TenantID string `json:"tenant_id"`
}

type DeviceLoginCodeService interface {
	// CreateDeviceLoginCode generates a short-lived code that the agent prints as an
	// accept-device URL. The code maps back to the device's UID and tenant; any
	// previous code issued to the same device is invalidated so a device has at most
	// one valid code at a time.
	CreateDeviceLoginCode(ctx context.Context, uid, tenantID string) (*models.DeviceLoginCode, error)

	// ResolveDeviceLoginCode resolves a login code into a device preview for the
	// accept page. It requires the requesting user to be a member of the device's
	// namespace; unknown, expired, and non-member codes all return the same
	// not-found error to avoid leaking the existence of a code.
	ResolveDeviceLoginCode(ctx context.Context, userID, code string) (*models.DeviceLoginCodePreview, error)
}

func (s *service) CreateDeviceLoginCode(ctx context.Context, uid, tenantID string) (*models.DeviceLoginCode, error) {
	code, err := pairingcode.New(pairingcode.DeviceCodeLength)
	if err != nil {
		return nil, err
	}

	var previous string
	if err := s.cache.Get(ctx, "login_code_device/"+uid, &previous); err == nil && previous != "" {
		if err := s.cache.Delete(ctx, "login_code/"+previous); err != nil {
			log.WithError(err).WithField("device_uid", uid).Warn("failed to invalidate previous device login code")
		}
	}

	if err := s.cache.Set(ctx, "login_code/"+code, &deviceLoginCode{UID: uid, TenantID: tenantID}, deviceLoginCodeTTL); err != nil {
		return nil, err
	}

	if err := s.cache.Set(ctx, "login_code_device/"+uid, code, deviceLoginCodeTTL); err != nil {
		return nil, err
	}

	return &models.DeviceLoginCode{Code: code, ExpiresIn: int(deviceLoginCodeTTL.Seconds())}, nil
}

func (s *service) ResolveDeviceLoginCode(ctx context.Context, userID, code string) (*models.DeviceLoginCodePreview, error) {
	code = pairingcode.Normalize(code)
	if !pairingcode.IsValid(code, pairingcode.DeviceCodeLength) {
		return nil, NewErrDeviceLoginCodeNotFound(code, nil)
	}

	data := new(deviceLoginCode)
	if err := s.cache.Get(ctx, "login_code/"+code, data); err != nil || data.UID == "" {
		pairing := new(devicePairing)
		if err := s.cache.Get(ctx, "pairing_code/"+code, pairing); err == nil && pairing.PublicKey != "" {
			return &models.DeviceLoginCodePreview{
				Kind:     models.DeviceLoginCodeKindPairing,
				Name:     pairingPreviewName(pairing),
				Identity: pairing.Identity,
				Info:     pairing.Info,
			}, nil
		}

		return nil, NewErrDeviceLoginCodeNotFound(code, err)
	}

	namespace, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, data.TenantID)
	if err != nil {
		return nil, NewErrDeviceLoginCodeNotFound(code, err)
	}

	if _, ok := namespace.FindMember(userID); !ok {
		return nil, NewErrDeviceLoginCodeNotFound(code, nil)
	}

	device, err := s.store.DeviceResolve(ctx, scope.MustBounded(namespace.TenantID), store.DeviceUIDResolver, data.UID)
	if err != nil {
		return nil, NewErrDeviceLoginCodeNotFound(code, err)
	}

	return &models.DeviceLoginCodePreview{
		Kind:      models.DeviceLoginCodeKindDevice,
		UID:       device.UID,
		Name:      device.Name,
		Identity:  device.Identity,
		Info:      device.Info,
		Namespace: namespace.Name,
		TenantID:  device.TenantID,
		Status:    device.Status,
	}, nil
}
