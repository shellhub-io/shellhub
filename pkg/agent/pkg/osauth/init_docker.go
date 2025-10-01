//go:build docker
// +build docker

package osauth

func init() {
	DefaultPasswdFilename = "/host/etc/passwd"
	DefaultShadowFilename = "/host/etc/shadow"
}
