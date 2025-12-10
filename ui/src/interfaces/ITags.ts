export interface ITag {
  tenant_id: string;
  name: string;
  created_at: string;
  updated_at: string;
}

export interface FetchTagsParams {
  tenant: string;
  perPage: number;
  page: number;
  filter?: string;
}

export interface IUpdateTagName {
  name: string;
}
