package internalclient

import (
	"errors"
	"net"
	"net/http"

	"github.com/go-resty/resty/v2"
)

// firewallAPI defines methods for interacting with firewall-related functionality.
type firewallAPI interface {
	// FirewallEvaluate evaluates firewall rules based on the provided lookup parameters.
	// It returns an error if the evaluation fails or if a firewall rule prohibits the connection.
	FirewallEvaluate(lookup map[string]string) error
}

var (
	ErrFirewallConnection = errors.New("failed to make the request to evaluate the firewall")
	ErrFirewallBlock      = errors.New("a firewall rule prohibit this connection")
)

func (c *client) FirewallEvaluate(lookup map[string]string) error {
	local := resty.New()
	local.AddRetryCondition(func(r *resty.Response, err error) bool {
		if _, ok := err.(net.Error); ok {
			return true
		}

		return r.StatusCode() >= http.StatusInternalServerError && r.StatusCode() != http.StatusNotImplemented
	})

	resp, err := local.
		SetRetryCount(10).
		R().
		SetQueryParams(lookup).
		Get("http://cloud-api:8080/internal/firewall/rules/evaluate")
	if err != nil {
		return ErrFirewallConnection
	}

	if resp.StatusCode() != http.StatusOK {
		return ErrFirewallBlock
	}

	return nil
}
