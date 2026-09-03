import { useQuery } from "@tanstack/react-query";
import { getLicense, getGetLicenseQueryKey } from "@/client/api";
import type { GetLicense200 } from "@/client/model";
import { useAuthStore } from "@/stores/authStore";
import { isSdkError } from "@/api/errors";
import { isCloud } from "@/env";

export { getGetLicenseQueryKey };

type LicenseData = GetLicense200 | null;

/**
 * The installed licence, or null when there is none. Not run on cloud, where licensing is the
 * provider's concern and the endpoint does not exist.
 */
export function useAdminLicense() {
  const isAdmin = useAuthStore((s) => s.isAdmin);
  const enabled = isAdmin && !isCloud();

  const query = useQuery<LicenseData>({
    queryKey: getGetLicenseQueryKey(),
    queryFn: async () => {
      try {
        return await getLicense();
      } catch (err) {
        if (isSdkError(err) && err.status === 400) return null;
        throw err;
      }
    },
    enabled,
    staleTime: 5 * 60 * 1000,
    retry: (count) => count < 1,
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
