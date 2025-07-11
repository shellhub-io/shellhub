export type HostnameFilter = {
  hostname: string;
}

export type TagsFilter = {
  tags: Set<string> | string[];
}

export type Filter = HostnameFilter | TagsFilter;
