package services

import (
	"context"
)

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
	// any transaction ctx carries. Inside a transaction, call it through
	// [store.TransactionStore.WithSavepoint].
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
