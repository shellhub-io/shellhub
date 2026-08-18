import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  getContainers as getContainersSdk,
  type GetContainersData,
} from "../client";
import type { Device, DeviceStatus } from "../client";
import { getContainersQueryKey } from "../client/@tanstack/react-query.gen";
import { paginatedQueryFn, type PaginatedResult } from "../api/pagination";
import { toBase64Json } from "@/utils/encoding";
import { normalizeDeviceTags } from "@/utils/deviceTags";

export type { TaggedDevice as NormalizedContainer } from "@/utils/deviceTags";

function buildFilter(search: string, tags: string[]): string {
  const filters: Record<string, unknown>[] = [];
  if (search) {
    filters.push({
      type: "property",
      params: { name: "name", operator: "contains", value: search },
    });
  }
  if (tags.length > 0) {
    filters.push({
      type: "property",
      params: { name: "tags.name", operator: "contains", value: tags },
    });
  }
  return toBase64Json(filters);
}

interface UseContainersParams {
  page?: number;
  perPage?: number;
  status?: DeviceStatus | "";
  search?: string;
  filterTags?: string[];
  sortBy?: string;
  orderBy?: "asc" | "desc";
}

export function useContainers({
  page = 1,
  perPage = 10,
  status = "",
  search = "",
  filterTags = [],
  sortBy = "last_seen",
  orderBy = "desc",
}: UseContainersParams = {}) {
  const query: GetContainersData["query"] = { page, per_page: perPage };
  if (status) query.status = status;
  if (search || filterTags.length > 0)
    query.filter = buildFilter(search, filterTags);
  query.sort_by = sortBy;
  query.order_by = orderBy;

  const options = { query };

  const result = useQuery<PaginatedResult<Device>>({
    queryKey: getContainersQueryKey(options),
    queryFn: paginatedQueryFn(getContainersSdk, options),
  });

  const containers = useMemo(
    () => result.data?.data.map(normalizeDeviceTags) ?? [],
    [result.data],
  );

  return {
    containers,
    totalCount: result.data?.totalCount ?? 0,
    isLoading: result.isLoading,
    error: result.error,
    refetch: result.refetch,
  };
}
