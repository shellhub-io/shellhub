package services

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/geoip"
)

// LocatorProviderFactory constructs the MaxMind GeoLite2 geoip.Locator. Server setup calls it only
// when SHELLHUB_GEOIP_DATABASE is maxmind, so nothing may depend on it running. Cloud packages
// register a factory via RegisterLocatorProvider in their init() functions.
//
// The factory may return (nil, nil) when no MaxMind licence or mirror is configured. That, or an
// error, makes the server fall back to DB-IP, and to the null locator when no DB-IP database opens.
type LocatorProviderFactory func(ctx context.Context) (geoip.Locator, error)

var locatorFactory LocatorProviderFactory

// RegisterLocatorProvider registers the factory function that creates the
// geoip.Locator. This must be called before the server's Setup() is invoked —
// typically from a cloud package's init() function.
func RegisterLocatorProvider(f LocatorProviderFactory) {
	locatorFactory = f
}

// LocatorFactory returns the registered LocatorProviderFactory, or nil in CE builds.
func LocatorFactory() LocatorProviderFactory {
	return locatorFactory
}
