package app

import (
	"context"
	"errors"
	"time"

	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/geoip"
	log "github.com/sirupsen/logrus"
)

const (
	geoipDownloadDir    = "/data/geoip"
	geoipBakedDir       = "/usr/share/shellhub/geoip"
	geoipUpdateInterval = 24 * time.Hour
	geoipUpdateTimeout  = 10 * time.Minute
)

const (
	geoipDatabaseDBIP    = "dbip"
	geoipDatabaseMaxMind = "maxmind"
	geoipDatabaseFake    = "fake"
)

const (
	geoipSourceDBIP    = "dbip"
	geoipSourceMaxMind = "maxmind"
)

var errUnknownGeoIPDatabase = errors.New("SHELLHUB_GEOIP_DATABASE must be dbip, maxmind or fake")

type locatorConfig struct {
	database    string
	factory     func(context.Context) (geoip.Locator, error)
	downloadDir string
	bakedDir    string
	update      bool
	updateURL   string
	clock       clock.Clock
}

type chosenLocator struct {
	locator   geoip.Locator
	source    string
	keepFresh func(context.Context)
}

func newLocator(ctx context.Context, cfg locatorConfig) (chosenLocator, error) {
	switch cfg.database {
	case geoipDatabaseMaxMind:
		if maxmind, ok := loadMaxMind(ctx, cfg.factory); ok {
			log.Info("GeoIP locator uses MaxMind GeoLite2")

			return chosenLocator{locator: maxmind, source: geoipSourceMaxMind}, nil
		}

		return openDBIP(cfg), nil
	case geoipDatabaseDBIP:
		return openDBIP(cfg), nil
	case geoipDatabaseFake:
		log.Warn("SHELLHUB_GEOIP_DATABASE is fake: devices on private addresses get made-up positions")

		chosen := openDBIP(cfg)
		chosen.locator = geoip.NewFake(chosen.locator)

		return chosen, nil
	default:
		return chosenLocator{}, errUnknownGeoIPDatabase
	}
}

func loadMaxMind(ctx context.Context, factory func(context.Context) (geoip.Locator, error)) (geoip.Locator, bool) {
	if factory == nil {
		log.Warn("SHELLHUB_GEOIP_DATABASE is maxmind, but this edition has no MaxMind: falling back to DB-IP")

		return nil, false
	}

	maxmind, err := factory(ctx)
	if err != nil {
		log.WithError(err).Warn("MaxMind GeoLite2 failed to load: falling back to DB-IP")

		return nil, false
	}

	if maxmind == nil {
		log.Warn("SHELLHUB_GEOIP_DATABASE is maxmind, but no MaxMind licence or mirror is set: falling back to DB-IP")

		return nil, false
	}

	return maxmind, true
}

func openDBIP(cfg locatorConfig) chosenLocator {
	dbip, err := geoip.OpenDBIP(cfg.downloadDir, cfg.bakedDir)
	if err != nil {
		log.WithError(err).Warn("no GeoIP database to open: devices get no position")

		return chosenLocator{locator: geoip.NewNullGeoLite()}
	}

	log.Info("GeoIP locator uses DB-IP City Lite")

	chosen := chosenLocator{locator: dbip, source: geoipSourceDBIP}
	if cfg.update {
		chosen.keepFresh = func(ctx context.Context) {
			dbip.KeepFresh(ctx, cfg.updateURL, cfg.clock, geoipUpdateInterval, geoipUpdateTimeout, func(err error) {
				log.WithError(err).Warn("failed to refresh the DB-IP GeoIP database; keeping the one in use")
			})
		}
	}

	return chosen
}
