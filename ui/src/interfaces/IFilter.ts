export type HostnameFilter = {
  hostname: string;
}

export type TagsFilter = {
  tags: string[];
}

export type Filter = HostnameFilter | TagsFilter;

export enum FormFilterOptions {
  All = "all",
  Hostname = "hostname",
  Tags = "tags",
}
