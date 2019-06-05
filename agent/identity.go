package main

import (
	"errors"
	"net"
	"sort"
)

var ErrNoInterfaceFound = errors.New("No interface found")

type DeviceIdentity struct {
	MAC string `json:"mac"`
}

func GetDeviceIdentity() (*DeviceIdentity, error) {
	d := &DeviceIdentity{}

	iface, err := primaryIface()
	if err != nil {
		return nil, err
	}

	d.MAC = iface.HardwareAddr.String()

	return d, nil
}

func primaryIface() (*net.Interface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	indexes := []int{}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback > 0 {
			continue
		}

		indexes = append(indexes, iface.Index)
	}

	if len(indexes) == 0 {
		return nil, ErrNoInterfaceFound
	}

	sort.Ints(indexes)

	iface, err := net.InterfaceByIndex(indexes[0])
	if err != nil {
		return nil, err
	}

	return iface, nil
}
