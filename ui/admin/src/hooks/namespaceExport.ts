import { IAdminExportFilter, AdminNamespaceFilterOptions } from "@admin/interfaces/IFilter";

const getFilter = (option: AdminNamespaceFilterOptions, numberOfDevices: number): IAdminExportFilter[] => {
  const filters: Record<AdminNamespaceFilterOptions, IAdminExportFilter[]> = {
    [AdminNamespaceFilterOptions.MoreThan]: [
      {
        type: "property",
        params: {
          name: "devices_accepted_count",
          operator: "gt",
          value: numberOfDevices,
        },
      },
    ],
    [AdminNamespaceFilterOptions.NoDevices]: [
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
    [AdminNamespaceFilterOptions.NoSessions]: [
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
