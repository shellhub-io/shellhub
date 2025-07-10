export type FirewallRuleFilter = {
  hostname?: string;
  tags?: Array<string>
}

export interface IFirewallRule {
  action: "allow" | "deny";
  active: boolean;
  filter: FirewallRuleFilter;
  id: string;
  tenant_id: string;
  priority: number;
  source_ip: string;
  username: string;
  status: string;
}
