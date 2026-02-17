export interface PublicKeyFilter {
  hostname?: string;
  tags?: string[];
}

export interface PublicKey {
  name: string;
  fingerprint: string;
  data: string;
  username: string;
  filter: PublicKeyFilter;
  tenant_id: string;
  created_at: string;
}
