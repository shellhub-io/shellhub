package services

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	log "github.com/sirupsen/logrus"
)

// BillingProviderFactory is a function that constructs a BillingProvider using the
// core store and cache. It is called during server setup when the -tags enterprise binary
// is built. Cloud packages register a factory via RegisterBillingProvider in their
// init() functions.
type BillingProviderFactory func(ctx context.Context, store store.Store, cache cache.Cache) (BillingProvider, error)

var billingFactory BillingProviderFactory

// RegisterBillingProvider registers the factory function that creates the billing
// provider. This must be called before the server's Setup() is invoked — typically
// from a cloud package's init() function.
func RegisterBillingProvider(f BillingProviderFactory) {
	billingFactory = f
}

// BillingFactory returns the registered BillingProviderFactory, or nil in CE builds.
func BillingFactory() BillingProviderFactory {
	return billingFactory
}

// BillingProvider defines the interface for enterprise/cloud billing integrations.
//
// The provider is injected into the core service using the WithBilling() option.
// When present, billing validation and reporting are performed in-process rather
// than via HTTP calls to external services.
//
// Implementations should be thin adapters that:
//   - Translate between core types and billing service types
//   - Map billing errors to core service error types
//   - Handle billing-specific business logic (quotas, subscriptions, etc.)
//
// The provider is optional - when nil, billing features are disabled (Community Edition).
//
// Example usage:
//
//	cloudBilling := billingAdapter.NewCoreBillingAdapter(cloudBillingService)
//	coreService := services.NewService(store, privKey, pubKey, cache,
//	    services.WithBilling(cloudBilling))
type BillingProvider interface {
	// Evaluate checks if a namespace can accept or connect more devices based on
	// its billing status, subscription state, and quota limits.
	//
	// Returns:
	//   - BillingEvaluation with CanAccept and CanConnect flags
	//   - Error if evaluation fails (network error, invalid subscription, etc.)
	Evaluate(ctx context.Context, tenant string) (*models.BillingEvaluation, error)

	// Report notifies the billing system of namespace usage for quota tracking and billing.
	//
	// Common actions:
	//   - BillingActionDeviceAccept: Report new device acceptance
	//   - BillingActionNamespaceDelete: Report namespace deletion
	//
	// Returns error if:
	//   - Namespace has no active subscription (ErrPaymentRequired)
	//   - Quota exceeded (ErrPaymentRequired)
	//   - Network/gateway error (ErrReport)
	Report(ctx context.Context, tenant string, action BillingAction) error
}

// BillingAction represents an action to report to the billing system.
type BillingAction string

const (
	BillingActionDeviceAccept    BillingAction = "device_accept"
	BillingActionNamespaceDelete BillingAction = "namespace_delete"
)

type BillingService interface {
	// EvaluateBilling reports whether billing lets an SSH connection to the
	// namespace through, returning ErrBillingBlocked when it does not.
	EvaluateBilling(ctx context.Context, tenant string) error

	// ReportBilling notifies the billing system of a namespace action. It
	// returns ErrPaymentRequired when the namespace's subscription forbids it —
	// deleting a namespace that still has an active subscription, for one.
	ReportBilling(ctx context.Context, tenant string, action BillingAction) error
}

// EvaluateBilling reports whether the namespace's billing state lets a connection to an
// already accepted device through. It reads CanConnect, not CanAccept: at the exact plan
// limit a namespace stops accepting new devices but keeps SSH working on the ones it
// already has.
//
// A missing provider allows the connection. Device acceptance guards on
// `IsCloud() && s.billing != nil` and so is never reached without a provider, but the
// SSH path guards on IsCloud() alone — blocking here would take down every connection on
// a cloud whose billing failed to initialize.
func (s *service) EvaluateBilling(ctx context.Context, tenant string) error {
	if s.billing == nil {
		return nil
	}

	evaluation, err := s.billing.Evaluate(ctx, tenant)
	if err != nil {
		return NewErrBillingEvaluate(err)
	}

	if !evaluation.CanConnect {
		log.WithFields(log.Fields{
			"tenant":  tenant,
			"blocked": evaluation.Blocked,
		}).Error("billing blocked the connection")

		return ErrBillingBlocked
	}

	return nil
}

// evaluateBilling reports what the namespace's billing state allows.
// It fails when the billing provider is not available (Community Edition).
func (s *service) evaluateBilling(ctx context.Context, tenant string) (*models.BillingEvaluation, error) {
	if s.billing == nil {
		return nil, ErrBillingNotAvailable
	}

	evaluation, err := s.billing.Evaluate(ctx, tenant)
	if err != nil {
		return nil, NewErrBillingEvaluate(err)
	}

	return evaluation, nil
}

func (s *service) ReportBilling(ctx context.Context, tenant string, action BillingAction) error {
	return s.reportBilling(ctx, tenant, action)
}

// reportBilling notifies the billing system of a namespace action.
func (s *service) reportBilling(ctx context.Context, tenant string, action BillingAction) error {
	if s.billing == nil {
		return ErrBillingNotAvailable
	}

	if err := s.billing.Report(ctx, tenant, action); err != nil {
		// The provider adapter already maps errors to appropriate types
		// (ErrPaymentRequired for subscription issues, ErrReport for others)
		return err
	}

	return nil
}

// validateBillingForDeviceAcceptance checks billing and reports device acceptance.
// This is called during device acceptance in cloud environments.
func (s *service) validateBillingForDeviceAcceptance(ctx context.Context, namespace *models.Namespace) error {
	if namespace.Billing.IsActive() {
		// Active subscription - report acceptance for quota tracking
		if err := s.reportBilling(ctx, namespace.TenantID, BillingActionDeviceAccept); err != nil {
			return NewErrBillingReportNamespaceDelete(err)
		}
	} else {
		// Inactive subscription - evaluate if namespace can still accept
		evaluation, err := s.evaluateBilling(ctx, namespace.TenantID)
		if err != nil {
			return NewErrBillingEvaluate(err)
		}

		if !evaluation.CanAccept {
			// The two denials need different answers from the user: one has to remove a device
			// or upgrade, the other has to finish or repair the subscription. Reporting both as
			// a device limit sends a namespace that is under its allowance to delete devices
			// that were never the problem.
			if evaluation.Blocked == models.BillingBlockedSubscription {
				return ErrDeviceBillingBlocked
			}

			return ErrDeviceLimit
		}
	}

	return nil
}
