import { useQuery } from "@tanstack/react-query";
import {
  getLicense,
  getLicenseQueryKey,
  type GetLicenseResponse,
} from "../client";
import { useAuthStore } from "../stores/authStore";
import { isSdkError } from "../api/errors";
import { isCloud } from "../env";

export { getLicenseQueryKey };

type LicenseData = GetLicenseResponse | null;

export function useAdminLicense() {
  const isAdmin = useAuthStore((s) => s.isAdmin);
  const enabled = isAdmin && !isCloud();

  const query = useQuery<LicenseData>({
    queryKey: getLicenseQueryKey(),
    queryFn: async ({ signal }) => {
      try {
        const { data } = await getLicense({ signal, throwOnError: true });
        return data;
      } catch (err) {
        if (isSdkError(err) && err.status === 400) return null;
        throw err;
      }
    },
    enabled,
    staleTime: 5 * 60 * 1000, // 5 minutes
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
