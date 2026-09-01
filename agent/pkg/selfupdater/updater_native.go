//go:build !docker

package selfupdater

// NewUpdater returns the [Updater] for an agent installed as a plain binary, reporting version
// as the one currently running.
func NewUpdater(version string) (Updater, error) {
	return &nativeUpdater{version}, nil
}
