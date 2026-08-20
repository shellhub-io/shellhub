import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  getDevicesAdmin,
  getDevicesAdminQueryKey,
  getDeviceAdminOptions,
  type GetDevicesAdminData,
  type Device,
  type DeviceStatus,
} from "../client";
import { paginatedQueryFn, type PaginatedResult } from "../api/pagination";
import { useAuthStore } from "../stores/authStore";
import { isSdkError } from "../api/errors";
import { toBase64Json } from "@/utils/encoding";
import { normalizeDeviceTags } from "@/utils/deviceTags";

export type { TaggedDevice as NormalizedDevice } from "@/utils/deviceTags";

function buildNameFilter(search: string): string {
  const filter = [
    {
      type: "property",
      params: { name: "name", operator: "contains", value: search },
    },
  ];
  return toBase64Json(filter);
}

interface UseAdminDevicesParams {
  page?: number;
  perPage?: number;
  search?: string;
  status?: DeviceStatus | "";
  sortBy?: string;
  orderBy?: "asc" | "desc";
}

export function useAdminDevices({
  page = 1,
  perPage = 10,
  search = "",
  status = "",
  sortBy = "last_seen",
  orderBy = "desc",
}: UseAdminDevicesParams = {}) {
  const isAdmin = useAuthStore((s) => s.isAdmin);

  const query: GetDevicesAdminData["query"] = {
    page,
    per_page: perPage,
    sort_by: sortBy,
    order_by: orderBy,
  };
  if (search) query.filter = buildNameFilter(search);
  if (status) query.status = status;
  const options = { query };

  const result = useQuery<PaginatedResult<Device>>({
    queryKey: getDevicesAdminQueryKey(options),
    queryFn: paginatedQueryFn(getDevicesAdmin, options),
    enabled: isAdmin,
    staleTime: 5 * 60 * 1000, // 5 minutes
    retry: (count, err) =>
      isSdkError(err) && err.status === 401 ? false : count < 1,
    refetchOnWindowFocus: false,
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

export function useAdminDevice(uid: string) {
  const isAdmin = useAuthStore((s) => s.isAdmin);

  return useQuery({
    ...getDeviceAdminOptions({ path: { uid } }),
    enabled: isAdmin && !!uid,
    staleTime: 5 * 60 * 1000, // 5 minutes
    retry: (count, err) =>
      isSdkError(err) && err.status === 401 ? false : count < 1,
    refetchOnWindowFocus: false,
    select: normalizeDeviceTags,
  });
}
