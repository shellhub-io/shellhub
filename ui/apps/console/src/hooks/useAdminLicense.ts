import { useQuery } from "@tanstack/react-query";
import { getLicenseQueryKey } from "../client/@tanstack/react-query.gen";
import { getLicense } from "../client";
import { useAuthStore } from "../stores/authStore";
import { isSdkError } from "../api/errors";
import { getConfig } from "../env";
import type { GetLicenseResponse } from "../client/types.gen";

export { getLicenseQueryKey };

type LicenseData = GetLicenseResponse | null;

export function useAdminLicense() {
  const isAdmin = useAuthStore((s) => s.isAdmin);
  const isCloud = getConfig().cloud;
  const enabled = isAdmin && !isCloud;

  const query = useQuery<LicenseData>({
    queryKey: getLicenseQueryKey(),
    queryFn: async ({ signal }) => {
      try {
        const { data } = await getLicense({ signal, throwOnError: true });
        return data;
      } catch (err) {
        // 400 means no license stored; normalize to "no license" instead of error state.
        if (isSdkError(err) && err.status === 400) return null;
        throw err;
      }
    },
    enabled,
    staleTime: 5 * 60 * 1000, // 5 minutes
    retry: (count) => count < 1,
    // Prevent refetch when the OS file picker closes and returns focus to the window.
    refetchOnWindowFocus: false,
  });

  const installedLicense =
    query.data && "grace_period" in query.data ? query.data : null;

  const isExpired =
    enabled &&
    !query.isLoading &&
    (!installedLicense || installedLicense.expired);

  return { ...query, installedLicense, isExpired };
}
