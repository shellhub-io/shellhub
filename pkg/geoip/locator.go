package geoip

import "net"

type Locator interface {
	GetCountry(ip net.IP) (string, error)
	GetPosition(ip net.IP) (Position, error)
}
