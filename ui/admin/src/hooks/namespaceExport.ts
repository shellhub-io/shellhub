import { IFilter, NamespaceFilterOptions } from "@admin/interfaces/IFilter";

const getFilter = (option: NamespaceFilterOptions, numberOfDevices: number): IFilter[] => {
  const filters: Record<NamespaceFilterOptions, IFilter[]> = {
    [NamespaceFilterOptions.MoreThan]: [
      {
        type: "property",
        params: {
          name: "devices_accepted_count",
          operator: "gt",
          value: numberOfDevices,
        },
      },
    ],
    [NamespaceFilterOptions.NoDevices]: [
      {
        type: "property",
        params: { name: "devices_accepted_count", operator: "eq", value: 0 },
      },
      {
        type: "property",
        params: { name: "devices_pending_count", operator: "eq", value: 0 },
      },
      {
        type: "property",
        params: { name: "devices_rejected_count", operator: "eq", value: 0 },
      },
      { type: "operator", params: { name: "and" } },
    ],
    [NamespaceFilterOptions.NoSessions]: [
      {
        type: "property",
        params: { name: "devices_accepted_count", operator: "gt", value: 0 },
      },
      {
        type: "property",
        params: { name: "sessions", operator: "eq", value: 0 },
      },
      { type: "operator", params: { name: "and" } },
    ],
  };

  return filters[option];
};

export default getFilter;
