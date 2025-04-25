export type filterType = {
  hostname?: string;
  tags?: Array<string>
}

export interface IFirewallRule {
  id: string;
  tenant_id: string;
  tenant_ip?: string;
  source_ip: string;
  username: string;
  priority: number;
  active: boolean;
  action: string;
  filter: {
    hostname?: string;
    tags?: string[];
  };
}
