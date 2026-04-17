import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  getContainers as getContainersSdk,
  type GetContainersData,
} from "../client";
import type { Device, DeviceStatus } from "../client";
import { getContainersQueryKey } from "../client/@tanstack/react-query.gen";
import { paginatedQueryFn, type PaginatedResult } from "../api/pagination";

export type NormalizedContainer = Omit<Device, "tags"> & { tags: string[] };

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
  return btoa(JSON.stringify(filters));
}

export function normalizeContainer(container: Device): NormalizedContainer {
  return {
    ...container,
    tags: Array.isArray(container.tags)
      ? container.tags.map((t) =>
          typeof t === "object" && t !== null && "name" in t
            ? t.name
            : String(t),
        )
      : [],
  };
}

interface UseContainersParams {
  page?: number;
  perPage?: number;
  status?: DeviceStatus | "";
  search?: string;
  filterTags?: string[];
}

export function useContainers({
  page = 1,
  perPage = 10,
  status = "",
  search = "",
  filterTags = [],
}: UseContainersParams = {}) {
  const query: GetContainersData["query"] = { page, per_page: perPage };
  if (status) query.status = status;
  if (search || filterTags.length > 0)
    query.filter = buildFilter(search, filterTags);

  const options = { query };

  const result = useQuery<PaginatedResult<Device>>({
    queryKey: getContainersQueryKey(options),
    queryFn: paginatedQueryFn(getContainersSdk, options),
  });

  const containers = useMemo(
    () => result.data?.data.map(normalizeContainer) ?? [],
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
