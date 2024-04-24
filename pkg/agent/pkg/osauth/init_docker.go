//go:build docker
// +build docker

package osauth

func init() {
	DefaultShadowFilename = "/host/etc/shadow"
}
