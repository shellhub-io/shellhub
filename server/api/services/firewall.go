package services

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

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
