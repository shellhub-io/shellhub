package sysinfo

import (
	"errors"
	"io/ioutil"
	"math"
	"net"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

var ErrNoInterfaceFound = errors.New("no interface found")

func PrimaryInterface() (*net.Interface, error) {
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
			continue
		}

		iftype, err := strconv.ParseUint(data, 10, 16)
		if err != nil {
			continue
		}

		if iftype != syscall.ARPHRD_ETHER {
			continue
		}

		data, err = readSysFs(iface.Name, "ifindex")
		if err != nil {
			continue
		}

		ifindex, err := strconv.ParseUint(data, 10, 16)
		if err != nil {
			continue
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

func readSysFs(iface, file string) (string, error) {
	data, err := ioutil.ReadFile(filepath.Join("/sys/class/net", iface, file))

	return strings.TrimSpace(string(data)), err
}
