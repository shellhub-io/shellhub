import { IFirewallRule } from "@/interfaces/IFirewallRule";

/**
 * Mock firewall rule data for testing.
 * Provides a complete firewall rule object with all required fields.
 */
export const mockFirewallRule: IFirewallRule = {
  id: "ffffffff-ffff-4fff-8fff-ffffffffffff",
  tenant_id: "fake-tenant-data",
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
export const mockFirewallRules: IFirewallRule[] = [mockFirewallRule];
