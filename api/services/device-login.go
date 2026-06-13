package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
)

// deviceLoginCodeTTL is how long a device login code remains valid. The code
// only deep-links a pending device into the console's accept page, so a short
// window is enough; the agent prints a new one on every `login` run.
const deviceLoginCodeTTL = 10 * time.Minute

// deviceLoginCode is the payload cached under `login_code/<code>`.
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
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}

	code := hex.EncodeToString(buf)

	// Invalidate the previous code issued to this device, if any, so issuing new
	// codes never accumulates multiple valid ones.
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
	data := new(deviceLoginCode)
	// NOTE: A cache miss is not an error; it leaves data untouched.
	if err := s.cache.Get(ctx, "login_code/"+code, data); err != nil || data.UID == "" {
		// Not a device-bound code; it may be a pairing code from a tenant-less
		// agent. The device does not exist yet, so there is no membership to
		// check: any authenticated user may see the preview, but accepting is
		// limited to namespaces where they hold the device accept permission.
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

	device, err := s.store.DeviceResolve(ctx, store.DeviceUIDResolver, data.UID, s.store.Options().InNamespace(data.TenantID))
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
