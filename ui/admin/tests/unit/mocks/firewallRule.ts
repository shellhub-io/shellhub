import { IAdminFirewallRule } from "@admin/interfaces/IFirewallRule";

export const mockFirewallRule: IAdminFirewallRule = {
  id: "rule-123",
  tenant_id: "tenant-123",
  priority: 1,
  action: "allow" as const,
  active: true,
  source_ip: "192.168.1.0/24",
  username: "testuser",
  filter: {
    hostname: ".*",
  },
};

export const mockFirewallRules: IAdminFirewallRule[] = [
  { ...mockFirewallRule, id: "rule-1", priority: 1, action: "allow" as const },
  { ...mockFirewallRule, id: "rule-2", priority: 2, action: "deny" as const, active: false },
  { ...mockFirewallRule, id: "rule-3", priority: 3, action: "allow" as const },
];
