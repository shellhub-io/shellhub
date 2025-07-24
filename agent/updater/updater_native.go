//go:build !docker
// +build !docker

package updater

func NewUpdater(version string) (Updater, error) {
	return &nativeUpdater{version}, nil
}
