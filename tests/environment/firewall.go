package environment

import (
	"context"
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

// The actions a [FirewallRule] takes on the connections it matches.
const (
	FirewallAllow = "allow"
	FirewallDeny  = "deny"
)

// FirewallRule is a namespace firewall rule as the enterprise API takes it. SourceIP, Username and
// Filter.Hostname are regular expressions matched against the whole value; Filter names either a
// hostname or tags, never both.
type FirewallRule struct {
	Priority int                `json:"priority"`
	Action   string             `json:"action"`
	Active   bool               `json:"active"`
	SourceIP string             `json:"source_ip"`
	Username string             `json:"username"`
	Filter   FirewallRuleFilter `json:"filter"`
}

// FirewallRuleFilter selects the devices a [FirewallRule] applies to, by a hostname pattern or by
// tag names, matching a device that carries any one of them. Every tag must already exist in the
// namespace.
type FirewallRuleFilter struct {
	Hostname string   `json:"hostname,omitempty"`
	Tags     []string `json:"tags,omitempty"`
}

// CreateFirewallRule adds rule to the authenticated namespace and returns its ID. It fails the test
// unless the server answers 200 with a non-empty ID, which it does only on an enterprise or cloud
// edition licensed for firewall rules, to a bearer allowed to create them, for a valid rule whose
// tags exist in the namespace.
func (dc *DockerCompose) CreateFirewallRule(t *testing.T, rule FirewallRule) string {
	t.Helper()

	created := struct {
		ID string `json:"id"`
	}{}

	resp, err := dc.R(t.Context()).SetBody(rule).SetResult(&created).Post("/api/firewall/rules")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
	require.NotEmpty(t, created.ID, resp.String())

	return created.ID
}

// DeleteFirewallRule removes the rule id from the authenticated namespace, failing the test unless
// the server answers 200. It runs on a context that outlives the test, so it can sit in a cleanup.
func (dc *DockerCompose) DeleteFirewallRule(t *testing.T, id string) {
	t.Helper()

	resp, err := dc.R(context.WithoutCancel(t.Context())).Delete("/api/firewall/rules/" + id)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
}

// FirewallRuleCount returns how many firewall rules the authenticated namespace has, all pages
// together, as the server's X-Total-Count header states it. It fails the test unless the server
// answers 200 with that header holding a number.
func (dc *DockerCompose) FirewallRuleCount(t *testing.T) int {
	t.Helper()

	resp, err := dc.R(t.Context()).Get("/api/firewall/rules")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	count, err := strconv.Atoi(resp.Header().Get("X-Total-Count"))
	require.NoError(t, err)

	return count
}
