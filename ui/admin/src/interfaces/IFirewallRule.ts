export type filterType = {
  hostname?: string;
  tags?: Array<string>
}

export interface IFirewallRule {
  action: boolean;
  active: boolean;
  filter: filterType;
  id: string;
  tenant_id: string;
  priority: number;
  source_ip: string;
  tenant_ip: string;
  username: string;
}
