import { useQuery } from "@tanstack/react-query";
import {
  getDevices as getDevicesSdk,
  type GetDevicesData,
} from "../client";
import { getDevicesQueryKey } from "../client/@tanstack/react-query.gen";
import { paginatedQueryFn, type PaginatedResult } from "../api/pagination";
import type { Device as GeneratedDevice } from "../client";

export type NormalizedDevice = Omit<GeneratedDevice, "tags"> & { tags: string[] };

function buildTagFilter(tags: string[]): string {
  const filter = [{
    type: "property",
    params: { name: "tags.name", operator: "contains", value: tags },
  }];
  return btoa(JSON.stringify(filter));
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
  status?: string;
  filterTags?: string[];
}

export function useDevices({
  page = 1,
  perPage = 10,
  status = "",
  filterTags = [],
}: UseDevicesParams = {}) {
  const query: GetDevicesData["query"] = { page, per_page: perPage };
  if (status) query.status = status as GetDevicesData["query"] extends { status?: infer S } ? S : never;
  if (filterTags.length > 0) query.filter = buildTagFilter(filterTags);

  const options = { query };

  const result = useQuery<PaginatedResult<GeneratedDevice>>({
    queryKey: getDevicesQueryKey(options),
    queryFn: paginatedQueryFn<GeneratedDevice>(
      getDevicesSdk as never,
      options,
    ),
  });

  return {
    devices: result.data?.data.map(normalizeDevice) ?? [],
    totalCount: result.data?.totalCount ?? 0,
    isLoading: result.isLoading,
    error: result.error,
    refetch: result.refetch,
  };
}
