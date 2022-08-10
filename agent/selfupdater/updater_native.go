//go:build !docker
// +build !docker

package selfupdater

func NewUpdater(version string) (Updater, error) {
	return &nativeUpdater{version}, nil
}
