import { useMemo } from "react";
import { useQueries, useQuery } from "@tanstack/react-query";
import {
  getDevices as getDevicesSdk,
  getDevicesQueryKey,
  type GetDevicesData,
  type DeviceStatus,
  type Device as GeneratedDevice,
} from "../client";
import { paginatedQueryFn, type PaginatedResult } from "../api/pagination";
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
  refetchInterval?:
    | number
    | false
    | ((page?: PaginatedResult<GeneratedDevice>) => number | false);
  sortBy?: string;
  orderBy?: "asc" | "desc";
}

/**
 * A page of the namespace's devices, filtered by status, search and tags. refetchInterval keeps it
 * polling, and as a function it gets the latest page and returns false to stop.
 */
export function useDevices({
  page = 1,
  perPage = 10,
  status = "",
  search = "",
  filterTags = [],
  enabled = true,
  refetchInterval = false,
  sortBy = "last_seen",
  orderBy = "desc",
}: UseDevicesParams = {}) {
  const query: GetDevicesData["query"] = { page, per_page: perPage };
  if (status) query.status = status;
  if (search || filterTags.length > 0)
    query.filter = buildFilter(search, filterTags);
  query.sort_by = sortBy;
  query.order_by = orderBy;

  const options = { query };

  const result = useQuery<PaginatedResult<GeneratedDevice>>({
    queryKey: getDevicesQueryKey(options),
    queryFn: paginatedQueryFn(getDevicesSdk, options),
    enabled,
    refetchInterval:
      typeof refetchInterval === "function"
        ? (query) => refetchInterval(query.state.data)
        : refetchInterval,
  });

  const devices = useMemo(
    () => result.data?.data.map(normalizeDeviceTags) ?? [],
    [result.data],
  );

  return {
    devices,
    totalCount: result.data?.totalCount ?? 0,
    isLoading: result.isLoading,
    error: result.error,
    refetch: result.refetch,
  };
}

function unverifiedDevicesOptions(status: "accepted" | "pending") {
  return {
    query: {
      page: 1,
      per_page: 1,
      status,
      filter: toBase64Json([
        {
          type: "property",
          params: { name: "key_verified", operator: "bool", value: false },
        },
      ]),
    },
  };
}

/**
 * How many accepted or pending devices have an agent that never signed a server challenge with
 * its device key, and so stop connecting once the namespace enforces device signatures. It reads
 * the totals of one-device pages rather than the devices. count is undefined until both totals
 * have been read once; after that it keeps the last totals, while disabled and after a failed
 * read too. isLoading is true while either total is being read, the first time or again, so a
 * caller that waits for it never acts on a stale count.
 */
export function useUnverifiedDeviceCount({ enabled = true } = {}) {
  const results = useQueries({
    queries: (["accepted", "pending"] as const).map((status) => {
      const options = unverifiedDevicesOptions(status);
      return {
        queryKey: getDevicesQueryKey(options),
        queryFn: paginatedQueryFn(getDevicesSdk, options),
        enabled,
      };
    }),
  });

  const counts = results.map((result) => result.data?.totalCount);

  return {
    count: counts.some((count) => count === undefined)
      ? undefined
      : counts.reduce<number>((sum, count) => sum + (count ?? 0), 0),
    isLoading: results.some((result) => result.isFetching),
    error: results.find((result) => result.error)?.error ?? null,
  };
}
