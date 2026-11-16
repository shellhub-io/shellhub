import { queryOptions, useQueries } from "@tanstack/react-query";
import {
  getContainers,
  getContainersQueryKey,
  getDevices,
  getDevicesQueryKey,
  type Device,
} from "@/client";
import { paginatedQueryFn, type PaginatedResult } from "@/api/pagination";
import { useDevice } from "@/hooks/useDeviceRequests";
import { propertyFilter, toBase64Json } from "@/utils/encoding";

const LOOKUP_FAILED =
  "Could not check whether accepting replaces an offline device. Close this and try again.";

const ONLINE_RECHECK_MS = 5_000;

function onlineMessage(name: string): string {
  return `${name} reports the same identity and is online. Stop its agent, then accept this device.`;
}

/**
 * useReplacedDevice looks up the accepted device or container in the current namespace that
 * reports the same identity as the device uid, an empty MAC included, since the server matches on
 * the identity alone, across platforms. replaced is that device while it is offline: the accept
 * replaces it, deleting its record and giving its name to the device accepted, and is null
 * otherwise. acceptBlocked is true while the lookup runs, once it has failed, and while that device
 * is online, when the server refuses the accept; the lookup then reads again every few seconds, so
 * replaced takes over once its agent stops. blockedMessage says why the accept is refused, and is
 * null while the lookup runs or when nothing refuses it. An empty uid reads nothing and blocks
 * nothing.
 */
export function useReplacedDevice(uid: string): {
  replaced: Device | null;
  acceptBlocked: boolean;
  blockedMessage: string | null;
} {
  const {
    device,
    isLoading: deviceLoading,
    error: deviceError,
  } = useDevice(uid);
  const identity =
    device && device.status !== "accepted" ? device.identity : undefined;
  const options = {
    query: {
      page: 1,
      per_page: 1,
      status: "accepted" as const,
      filter: toBase64Json([propertyFilter("mac", "eq", identity?.mac ?? "")]),
    },
  };
  const enabled = !!uid && !!identity;
  const recheckWhileOnline = (query: {
    state: { data?: PaginatedResult<Device> };
  }) => (query.state.data?.data[0]?.online ? ONLINE_RECHECK_MS : false);
  const sameIdentity = useQueries({
    queries: [
      queryOptions<PaginatedResult<Device>>({
        queryKey: getDevicesQueryKey(options),
        queryFn: paginatedQueryFn(getDevices, options),
        enabled,
        refetchInterval: recheckWhileOnline,
      }),
      queryOptions<PaginatedResult<Device>>({
        queryKey: getContainersQueryKey(options),
        queryFn: paginatedQueryFn(getContainers, options),
        enabled,
        refetchInterval: recheckWhileOnline,
      }),
    ],
    combine: (results) => ({
      device: results.flatMap((r) => r.data?.data ?? [])[0],
      isLoading: results.some((r) => r.isLoading),
      isError: results.some((r) => r.isError),
    }),
  });

  const sameIdentityDevice =
    sameIdentity.device && sameIdentity.device.uid !== uid
      ? sameIdentity.device
      : null;

  const online = sameIdentityDevice?.online ? sameIdentityDevice : null;
  const failed = !!deviceError || sameIdentity.isError;

  return {
    replaced: sameIdentityDevice && !online ? sameIdentityDevice : null,
    acceptBlocked:
      deviceLoading || sameIdentity.isLoading || failed || !!online,
    blockedMessage: failed
      ? LOOKUP_FAILED
      : online
        ? onlineMessage(online.name)
        : null,
  };
}
