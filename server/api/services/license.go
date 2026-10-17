package services

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/server/api/store"
)

// LicenseEvaluatorFactoryFunc constructs a LicenseEvaluator using
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
	//   - ErrLicenseNotAuthorized if the installed license does not authorize this edition
	//   - another error if the check itself fails (e.g., license fetch error)
	//
	// It reads the license and the device count through ctx, so a read that fails aborts
	// any transaction ctx carries. Call it before opening one.
	CanAcceptDevice(ctx context.Context) (bool, error)

	// CanConnectDevice reports whether the current license allows connecting to an
	// already registered device. It differs from CanAcceptDevice at the exact
	// limit: a fleet sitting on its licensed device count keeps working, it just
	// cannot grow. It returns ErrLicenseNotAuthorized when the installed license does not
	// authorize this edition, and another error when the check itself fails.
	CanConnectDevice(ctx context.Context) (bool, error)
}

// LicenseService answers whether the instance's licence permits an action. The community
// edition permits everything; the enterprise binary registers a real evaluator.
type LicenseService interface {
	// EvaluateLicense reports whether the license lets an SSH connection through. It
	// returns ErrLicenseBlocked when the device limit refuses it, ErrLicenseNotAuthorized
	// when the installed license does not authorize this edition, and the evaluator's error
	// when the license cannot be evaluated.
	EvaluateLicense(ctx context.Context) error
}

// EvaluateLicense reports whether the license lets an SSH connection through. Without
// an evaluator (Community Edition) there is no license to exceed, so the connection is
// allowed. It returns ErrLicenseBlocked when the device limit refuses the connection,
// ErrLicenseNotAuthorized when the installed license does not authorize this edition,
// and the evaluator's error when the license cannot be evaluated.
//
// Unlike the device acceptance path, which fails open outside Cloud, a license that
// cannot be evaluated always blocks the connection: that is what the HTTP evaluation
// this replaced did, and it is the conservative direction for a paid limit.
func (s *service) EvaluateLicense(ctx context.Context) error {
	if s.licenseEvaluator == nil {
		return nil
	}

	canConnect, err := s.licenseEvaluator.CanConnectDevice(ctx)
	if err != nil {
		return err
	}

	if !canConnect {
		return ErrLicenseBlocked
	}

	return nil
}
