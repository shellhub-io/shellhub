export type filterType = {
  hostname?: string;
  tags?: Array<string>
}

export type FilterHostname = {
  hostname: string;
}

export type FilterTags = {
  tags: Set<string>;
}

export interface IPublicKey {
  created_at: string;
  data: string;
  filter: FilterHostname | FilterTags;
  fingerprint: string;
  name: string;
  tenant_id: string;
  username: string;
}
