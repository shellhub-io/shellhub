package selfupdater

import (
	"github.com/Masterminds/semver"
)

type Updater interface {
	CurrentVersion() (*semver.Version, error)
	ApplyUpdate(v *semver.Version) error
	CompleteUpdate() error
}

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
