import { useQuery } from "@tanstack/react-query";
import { useAuthStore } from "@/stores/authStore";
import { storedBrowserKeyFingerprint } from "@/utils/browserKey";

/** Invalidated by whoever enrolls this browser's key, so the list stops calling
 *  it somebody else's browser the moment it becomes ours. */
export const BROWSER_KEY_QUERY_KEY = ["browser-key"];

/**
 * The fingerprint of the SSH key this browser holds for the signed-in account,
 * or null when it holds none. It is what lets a list of browser identities say
 * which one belongs to the browser the person is actually sitting at — every
 * other one is a browser they would have to go back to.
 *
 * The read is local (IndexedDB), so it is cached for the session rather than
 * refetched: the key only changes when this tab enrolls one.
 */
export function useBrowserKeyFingerprint(): string | null {
  const tenant = useAuthStore((s) => s.tenant);
  const userId = useAuthStore((s) => s.userId);
  const scope = userId && tenant ? `${userId}:${tenant}` : null;

  const { data } = useQuery({
    queryKey: BROWSER_KEY_QUERY_KEY.concat(scope ?? ""),
    queryFn: () => (scope ? storedBrowserKeyFingerprint(scope) : null),
    staleTime: Infinity,
  });

  return data ?? null;
}

/**
 * The fingerprint of the key this browser holds in each of the given namespaces for the signed-in
 * user, keyed by tenant; a namespace without one maps to null. Read locally, like
 * useBrowserKeyFingerprint, and invalidated with it.
 */
export function useBrowserKeyFingerprints(
  tenants: string[],
): Record<string, string | null> {
  const userId = useAuthStore((s) => s.userId);

  const { data } = useQuery({
    queryKey: BROWSER_KEY_QUERY_KEY.concat(userId ?? "", ...tenants),
    queryFn: async () =>
      Object.fromEntries(
        await Promise.all(
          tenants.map(
            async (tenant) =>
              [
                tenant,
                await storedBrowserKeyFingerprint(`${userId}:${tenant}`),
              ] as const,
          ),
        ),
      ),
    enabled: !!userId && tenants.length > 0,
    staleTime: Infinity,
  });

  return data ?? {};
}
