package services

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/cache"
)

// ========================================
// LICENSE EVALUATOR FACTORY
// ========================================

// LicenseEvaluatorFactory is a function that constructs a LicenseEvaluator using
// the core store and cache. It is called during server setup when the enterprise binary
// is built. Enterprise packages register a factory via RegisterLicenseEvaluator in their
// init() functions.
type LicenseEvaluatorFactoryFunc func(ctx context.Context, store store.Store, cache cache.Cache) (LicenseEvaluator, error)

var licenseEvaluatorFactory LicenseEvaluatorFactoryFunc

// RegisterLicenseEvaluator registers the factory function that creates the license
// evaluator. This must be called before the server's Setup() is invoked — typically
// from an enterprise package's init() function.
func RegisterLicenseEvaluator(f LicenseEvaluatorFactoryFunc) {
	licenseEvaluatorFactory = f
}

// LicenseEvaluatorFactory returns the registered LicenseEvaluatorFactoryFunc, or nil
// in Community Edition builds.
func LicenseEvaluatorFactory() LicenseEvaluatorFactoryFunc {
	return licenseEvaluatorFactory
}

// ========================================
// INTERFACES
// ========================================

// LicenseEvaluator defines the interface for license-based device acceptance checks.
//
// The evaluator is injected into the core service to gate device acceptance against
// license limits. When present, license validation is performed in-process.
//
// The evaluator is optional — when nil, license checks are skipped (Community Edition).
type LicenseEvaluator interface {
	// CanAcceptDevice reports whether the current license allows accepting an
	// additional device.
	//
	// Returns:
	//   - true if the license permits acceptance
	//   - false if the device limit has been reached
	//   - error if the check itself fails (e.g., license fetch error)
	CanAcceptDevice(ctx context.Context) (bool, error)
}
