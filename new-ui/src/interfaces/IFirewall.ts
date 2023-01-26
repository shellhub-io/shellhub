interface IFilter {
  hostname?: string;
  tags?: Array<string>;
}

export interface IFirewall {
  id: string;
  tenant_id: string;
  priority: number;
  action: string;
  active: boolean;
  source_ip: string;
  username: string;
  filter: IFilter;
}
