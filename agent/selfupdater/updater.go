package selfupdater

import (
	"github.com/Masterminds/semver"
)

type Updater interface {
	CurrentVersion() (*semver.Version, error)
	ApplyUpdate(v *semver.Version) error
	CompleteUpdate() error
}
