package geoip

import (
	"hash/fnv"
	"net"
)

type fakeCity struct {
	country   string
	latitude  float64
	longitude float64
}

var fakeCities = []fakeCity{
	{"CH", 47.3769, 8.5417},
	{"CH", 46.9480, 7.4474},
	{"CH", 46.2044, 6.1432},
	{"DE", 52.5200, 13.4050},
	{"DE", 48.1351, 11.5820},
	{"DE", 53.5511, 9.9937},
	{"DE", 50.9375, 6.9603},
	{"AT", 48.2082, 16.3738},
	{"FR", 48.8566, 2.3522},
	{"FR", 45.7640, 4.8357},
	{"NL", 52.3676, 4.9041},
	{"BE", 50.8503, 4.3517},
	{"IT", 45.4642, 9.1900},
	{"IT", 41.9028, 12.4964},
	{"ES", 40.4168, -3.7038},
	{"ES", 41.3874, 2.1686},
	{"PT", 38.7223, -9.1393},
	{"GB", 51.5072, -0.1276},
	{"IE", 53.3498, -6.2603},
	{"SE", 59.3293, 18.0686},
	{"NO", 59.9139, 10.7522},
	{"PL", 52.2297, 21.0122},
	{"CZ", 50.0755, 14.4378},
	{"US", 40.7128, -74.0060},
	{"US", 37.7749, -122.4194},
	{"US", 41.8781, -87.6298},
	{"US", 30.2672, -97.7431},
	{"CA", 43.6532, -79.3832},
	{"MX", 19.4326, -99.1332},
	{"BR", -23.5505, -46.6333},
	{"BR", -22.9068, -43.1729},
	{"BR", -30.0346, -51.2177},
	{"AR", -34.6037, -58.3816},
	{"CL", -33.4489, -70.6693},
	{"ZA", -26.2041, 28.0473},
	{"IN", 19.0760, 72.8777},
	{"IN", 12.9716, 77.5946},
	{"SG", 1.3521, 103.8198},
	{"JP", 35.6762, 139.6503},
	{"AU", -33.8688, 151.2093},
}

type fake struct {
	next Locator
}

// NewFake returns a Locator for development, where agents connect from private addresses no
// database places. It puts a private, loopback, link-local or unspecified address in one of a fixed
// list of real cities, the same city for the same address every time, never failing for one. It
// asks next for any other address and returns next's answer and error unchanged.
func NewFake(next Locator) Locator {
	return &fake{next: next}
}

func isLocal(ip net.IP) bool {
	return ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsUnspecified()
}

func cityOf(ip net.IP) fakeCity {
	hash := fnv.New32a()
	_, _ = hash.Write(ip.To16())

	return fakeCities[int(hash.Sum32()%uint32(len(fakeCities)))] //nolint:gosec // len(fakeCities) is a small constant, far below uint32
}

func (f *fake) GetCountry(ip net.IP) (string, error) {
	if isLocal(ip) {
		return cityOf(ip).country, nil
	}

	return f.next.GetCountry(ip)
}

func (f *fake) GetPosition(ip net.IP) (Position, error) {
	if isLocal(ip) {
		city := cityOf(ip)

		return Position{Latitude: city.latitude, Longitude: city.longitude}, nil
	}

	return f.next.GetPosition(ip)
}
