import { useQuery } from "@tanstack/react-query";
import {
  getDevices as getDevicesSdk,
  getDevicesQueryKey,
  type Device,
} from "../client";
import { paginatedQueryFn, type PaginatedResult } from "../api/pagination";
import { toBase64Json } from "@/utils/encoding";

/**
 * How many of a member's devices a departure dialog can list, and so keep. It matches the
 * `maxItems` of `keep_devices` on the member endpoints.
 */
export const MAX_KEPT_DEVICES = 100;

/**
 * The accepted devices a member paired in the current namespace: the ones that leave with them
 * when they leave or can no longer accept devices. Only the first MAX_KEPT_DEVICES are fetched;
 * totalCount says how many there are.
 */
export function useMemberDevices(memberId: string, enabled = true) {
  const options = {
    query: {
      page: 1,
      per_page: MAX_KEPT_DEVICES,
      status: "accepted" as const,
      filter: toBase64Json([
        {
          type: "property",
          params: { name: "owner_id", operator: "eq", value: memberId },
        },
      ]),
    },
  };

  const result = useQuery<PaginatedResult<Device>>({
    queryKey: getDevicesQueryKey(options),
    queryFn: paginatedQueryFn(getDevicesSdk, options),
    enabled: enabled && !!memberId,
  });

  return {
    devices: result.data?.data ?? [],
    totalCount: result.data?.totalCount ?? 0,
    isLoading: result.isLoading,
    isError: result.isError,
  };
}
