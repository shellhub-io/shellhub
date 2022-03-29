package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net"
	"strings"

	"github.com/shellhub-io/shellhub/api/store"
	req "github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	hp "github.com/shellhub-io/shellhub/pkg/requests"
	"github.com/shellhub-io/shellhub/pkg/validator"
	"github.com/sirupsen/logrus"
)

const StatusAccepted = "accepted"

type DeviceService interface {
	ListDevices(ctx context.Context, pagination paginator.Query, filter string, status string, sort string, order string) ([]models.Device, int, error)
	GetDevice(ctx context.Context, uid models.UID) (*models.Device, error)
	DeleteDevice(ctx context.Context, uid models.UID, tenant string) error
	RenameDevice(ctx context.Context, uid models.UID, name, tenant string) error
	LookupDevice(ctx context.Context, namespace, name string) (*models.Device, error)
	UpdateDeviceStatus(ctx context.Context, uid models.UID, online bool) error
	UpdatePendingStatus(ctx context.Context, uid models.UID, status, tenant string) error
	HandleReportUsage(ns *models.Namespace, ui models.UID, inc bool, device *models.Device) error
	SetDevicePosition(ctx context.Context, uid models.UID, ip string) error
	DeviceHeartbeat(ctx context.Context, uid models.UID) error
}

func (s *service) HandleReportUsage(ns *models.Namespace, uid models.UID, inc bool, device *models.Device) error {
	if !hp.HasBillingInstance(ns) {
		return nil
	}

	record := &models.UsageRecord{
		Device:    device,
		Inc:       inc,
		Timestamp: clock.Now().Unix(),
		Namespace: ns,
	}

	status, err := s.client.(req.Client).ReportUsage(
		record,
	)
	if err != nil {
		return err
	}

	return handleStatusResponse(status)
}

func (s *service) ListDevices(ctx context.Context, pagination paginator.Query, filterB64 string, status string, sort string, order string) ([]models.Device, int, error) {
	raw, err := base64.StdEncoding.DecodeString(filterB64)
	if err != nil {
		return nil, 0, err
	}

	var filter []models.Filter
	if err := json.Unmarshal(raw, &filter); len(raw) > 0 && err != nil {
		return nil, 0, err
	}

	return s.store.DeviceList(ctx, pagination, filter, status, sort, order)
}

func (s *service) GetDevice(ctx context.Context, uid models.UID) (*models.Device, error) {
	return s.store.DeviceGet(ctx, uid)
}

func (s *service) DeleteDevice(ctx context.Context, uid models.UID, tenant string) error {
	device, err := s.store.DeviceGetByUID(ctx, uid, tenant)
	if err != nil {
		return err
	}

	ns, err := s.store.NamespaceGet(ctx, tenant)
	if err != nil {
		return err
	}

	if err = s.HandleReportUsage(ns, uid, false, device); err != nil {
		return err
	}

	return s.store.DeviceDelete(ctx, uid)
}

func (s *service) RenameDevice(ctx context.Context, uid models.UID, name, tenant string) error {
	device, err := s.store.DeviceGetByUID(ctx, uid, tenant)
	if err != nil {
		return err
	}

	updatedDevice := &models.Device{
		UID:       device.UID,
		Name:      strings.ToLower(name),
		Identity:  device.Identity,
		Info:      device.Info,
		PublicKey: device.PublicKey,
		TenantID:  device.TenantID,
		LastSeen:  device.LastSeen,
		Online:    device.Online,
		Namespace: device.Namespace,
		Status:    device.Status,
	}

	if _, err = validator.ValidateStruct(updatedDevice); err != nil {
		return ErrInvalidFormat
	}

	if device.Name == updatedDevice.Name {
		return nil
	}

	otherDevice, err := s.store.DeviceGetByName(ctx, updatedDevice.Name, tenant)
	if err != nil && err != store.ErrNoDocuments {
		return err
	}

	if otherDevice != nil {
		return ErrDuplicatedDeviceName
	}

	return s.store.DeviceRename(ctx, uid, name)
}

func (s *service) LookupDevice(ctx context.Context, namespace, name string) (*models.Device, error) {
	device, err := s.store.DeviceLookup(ctx, namespace, name)
	if err != nil && err == store.ErrNoDocuments {
		return nil, ErrNotFound
	}

	return device, err
}

func (s *service) UpdateDeviceStatus(ctx context.Context, uid models.UID, online bool) error {
	return s.store.DeviceSetOnline(ctx, uid, online)
}

func (s *service) UpdatePendingStatus(ctx context.Context, uid models.UID, status, tenant string) error {
	validateStatus := map[string]bool{
		"accepted": true,
		"pending":  true,
		"rejected": true,
	}

	if _, ok := validateStatus[status]; !ok {
		return ErrBadRequest
	}

	device, err := s.store.DeviceGetByUID(ctx, uid, tenant)
	if err != nil {
		return err
	}

	if device.Status == StatusAccepted {
		return ErrForbidden
	}

	if status != StatusAccepted {
		return s.store.DeviceUpdateStatus(ctx, uid, status)
	}

	sameMacDev, err := s.store.DeviceGetByMac(ctx, device.Identity.MAC, device.TenantID, "accepted")
	if err != nil && err != store.ErrNoDocuments {
		return err
	}

	if sameMacDev != nil && sameMacDev.UID != device.UID { //nolint:nestif
		if err := s.store.SessionUpdateDeviceUID(ctx, models.UID(sameMacDev.UID), models.UID(device.UID)); err != nil {
			return err
		}
		if err := s.store.DeviceDelete(ctx, models.UID(sameMacDev.UID)); err != nil {
			return err
		}
		if err := s.store.DeviceRename(ctx, models.UID(device.UID), sameMacDev.Name); err != nil {
			return err
		}
	} else {
		ns, err := s.store.NamespaceGet(ctx, device.TenantID)
		if err != nil {
			return err
		}

		if err := s.HandleReportUsage(ns, uid, true, device); err != nil {
			return err
		}

		if ns.MaxDevices > 0 && ns.MaxDevices <= ns.DevicesCount {
			return ErrMaxDeviceCountReached
		}
	}

	return s.store.DeviceUpdateStatus(ctx, uid, status)
}

func (s *service) SetDevicePosition(ctx context.Context, uid models.UID, ip string) error {
	ipParsed := net.ParseIP(ip)
	position, err := s.locator.GetPosition(ipParsed)
	if err != nil {
		logrus.
			WithError(err).
			WithFields(logrus.Fields{
				"uid": uid,
				"ip":  ip,
			}).Error("Failed to get device's position")
	}

	devicePosition := models.DevicePosition{
		Longitude: position.Longitude,
		Latitude:  position.Latitude,
	}

	err = s.store.DeviceSetPosition(ctx, uid, devicePosition)
	if err != nil {
		logrus.
			WithError(err).
			WithFields(logrus.Fields{
				"uid": uid,
				"ip":  ip,
			}).Error("Failed to set device's position to database")

		return err
	}
	logrus.WithFields(logrus.Fields{
		"uid":      uid,
		"ip":       ip,
		"position": position,
	}).Debug("Success to set device's position")

	return nil
}

func (s *service) DeviceHeartbeat(ctx context.Context, uid models.UID) error {
	return s.store.DeviceSetOnline(ctx, uid, true)
}
