import { useMemo } from "react";
import { useGetDevices } from "@/client/api";
import type { DeviceStatus, GetDevicesParams } from "@/client/model";
import { totalCount } from "@/api/pagination";
import { toBase64Json } from "@/utils/encoding";
import { normalizeDeviceTags } from "@/utils/deviceTags";

export type { TaggedDevice as NormalizedDevice } from "@/utils/deviceTags";

/**
 * Builds the device list filter the API expects: a search across name and hostname, and a tag
 * match, combined so an empty search or an empty tag list simply drops out.
 */
export function buildFilter(search: string, tags: string[]): string {
  const filters: Record<string, unknown>[] = [];
  if (search) {
    filters.push(
      { type: "operator", params: { name: "or" } },
      {
        type: "property",
        params: { name: "name", operator: "contains", value: search },
      },
      { type: "operator", params: { name: "or" } },
      {
        type: "property",
        params: { name: "custom_fields", operator: "contains", value: search },
      },
    );
  }
  if (tags.length > 0) {
    filters.push({
      type: "property",
      params: { name: "tags.name", operator: "contains", value: tags },
    });
  }
  return toBase64Json(filters);
}

interface UseDevicesParams {
  page?: number;
  perPage?: number;
  status?: DeviceStatus | "";
  search?: string;
  filterTags?: string[];
  enabled?: boolean;
  sortBy?: string;
  orderBy?: "asc" | "desc";
}

/**
 * A page of the namespace's devices, filtered by status, search and tags.
 */
export function useDevices({
  page = 1,
  perPage = 10,
  status = "",
  search = "",
  filterTags = [],
  enabled = true,
  sortBy = "last_seen",
  orderBy = "desc",
}: UseDevicesParams = {}) {
  const params: GetDevicesParams = {
    page,
    per_page: perPage,
    sort_by: sortBy,
    order_by: orderBy,
  };
  if (status) params.status = status;
  if (search || filterTags.length > 0)
    params.filter = buildFilter(search, filterTags);

  const result = useGetDevices(params, { query: { enabled } });

  const devices = useMemo(
    () => (result.data ?? []).map(normalizeDeviceTags),
    [result.data],
  );

  return {
    devices,
    totalCount: totalCount(result.data),
    isLoading: result.isLoading,
    error: result.error,
    refetch: result.refetch,
  };
}
