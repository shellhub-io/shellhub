package main

import (
	"context"
	"net/http"
	"regexp"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

const (
	firewallBlockedLog   = "destination device is blocked by a firewall rule"
	errFirewallRuleBlock = "because a firewall rule block your connection"
	anyValue             = ".*"
)

var anyDevice = environment.FirewallRuleFilter{Hostname: anyValue}

// TestEnterpriseFirewall covers how the legacy-mode firewall decides an SSH connection: the first
// active rule, lowest priority value first, matching the client's source address, the device user
// and the device by hostname or tag decides; a connection no rule matches, or with no rules at all,
// passes; an inactive rule is skipped; and a namespace in the identity mode decides by its access
// policies instead, never reading its firewall rules. The cases share one stack and one device,
// and each removes the rules it adds.
func TestEnterpriseFirewall(t *testing.T) {
	ctx := context.Background()

	compose := newEnterpriseEnvironment(t, ctx, environment.New(t, run))
	signer := registerDeviceKey(t, ctx, compose)
	_, device := startAcceptedAgent(t, ctx, compose)

	sshid := deviceSSHID(device)
	source := regexp.QuoteMeta(clientAddress(t, compose, sshid, signer))
	hostname := environment.FirewallRuleFilter{Hostname: regexp.QuoteMeta(device.Name)}
	denyAll := deny(100, matching(anyValue, anyValue, anyDevice))

	createTag(t, compose, "firewalled")
	createTag(t, compose, "elsewhere")
	tagDevice(t, compose, device.UID, "firewalled")

	t.Run("a connection is let through when the namespace has no rules", func(t *testing.T) {
		require.Zero(t, compose.FirewallRuleCount(t))

		requireKeyLogsIn(t, compose, sshid, signer)
	})

	t.Run("an allow rule matching the source address, the user and the hostname lets the connection through", func(t *testing.T) {
		addFirewallRule(t, compose, denyAll)
		addFirewallRule(t, compose, allow(1, matching(source, ShellHubAgentUsername, hostname)))

		requireKeyLogsIn(t, compose, sshid, signer)
	})

	denials := []struct {
		description string
		rule        environment.FirewallRule
	}{
		{description: "source address", rule: matching(source, anyValue, anyDevice)},
		{description: "user", rule: matching(anyValue, ShellHubAgentUsername, anyDevice)},
		{description: "hostname", rule: matching(anyValue, anyValue, hostname)},
		{description: "tag", rule: matching(anyValue, anyValue, environment.FirewallRuleFilter{Tags: []string{"firewalled"}})},
	}

	for _, tc := range denials {
		t.Run("a deny rule matching the "+tc.description+" blocks the connection", func(t *testing.T) {
			addFirewallRule(t, compose, deny(1, tc.rule))

			requireAccessDenied(t, compose, sshid, signer, firewallBlockedLog, errFirewallRuleBlock)
		})
	}

	misses := []struct {
		description string
		rule        environment.FirewallRule
	}{
		{description: "another source address", rule: matching(`192\.0\.2\.1`, anyValue, anyDevice)},
		{description: "another user", rule: matching(anyValue, "nobody", anyDevice)},
		{description: "another hostname", rule: matching(anyValue, anyValue, environment.FirewallRuleFilter{Hostname: "elsewhere"})},
		{description: "a tag the device does not carry", rule: matching(anyValue, anyValue, environment.FirewallRuleFilter{Tags: []string{"elsewhere"}})},
	}

	for _, tc := range misses {
		t.Run("a deny rule for "+tc.description+" lets the connection through", func(t *testing.T) {
			addFirewallRule(t, compose, deny(1, tc.rule))

			requireKeyLogsIn(t, compose, sshid, signer)
		})
	}

	t.Run("an allow rule with a lower priority value wins over a matching deny rule", func(t *testing.T) {
		addFirewallRule(t, compose, denyAll)
		addFirewallRule(t, compose, allow(1, matching(anyValue, anyValue, anyDevice)))

		requireKeyLogsIn(t, compose, sshid, signer)
	})

	t.Run("a deny rule with a lower priority value wins over a matching allow rule", func(t *testing.T) {
		addFirewallRule(t, compose, allow(200, matching(anyValue, anyValue, anyDevice)))
		addFirewallRule(t, compose, denyAll)

		requireAccessDenied(t, compose, sshid, signer, firewallBlockedLog, errFirewallRuleBlock)
	})

	t.Run("an inactive deny rule is skipped", func(t *testing.T) {
		inactive := denyAll
		inactive.Active = false
		addFirewallRule(t, compose, inactive)

		requireKeyLogsIn(t, compose, sshid, signer)
	})

	t.Run("a namespace in the identity mode is not held to its firewall rules", func(t *testing.T) {
		compose.EditSSHAccessMode(t, ShellHubNamespace, models.SSHAccessModeIdentity)
		t.Cleanup(func() { compose.EditSSHAccessMode(t, ShellHubNamespace, models.SSHAccessModeLegacy) })

		owner, _ := enrollOwnerKey(t, compose)
		addFirewallRule(t, compose, denyAll)

		requireStraightThrough(t, compose, sshid, owner)
	})
}

func matching(sourceIP, username string, filter environment.FirewallRuleFilter) environment.FirewallRule {
	return environment.FirewallRule{Active: true, SourceIP: sourceIP, Username: username, Filter: filter}
}

func allow(priority int, rule environment.FirewallRule) environment.FirewallRule {
	rule.Priority = priority
	rule.Action = environment.FirewallAllow

	return rule
}

func deny(priority int, rule environment.FirewallRule) environment.FirewallRule {
	rule.Priority = priority
	rule.Action = environment.FirewallDeny

	return rule
}

func addFirewallRule(t *testing.T, compose *environment.DockerCompose, rule environment.FirewallRule) {
	t.Helper()

	id := compose.CreateFirewallRule(t, rule)
	t.Cleanup(func() { compose.DeleteFirewallRule(t, id) })
}

func clientAddress(t *testing.T, compose *environment.DockerCompose, sshid string, signer ssh.Signer) string {
	t.Helper()

	requireKeyLogsIn(t, compose, sshid, signer)

	sessions := []models.Session{}

	resp, err := compose.R(t.Context()).SetResult(&sessions).Get("/api/sessions")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
	require.NotEmpty(t, sessions)
	require.NotEmpty(t, sessions[0].IPAddress)

	return sessions[0].IPAddress
}
