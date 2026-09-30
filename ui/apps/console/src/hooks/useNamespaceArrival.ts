import { useEffect } from "react";
import { useQuery } from "@tanstack/react-query";
import { getNamespacesOptions } from "@/client";
import { useSwitchNamespace } from "@/hooks/useNamespaceMutations";

const CHECK_INTERVAL_MS = 5000;

/**
 * Waits for someone else to add the user to a namespace: checks now and every few seconds until
 * one exists, reports `ready` then, and `enter` switches into it. With `autoEnter`, it switches in
 * by itself as soon as the namespace shows up. `pending` covers a switch under way, and `failed`
 * one that did not go through; the automatic one is tried once, so the caller offers `enter`
 * again.
 */
export function useNamespaceArrival({ autoEnter = false } = {}) {
  const switchNs = useSwitchNamespace();
  const { data } = useQuery({
    ...getNamespacesOptions({ query: { page: 1, per_page: 1 } }),
    refetchInterval: (query) =>
      query.state.data?.length ? false : CHECK_INTERVAL_MS,
  });
  const tenantId = data?.[0]?.tenant_id ?? null;

  const enter = () => {
    if (tenantId) switchNs.mutate({ tenantId });
  };

  useEffect(() => {
    if (autoEnter && tenantId && switchNs.isIdle) {
      switchNs.mutate({ tenantId });
    }
  }, [autoEnter, tenantId, switchNs]);

  return {
    ready: tenantId !== null,
    enter,
    pending: switchNs.isPending,
    failed: switchNs.isError,
  };
}
