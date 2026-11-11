package services

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/geoip"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
)

// FirewallEvaluatorFactoryFunc constructs a FirewallEvaluator from the core store, the cache and
// the locator the server places connections with, the one SHELLHUB_GEOIP_DATABASE chose. The
// locator is never nil: with no database to open it answers an empty country. Enterprise packages
// register a factory via RegisterFirewallEvaluator in their init() functions; it runs during
// server setup.
type FirewallEvaluatorFactoryFunc func(ctx context.Context, store store.Store, cache cache.Cache, locator geoip.Locator) (FirewallEvaluator, error)

var firewallEvaluatorFactory FirewallEvaluatorFactoryFunc

// RegisterFirewallEvaluator registers the factory that creates the firewall evaluator.
// It must be called before the server's Setup() runs.
func RegisterFirewallEvaluator(f FirewallEvaluatorFactoryFunc) {
	firewallEvaluatorFactory = f
}

// FirewallEvaluatorFactory returns the registered factory, or nil in Community Edition
// builds.
func FirewallEvaluatorFactory() FirewallEvaluatorFactoryFunc {
	return firewallEvaluatorFactory
}

// FirewallEvaluator reports whether the legacy firewall (a key/username/source-IP
// blocklist) lets an SSH connection through. It is enterprise-only and is bypassed
// entirely when the namespace is new, where Access Policies are
// the authorization model.
type FirewallEvaluator interface {
	// AllowsConnection returns nil when the connection may proceed,
	// ErrFirewallBlocked when a rule, a licensed region or a licensed device limit
	// denies it, and any other error when the evaluation itself failed. Rules match
	// conn.IPAddress, the client's address; the licensed regions apply to
	// conn.DeviceIPAddress, the address the device's agent last authenticated from,
	// and a device whose country cannot be resolved is let through.
	AllowsConnection(ctx context.Context, conn models.FirewallConnection) error
}

// FirewallService answers whether a namespace's firewall rules permit a connection.
type FirewallService interface {
	// EvaluateFirewall reports whether the firewall lets conn through, returning
	// ErrFirewallBlocked when it does not.
	EvaluateFirewall(ctx context.Context, conn models.FirewallConnection) error
}

// EvaluateFirewall reports whether the firewall lets conn through. Without an evaluator
// (Community Edition) there are no rules, so the connection is allowed.
func (s *service) EvaluateFirewall(ctx context.Context, conn models.FirewallConnection) error {
	if s.firewallEvaluator == nil {
		return nil
	}

	return s.firewallEvaluator.AllowsConnection(ctx, conn)
}
