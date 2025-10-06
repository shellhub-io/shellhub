//go:build docker
// +build docker

package osauth

func init() {
	DefaultPasswdFilename = "/host/etc/passwd" //nolint: gosec
	DefaultShadowFilename = "/host/etc/shadow"
}
