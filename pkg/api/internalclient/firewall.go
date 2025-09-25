package internalclient

import (
	"context"
	"errors"
	"net/http"
)

// firewallAPI defines methods for interacting with firewall-related functionality.
type firewallAPI interface {
	// FirewallEvaluate evaluates firewall rules based on the provided lookup parameters.
	// It returns an error if the evaluation fails or if a firewall rule prohibits the connection.
	FirewallEvaluate(ctx context.Context, lookup map[string]string) error
}

var (
	// ErrFirewallEvaluationRequest is returned when the firewall evaluation request fails for any reason.
	ErrFirewallEvaluationRequest = errors.New("failed to evaluate the firewall")
	// ErrFirewallBlock is returned when a firewall rule prohibits the connection.
	ErrFirewallBlock = errors.New("a firewall rule prohibit this connection")
)

func (c *client) FirewallEvaluate(ctx context.Context, lookup map[string]string) error {
	resp, err := c.http.
		R().
		SetContext(ctx).
		SetQueryParams(lookup).
		Get(c.Config.EnterpriseBaseURL + "/internal/firewall/rules/evaluate")
	if err != nil {
		return errors.Join(ErrFirewallEvaluationRequest, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return ErrFirewallBlock
	}

	return nil
}
