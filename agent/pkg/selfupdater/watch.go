package selfupdater

import (
	"context"
	"time"

	"github.com/Masterminds/semver/v3"
	log "github.com/sirupsen/logrus"
)

// Watch checks the server's version through check as it starts and again on every tick, applying
// it through updater whenever it is greater than current, until ctx is done. A failed check or a
// failed update is logged through logger and waits for the next tick. An agent whose version is
// "latest" is a development build, so Watch returns at once without checking.
func Watch(ctx context.Context, logger *log.Entry, version string, current *semver.Version, check func() (*semver.Version, error), updater Updater, tick <-chan time.Time) {
	if version == "latest" {
		return
	}

	for {
		nextVersion, err := check()
		if err != nil {
			logger.WithError(err).WithFields(log.Fields{
				"version": version,
			}).Error("Failed to check update")
		} else if nextVersion.GreaterThan(current) {
			if err := updater.ApplyUpdate(nextVersion); err != nil {
				logger.WithError(err).WithFields(log.Fields{
					"version":      version,
					"next_version": nextVersion.String(),
				}).Error("Failed to apply update")
			} else {
				logger.WithFields(log.Fields{
					"version":      current,
					"next_version": nextVersion.String(),
				}).Info("Update successfully applied")
			}
		}

		logger.WithFields(log.Fields{
			"version": version,
		}).Info("Waiting for the next update check")

		select {
		case <-ctx.Done():
			return
		case <-tick:
		}
	}
}
