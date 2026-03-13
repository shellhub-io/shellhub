export interface FirewallFilter {
  hostname?: string;
  tags?: string[];
}

export interface FirewallRule {
  id: string;
  tenant_id: string;
  priority: number;
  action: "allow" | "deny";
  active: boolean;
  source_ip: string;
  username: string;
  filter: FirewallFilter;
}
