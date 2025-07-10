export enum AdminNamespaceFilterOptions {
  MoreThan = "moreThan",
  NoDevices = "noDevices",
  NoSessions = "noSessions",
}

export interface IAdminFilter {
  type: "property" | "operator";
  params: {
    name?: string;
    operator?: string;
    value?: number;
  };
}
