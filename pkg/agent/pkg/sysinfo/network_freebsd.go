//go:build freebsd
// +build freebsd

package sysinfo

import (
	"net"
)

func PrimaryInterface() (*net.Interface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, ErrNoInterfaceFound
	}

	var ifdev *net.Interface

	for i, iface := range interfaces {
		if iface.Flags&net.FlagLoopback > 0 {
			continue
		}

		if iface.Flags&net.FlagRunning > 0 {
			ifdev = &interfaces[i]

			break
		}
	}

	if ifdev == nil {
		return nil, ErrNoInterfaceFound
	}

	return ifdev, nil
}
