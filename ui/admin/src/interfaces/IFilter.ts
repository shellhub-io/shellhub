export enum NamespaceFilterOptions {
  MoreThan = "moreThan",
  NoDevices = "noDevices",
  NoSessions = "noSessions",
}

export interface IFilter {
  type: "property" | "operator";
  params: {
    name?: string;
    operator?: string;
    value?: number;
  };
}
