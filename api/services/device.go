package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
)

const StatusAccepted = "accepted"

type DeviceService interface {
	ListDevices(ctx context.Context, req *requests.DeviceList) ([]models.Device, int, error)
	GetDevice(ctx context.Context, uid models.UID) (*models.Device, error)

	// ResolveDevice attempts to resolve a device by searching for either its UID or hostname. When both are provided,
	// UID takes precedence over hostname. The search is scoped to the namespace's tenant ID to limit results.
	//
	// It returns the resolved device and any error encountered.
	ResolveDevice(ctx context.Context, req *requests.ResolveDevice) (*models.Device, error)

	DeleteDevice(ctx context.Context, uid models.UID, tenant string) error
	RenameDevice(ctx context.Context, uid models.UID, name, tenant string) error
	LookupDevice(ctx context.Context, namespace, name string) (*models.Device, error)
	OfflineDevice(ctx context.Context, uid models.UID) error
	UpdateDeviceStatus(ctx context.Context, tenant string, uid models.UID, status models.DeviceStatus) error

	UpdateDevice(ctx context.Context, req *requests.DeviceUpdate) error
}

func (s *service) ListDevices(ctx context.Context, req *requests.DeviceList) ([]models.Device, int, error) {
	if req.DeviceStatus == models.DeviceStatusRemoved {
		// TODO: unique DeviceList
		removed, count, err := s.store.DeviceRemovedList(ctx, req.TenantID, req.Paginator, req.Filters, req.Sorter)
		if err != nil {
			return nil, 0, err
		}

		devices := make([]models.Device, 0, len(removed))
		for _, device := range removed {
			devices = append(devices, *device.Device)
		}

		return devices, count, nil
	}

	if req.TenantID != "" {
		ns, err := s.store.NamespaceGet(ctx, req.TenantID)
		if err != nil {
			return nil, 0, NewErrNamespaceNotFound(req.TenantID, err)
		}

		if ns.HasMaxDevices() {
			switch {
			case envs.IsCloud():
				removed, err := s.store.DeviceRemovedCount(ctx, ns.TenantID)
				if err != nil {
					return nil, 0, NewErrDeviceRemovedCount(err)
				}

				if ns.HasLimitDevicesReached(removed) {
					return s.store.DeviceList(ctx, req.DeviceStatus, req.Paginator, req.Filters, req.Sorter, store.DeviceAcceptableFromRemoved)
				}
			case envs.IsEnterprise():
				fallthrough
			case envs.IsCommunity():
				if ns.HasMaxDevicesReached() {
					return s.store.DeviceList(ctx, req.DeviceStatus, req.Paginator, req.Filters, req.Sorter, store.DeviceAcceptableAsFalse)
				}
			}
		}
	}

	return s.store.DeviceList(ctx, req.DeviceStatus, req.Paginator, req.Filters, req.Sorter, store.DeviceAcceptableIfNotAccepted)
}

func (s *service) GetDevice(ctx context.Context, uid models.UID) (*models.Device, error) {
	device, err := s.store.DeviceResolve(ctx, store.DeviceUIDResolver, string(uid))
	if err != nil {
		return nil, NewErrDeviceNotFound(uid, err)
	}

	return device, nil
}

func (s *service) ResolveDevice(ctx context.Context, req *requests.ResolveDevice) (*models.Device, error) {
	n, err := s.store.NamespaceGet(ctx, req.TenantID)
	if err != nil {
		return nil, NewErrNamespaceNotFound(req.TenantID, err)
	}

	var device *models.Device
	switch {
	case req.UID != "":
		device, err = s.store.DeviceResolve(ctx, store.DeviceUIDResolver, req.UID, s.store.Options().InNamespace(n.TenantID))
	case req.Hostname != "":
		device, err = s.store.DeviceResolve(ctx, store.DeviceHostnameResolver, req.Hostname, s.store.Options().InNamespace(n.TenantID))
	}

	if err != nil {
		// TODO: refactor this error to accept a string instead of models.UID
		return nil, NewErrDeviceNotFound(models.UID(""), err)
	}

	return device, nil
}

// DeleteDevice deletes a device from a namespace.
//
// It receives a context, used to "control" the request flow and, the device UID from models.Device and the tenant ID
// from models.Namespace.
//
// It can return an error if the device is not found, NewErrDeviceNotFound(uid, err), if the namespace is not found,
// NewErrNamespaceNotFound(tenant, err), if the usage cannot be reported, ErrReport or if the store function that
// delete the device fails.
func (s *service) DeleteDevice(ctx context.Context, uid models.UID, tenant string) error {
	device, err := s.store.DeviceResolve(ctx, store.DeviceUIDResolver, string(uid), s.store.Options().InNamespace(tenant))
	if err != nil {
		return NewErrDeviceNotFound(uid, err)
	}

	ns, err := s.store.NamespaceGet(ctx, tenant)
	if err != nil {
		return NewErrNamespaceNotFound(tenant, err)
	}

	// If the namespace has a limit of devices, we change the device's slot status to removed.
	// This way, we can keep track of the number of devices that were removed from the namespace and void the device
	// switching.
	if envs.IsCloud() && envs.HasBilling() && !ns.Billing.IsActive() {
		if err := s.store.DeviceRemovedInsert(ctx, tenant, device); err != nil {
			return NewErrDeviceRemovedInsert(err)
		}
	}

	if err := s.store.DeviceDelete(ctx, uid); err != nil {
		return err
	}

	if err := s.store.NamespaceIncrementDeviceCount(ctx, tenant, device.Status, -1); err != nil { //nolint:revive
		return err
	}

	return nil
}

func (s *service) RenameDevice(ctx context.Context, uid models.UID, name, tenant string) error {
	device, err := s.store.DeviceResolve(ctx, store.DeviceUIDResolver, string(uid), s.store.Options().InNamespace(tenant))
	if err != nil {
		return NewErrDeviceNotFound(uid, err)
	}

	updatedDevice := &models.Device{
		UID:        device.UID,
		Name:       strings.ToLower(name),
		Identity:   device.Identity,
		Info:       device.Info,
		PublicKey:  device.PublicKey,
		TenantID:   device.TenantID,
		LastSeen:   device.LastSeen,
		Online:     device.Online,
		Namespace:  device.Namespace,
		Status:     device.Status,
		CreatedAt:  time.Time{},
		RemoteAddr: "",
		Position:   &models.DevicePosition{},
		Tags:       []string{},
	}

	if ok, err := s.validator.Struct(updatedDevice); !ok || err != nil {
		return NewErrDeviceInvalid(nil, err)
	}

	if device.Name == updatedDevice.Name {
		return nil
	}

	otherDevice, err := s.store.DeviceResolve(ctx, store.DeviceHostnameResolver, updatedDevice.Name, s.store.Options().WithDeviceStatus(models.DeviceStatusAccepted), s.store.Options().InNamespace(tenant))
	if err != nil && err != store.ErrNoDocuments {
		return NewErrDeviceNotFound(models.UID(updatedDevice.UID), err)
	}

	if otherDevice != nil {
		return NewErrDeviceDuplicated(otherDevice.Name, err)
	}

	return s.store.DeviceRename(ctx, uid, name)
}

// LookupDevice looks for a device in a namespace.
//
// It receives a context, used to "control" the request flow and, the namespace name from a models.Namespace and a
// device name from models.Device.
func (s *service) LookupDevice(ctx context.Context, namespace, name string) (*models.Device, error) {
	n, err := s.store.NamespaceGetByName(ctx, namespace)
	if err != nil {
		return nil, NewErrNamespaceNotFound(namespace, err)
	}

	device, err := s.store.DeviceResolve(ctx, store.DeviceHostnameResolver, name, s.store.Options().InNamespace(n.TenantID))
	if err != nil || device == nil {
		return nil, NewErrDeviceNotFound(models.UID(name), err)
	}

	return device, nil
}

func (s *service) OfflineDevice(ctx context.Context, uid models.UID) error {
	now := clock.Now()
	if err := s.store.DeviceUpdate(ctx, "", string(uid), &models.DeviceChanges{DisconnectedAt: &now}); err != nil {
		if errors.Is(err, store.ErrNoDocuments) {
			return NewErrDeviceNotFound(uid, err)
		}

		return err
	}

	return nil
}

// UpdateDeviceStatus updates the device status.
func (s *service) UpdateDeviceStatus(ctx context.Context, req *requests.UpdateDeviceStatus) error {
	namespace, err := s.store.NamespaceGet(ctx, req.TenantID)
	if err != nil {
		return NewErrNamespaceNotFound(req.TenantID, err)
	}

	device, err := s.store.DeviceResolve(ctx, store.DeviceUIDResolver, req.UID, s.store.Options().InNamespace(req.TenantID))
	if err != nil {
		return NewErrDeviceNotFound(models.UID(req.UID), err)
	}

	if device.Status == models.DeviceStatusAccepted {
		return NewErrDeviceStatusAccepted(nil)
	}

	oldStatus := device.Status
	newStatus := models.DeviceStatus(req.Status)
	opts := []store.QueryOption{s.store.Options().WithDeviceStatus(models.DeviceStatusAccepted), s.store.Options().InNamespace(device.TenantID)}

	// Busca se já existe um device aceito com o mesmo MAC address
	// Se existir, precisaremos fazer merge: transferir sessões, deletar o antigo, renomear o novo
	existingMacDevice, _ := s.store.DeviceResolve(ctx, store.DeviceMACResolver, device.Identity.MAC, opts...)
	if existingMacDevice != nil && existingMacDevice.UID != device.UID {
		// ANTES do merge, verifica se vai dar conflito de hostname
		// Busca se já existe outro device aceito usando o hostname atual
		hostnameConflictDevice, _ := s.store.DeviceResolve(ctx, store.DeviceHostnameResolver, device.Name, opts...)
		if hostnameConflictDevice != nil && hostnameConflictDevice.Identity.MAC != device.Identity.MAC {
			return NewErrDeviceDuplicated(device.Name, nil)
		}

		// if err := s.mergeDevice(ctx, existingMacDevice, device); err != nil {
		// 	return err
		// }

		// OK para fazer merge! Transfere sessões do device antigo para o novo
		_ = s.store.SessionUpdateDeviceUID(ctx, models.UID(existingMacDevice.UID), models.UID(device.UID))
		// Renomeia o device novo para usar o nome do device antigo (preserva identidade)
		_ = s.store.DeviceRename(ctx, models.UID(device.UID), existingMacDevice.Name) // TODO: DeviceUpdate
		// Deleta o device antigo (já transferiu sessões e nome)
		_ = s.store.DeviceDelete(ctx, models.UID(existingMacDevice.UID)) // TODO: mergear tunnels?
		// Decrementa contador porque deletamos um device aceito
		_ = s.store.NamespaceIncrementDeviceCount(ctx, req.TenantID, models.DeviceStatusAccepted, -1)
	} else {
		// Pq aqui ele n verifica o MAC para saber se eh igual ou nao?
		sameDevice, _ := s.store.DeviceResolve(ctx, store.DeviceHostnameResolver, device.Name, opts...)
		if sameDevice != nil {
			return NewErrDeviceDuplicated(device.Name, nil)
		}

		switch {
		case envs.IsCommunity(), envs.IsEnterprise():
			// Versões Community/Enterprise: só verifica limite simples de devices
			// Mas isso faz sentido? Community n tem limite
			if namespace.HasMaxDevices() && namespace.HasMaxDevicesReached() {
				return NewErrDeviceMaxDevicesReached(namespace.MaxDevices)
			}
		case envs.IsCloud():
			// Versão Cloud: verifica billing e limites mais complexos
			if namespace.Billing.IsActive() {
				// Billing ativo: reporta o device aceito para cobrança
				if err := s.BillingReport(s.client, namespace.TenantID, ReportDeviceAccept); err != nil {
					return NewErrBillingReportNamespaceDelete(err)
				}
			} else {
				// Billing inativo: verifica sistema de "devices removidos" para limite
				removed, err := s.store.DeviceRemovedGet(ctx, req.TenantID, models.UID(req.UID))
				if err != nil && err != store.ErrNoDocuments {
					return NewErrDeviceRemovedGet(err)
				}

				if removed != nil {
					// Se device estava na lista de "removidos", remove da lista (pode aceitar de novo)
					if err := s.store.DeviceRemovedDelete(ctx, req.TenantID, models.UID(req.UID)); err != nil {
						return NewErrDeviceRemovedDelete(err)
					}
				} else {
					// Device novo: verifica se ainda tem "slot" disponível baseado em devices removidos
					count, err := s.store.DeviceRemovedCount(ctx, req.TenantID)
					if err != nil {
						return NewErrDeviceRemovedCount(err)
					}

					if namespace.HasMaxDevices() && namespace.HasLimitDevicesReached(count) {
						return NewErrDeviceRemovedFull(namespace.MaxDevices, nil)
					}
				}

				// Verifica com sistema de billing se pode aceitar mais devices
				ok, err := s.BillingEvaluate(s.client, namespace.TenantID)
				if err != nil {
					return NewErrBillingEvaluate(err)
				}

				if !ok {
					return ErrDeviceLimit
				}
			}
		}
	}

	if err := s.store.DeviceUpdateStatus(ctx, models.UID(req.UID), newStatus); err != nil { // TODO: DeviceUpdate
		return err
	}

	if err := s.adjustDeviceCounters(ctx, req.TenantID, oldStatus, newStatus); err != nil { // nolint:revive
		return err
	}

	return nil
}

func (s *service) UpdateDevice(ctx context.Context, req *requests.DeviceUpdate) error {
	device, err := s.store.DeviceResolve(ctx, store.DeviceUIDResolver, req.UID, s.store.Options().InNamespace(req.TenantID))
	if err != nil {
		return NewErrDeviceNotFound(models.UID(req.UID), err)
	}

	conflictsTarget := &models.DeviceConflicts{Name: req.Name}
	conflictsTarget.Distinct(device)
	if _, has, err := s.store.DeviceConflicts(ctx, conflictsTarget); err != nil || has {
		return NewErrDeviceDuplicated(req.Name, err)
	}

	// We pass DisconnectedAt because we don't want to update it to nil
	changes := &models.DeviceChanges{DisconnectedAt: device.DisconnectedAt}
	if req.Name != "" && strings.ToLower(req.Name) != device.Name {
		changes.Name = strings.ToLower(req.Name)
	}

	return s.store.DeviceUpdate(ctx, req.TenantID, req.UID, changes)
}
