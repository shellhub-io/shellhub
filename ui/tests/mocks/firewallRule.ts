import { IFirewallRule } from "@/interfaces/IFirewallRule";
import { mockTags } from "@tests/mocks/tag";

/**
 * Mock firewall rule data for testing.
 * Provides a complete firewall rule object with all required fields.
 */
export const mockFirewallRule: IFirewallRule = {
  id: "rule-1",
  tenant_id: "fake-tenant",
  priority: 1,
  action: "allow",
  active: true,
  filter: { hostname: ".*" },
  source_ip: ".*",
  username: ".*",
};

/**
 * Mock firewall rules array for testing lists.
 * Provides multiple firewall rules for list/table testing scenarios.
 */
export const mockFirewallRules: IFirewallRule[] = [
  mockFirewallRule,
  {
    id: "rule-2",
    tenant_id: "fake-tenant",
    priority: 2,
    action: "deny",
    active: false,
    source_ip: "192.168.1.1",
    username: "testuser",
    filter: {
      tags: mockTags,
    },
  },
  {
    id: "rule-3",
    tenant_id: "fake-tenant",
    priority: 3,
    action: "allow",
    active: true,
    source_ip: "10.0.0.0/8",
    username: "admin",
    filter: {
      hostname: "server-.*",
    },
  },
];
