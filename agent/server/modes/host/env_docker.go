//go:build docker

package host

// The agent image runs from scratch and reaches the host's filesystem through /host, the same
// mount osauth reads passwd and shadow from.
func init() {
	for i, path := range localeFiles {
		localeFiles[i] = "/host" + path
	}
}
