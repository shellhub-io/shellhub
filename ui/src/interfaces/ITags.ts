export interface Tags {
  name: string,
}

export interface FetchTagsParams {
  tenant: string;
  perPage: number;
  page: number;
  filter?: string;
}
