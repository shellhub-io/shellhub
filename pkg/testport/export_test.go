//go:build linux

package testport

var EphemeralPortStart = ephemeralPortStart

func SetEphemeralPortRangePath(path string) (restore func()) {
	previous := ephemeralPortRangePath
	ephemeralPortRangePath = path

	return func() { ephemeralPortRangePath = previous }
}
