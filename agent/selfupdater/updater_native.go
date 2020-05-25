// +build !docker

package selfupdater

import (
	"github.com/Masterminds/semver"
)

type nativeUpdater struct {
	version string
}

func (n *nativeUpdater) CurrentVersion() (*semver.Version, error) {
	return semver.NewVersion(n.version)
}

func (n *nativeUpdater) ApplyUpdate(_ *semver.Version) error {
	return nil
}

func (n *nativeUpdater) CompleteUpdate() error {
	return nil
}

func NewUpdater(version string) (Updater, error) {
	return &nativeUpdater{version}, nil
}
