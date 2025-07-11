import { HostnameFilter, TagsFilter } from "@/interfaces/IFilter";

export enum AdminNamespaceFilterOptions {
  MoreThan = "moreThan",
  NoDevices = "noDevices",
  NoSessions = "noSessions",
}

export interface IAdminExportFilter {
  type: "property" | "operator";
  params: {
    name?: string;
    operator?: string;
    value?: number;
  };
}

export type AdminHostnameFilter = HostnameFilter;
export type AdminTagsFilter = TagsFilter;
export type AdminFilter = AdminHostnameFilter | AdminTagsFilter;
