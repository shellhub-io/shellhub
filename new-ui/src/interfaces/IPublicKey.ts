export type filterType = {
  hostname?: string;
  tags?: Array<string>
}

export interface IPublicKey {
  created_at: string;
  data: string;
  filter: filterType;
  fingerprint: string;
  name: string;
  tenant_id: string;
  username: string;
}