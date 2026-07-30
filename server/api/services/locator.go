package services

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/geoip"
)

// LocatorProviderFactory is a function that constructs the geoip.Locator used to
// resolve IP addresses into countries and geographic positions. It is called
// during server setup when the -tags enterprise binary is built. Cloud packages
// register a factory via RegisterLocatorProvider in their init() functions.
//
// The factory may return (nil, nil) when the locator is not configured for the
// current deployment; in that case the service falls back to the null locator.
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
