package main

import (
	"github.com/shellhub-io/shellhub/agent/pkg/sysinfo"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type Device struct {
	Identity *models.DeviceIdentity
	Info     *models.DeviceInfo
}

func NewDevice() (*Device, error) {
	identity, err := getDeviceIdentity()
	if err != nil {
		return nil, err
	}

	info, err := getDeviceInfo()
	if err != nil {
		return nil, err
	}

	return &Device{identity, info}, nil
}

func getDeviceIdentity() (*models.DeviceIdentity, error) {
	d := &models.DeviceIdentity{}

	iface, err := sysinfo.PrimaryInterface()
	if err != nil {
		return nil, err
	}

	d.MAC = iface.HardwareAddr.String()

	return d, nil
}

func getDeviceInfo() (*models.DeviceInfo, error) {
	osrelease, err := sysinfo.GetOSRelease()
	if err != nil {
		return nil, err
	}

	return &models.DeviceInfo{ID: osrelease.ID, PrettyName: osrelease.Name}, nil
}
