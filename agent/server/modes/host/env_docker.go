//go:build docker

package host

func init() {
	for i, path := range localeFiles {
		localeFiles[i] = "/host" + path
	}
}
