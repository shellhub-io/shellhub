import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  getDevices as getDevicesSdk,
  type GetDevicesData,
  type DeviceStatus,
} from "../client";
import { getDevicesQueryKey } from "../client/@tanstack/react-query.gen";
import { paginatedQueryFn, type PaginatedResult } from "../api/pagination";
import type { Device as GeneratedDevice } from "../client";

export type NormalizedDevice = Omit<GeneratedDevice, "tags"> & { tags: string[] };

export function buildFilter(search: string, tags: string[]): string {
  const filters: Record<string, unknown>[] = [];
  if (search) {
    filters.push(
      { type: "operator", params: { name: "or" } },
      { type: "property", params: { name: "name", operator: "contains", value: search } },
      { type: "operator", params: { name: "or" } },
      { type: "property", params: { name: "custom_fields", operator: "contains", value: search } },
    );
  }
  if (tags.length > 0) {
    filters.push({
      type: "property",
      params: { name: "tags.name", operator: "contains", value: tags },
    });
  }
  return btoa(JSON.stringify(filters));
}

function normalizeDevice(device: GeneratedDevice): NormalizedDevice {
  return {
    ...device,
    tags: Array.isArray(device.tags)
      ? device.tags.map((t) =>
        typeof t === "object" && t !== null && "name" in t
          ? t.name
          : String(t),
      )
      : [],
  };
}

interface UseDevicesParams {
  page?: number;
  perPage?: number;
  status?: DeviceStatus | "";
  search?: string;
  filterTags?: string[];
}

export function useDevices({
  page = 1,
  perPage = 10,
  status = "",
  search = "",
  filterTags = [],
}: UseDevicesParams = {}) {
  const query: GetDevicesData["query"] = { page, per_page: perPage };
  if (status) query.status = status;
  if (search || filterTags.length > 0) query.filter = buildFilter(search, filterTags);

  const options = { query };

  const result = useQuery<PaginatedResult<GeneratedDevice>>({
    queryKey: getDevicesQueryKey(options),
    queryFn: paginatedQueryFn(getDevicesSdk, options),
  });

  const devices = useMemo(() => result.data?.data.map(normalizeDevice) ?? [], [result.data]);

  return {
    devices,
    totalCount: result.data?.totalCount ?? 0,
    isLoading: result.isLoading,
    error: result.error,
    refetch: result.refetch,
  };
}
