package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net"
	"strconv"
	"strings"

	utils "github.com/shellhub-io/shellhub/api/pkg/namespace"
	"github.com/shellhub-io/shellhub/api/store"
	req "github.com/shellhub-io/shellhub/pkg/api/client"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/validator"
	"github.com/sirupsen/logrus"
)

type DeviceService interface {
	ListDevices(ctx context.Context, pagination paginator.Query, filter string, status string, sort string, order string) ([]models.Device, int, error)
	GetDevice(ctx context.Context, uid models.UID) (*models.Device, error)
	DeleteDevice(ctx context.Context, uid models.UID, tenant, ownerID string) error
	RenameDevice(ctx context.Context, uid models.UID, name, tenant, ownerID string) error
	LookupDevice(ctx context.Context, namespace, name string) (*models.Device, error)
	UpdateDeviceStatus(ctx context.Context, uid models.UID, online bool) error
	UpdatePendingStatus(ctx context.Context, uid models.UID, status, tenant, ownerID string) error
	HandleReports(ns *models.Namespace, ui models.UID, inc bool, device *models.Device) error
	SetDevicePosition(ctx context.Context, uid models.UID, ip string) error
	CreateTag(ctx context.Context, uid models.UID, name string) error
	DeleteTag(ctx context.Context, uid models.UID, name string) error
	RenameTag(ctx context.Context, uid models.UID, currentName string, newName string) error
	ListTag(ctx context.Context) ([]string, int, error)
	UpdateTag(ctx context.Context, uid models.UID, tags []string) error
}

func (s *service) HandleReports(ns *models.Namespace, uid models.UID, inc bool, device *models.Device) error {
	if ns.Billing == nil || !ns.Billing.Active || ns.MaxDevices != -1 {
		return nil
	}

	record := &models.UsageRecord{
		UUID:      string(uid),
		Inc:       inc,
		Timestamp: clock.Now().Unix(),
		Namespace: ns,
		Created:   strconv.Itoa(int(device.CreatedAt.Unix())),
	}

	status, err := s.client.(req.Client).ReportUsage(
		record,
		"billing-api",
	)
	if err != nil {
		return err
	}

	switch status {
	case 402:
		return nil
	case 200:
		return nil
	case 400:
		return nil
	}

	return ErrReportUsage
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

func (s *service) DeleteDevice(ctx context.Context, uid models.UID, tenant, ownerID string) error {
	if err := utils.IsNamespaceOwner(ctx, s.store, tenant, ownerID); err != nil {
		return ErrUnauthorized
	}

	device, err := s.store.DeviceGetByUID(ctx, uid, tenant)
	if err != nil {
		return err
	}

	ns, err := s.store.NamespaceGet(ctx, tenant)
	if err != nil {
		return err
	}

	if err = s.HandleReports(ns, uid, false, device); err != nil {
		return err
	}

	return s.store.DeviceDelete(ctx, uid)
}

func (s *service) RenameDevice(ctx context.Context, uid models.UID, name, tenant, ownerID string) error {
	if err := utils.IsNamespaceOwner(ctx, s.store, tenant, ownerID); err != nil {
		return ErrUnauthorized
	}

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

func (s *service) UpdatePendingStatus(ctx context.Context, uid models.UID, status, tenant, ownerID string) error {
	if err := utils.IsNamespaceOwner(ctx, s.store, tenant, ownerID); err != nil {
		return ErrUnauthorized
	}

	device, err := s.store.DeviceGetByUID(ctx, uid, tenant)
	if err != nil {
		return err
	}

	if status == "accepted" { //nolint:nestif
		sameMacDev, err := s.store.DeviceGetByMac(ctx, device.Identity.MAC, device.TenantID, "accepted")
		if err != nil && err != store.ErrNoDocuments {
			return err
		}

		if sameMacDev != nil && sameMacDev.UID != device.UID {
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

			if err := s.HandleReports(ns, uid, true, device); err != nil {
				return err
			}

			if ns.MaxDevices > 0 && ns.MaxDevices <= ns.DevicesCount {
				return ErrMaxDeviceCountReached
			}
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

func (s *service) CreateTag(ctx context.Context, uid models.UID, name string) error {
	if err := validateTagName(name); err != nil {
		return err
	}

	device, err := s.store.DeviceGet(ctx, uid)
	if err != nil {
		return err
	}

	if device == nil {
		return ErrDeviceNotFound
	}

	if len(device.Tags) == 5 {
		return ErrMaxTagReached
	}

	if contains(device.Tags, name) {
		return ErrDuplicateTagName
	}

	return s.store.DeviceCreateTag(ctx, uid, name)
}

func (s *service) DeleteTag(ctx context.Context, uid models.UID, name string) error {
	device, err := s.store.DeviceGet(ctx, uid)
	if err != nil {
		return err
	}

	if device == nil {
		return ErrDeviceNotFound
	}

	return s.store.DeviceDeleteTag(ctx, uid, name)
}

func (s *service) RenameTag(ctx context.Context, uid models.UID, currentName string, newName string) error {
	if _, err := validator.ValidateVar(newName, "required,min=3,max=255,alphanum,ascii"); err != nil {
		return ErrInvalidFormat
	}

	device, err := s.store.DeviceGet(ctx, uid)
	if err != nil {
		return err
	}

	if device == nil {
		return ErrDeviceNotFound
	}

	if contains(device.Tags, newName) {
		return ErrDuplicateTagName
	}

	if !contains(device.Tags, currentName) {
		return ErrNotFound
	}

	return s.store.DeviceRenameTag(ctx, uid, currentName, newName)
}

func (s *service) ListTag(ctx context.Context) ([]string, int, error) {
	return s.store.DeviceListTag(ctx)
}

func (s *service) UpdateTag(ctx context.Context, uid models.UID, tags []string) error {
	for _, tag := range tags {
		if err := validateTagName(tag); err != nil {
			return err
		}
	}

	device, err := s.store.DeviceGet(ctx, uid)
	if err != nil {
		return err
	}

	if device == nil {
		return ErrDeviceNotFound
	}

	if len(device.Tags) == 5 {
		return ErrMaxTagReached
	}

	return s.store.DeviceUpdateTag(ctx, uid, tags)
}

func contains(s []string, name string) bool {
	for _, tag := range s {
		if tag == name {
			return true
		}
	}

	return false
}

func validateTagName(tag string) error {
	if _, err := validator.ValidateVar(tag, "required,min=3,max=255,alphanum,ascii,excludes=/@&:"); err != nil {
		return ErrInvalidFormat
	}

	return nil
}
