package main

import (
	"errors"
	"io/ioutil"
	"math"
	"net"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/shellhub-io/shellhub/agent/internal/osrelease"
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

func GetDeviceInfo() (*DeviceInfo, error) {
	attr := &DeviceInfo{}

	id, err := osrelease.GetValue("ID")
	if err != nil {
		return nil, err
	}

	name, err := osrelease.GetValue("PRETTY_NAME")
	if err != nil {
		return nil, err
	}

	attr.ID = id
	attr.PrettyName = name

	return attr, nil
}

func primaryIface() (*net.Interface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, ErrNoInterfaceFound
	}

	var ifdev *net.Interface
	min := uint64(math.MaxUint16)

	for i, iface := range interfaces {
		if iface.Flags&net.FlagLoopback > 0 {
			continue
		}

		data, err := readSysFs(iface.Name, "type")
		if err != nil {
			break
		}

		iftype, err := strconv.ParseUint(data, 10, 16)
		if err != nil {
			break
		}

		if iftype != syscall.ARPHRD_ETHER {
			break
		}

		data, err = readSysFs(iface.Name, "ifindex")
		if err != nil {
			break
		}

		ifindex, err := strconv.ParseUint(data, 10, 16)
		if err != nil {
			break
		}

		if ifindex < min {
			min = ifindex
			ifdev = &interfaces[i]
		}
	}

	if ifdev == nil {
		return nil, ErrNoInterfaceFound
	}

	return ifdev, nil
}

func readSysFs(iface string, file string) (string, error) {
	data, err := ioutil.ReadFile(filepath.Join("/sys/class/net", iface, file))
	return strings.TrimSpace(string(data)), err
}
