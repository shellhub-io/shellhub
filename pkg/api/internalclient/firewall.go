package internalclient

import (
	"context"
)

// firewallAPI defines methods for interacting with firewall-related functionality.
type firewallAPI interface {
	// FirewallEvaluate evaluates firewall rules based on the provided lookup parameters.
	// It returns an error if the evaluation fails or if a firewall rule prohibits the connection.
	FirewallEvaluate(ctx context.Context, lookup map[string]string) error
}

func (c *client) FirewallEvaluate(ctx context.Context, lookup map[string]string) error {
	resp, err := c.http.
		R().
		SetContext(ctx).
		SetQueryParams(lookup).
		Get(c.config.APIBaseURL + "/internal/firewall/rules/evaluate")

	return HasError(resp, err)
}
