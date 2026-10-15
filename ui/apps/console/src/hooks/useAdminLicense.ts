import { useQuery } from "@tanstack/react-query";
import {
  getLicense,
  getLicenseQueryKey,
  type GetLicenseResponse,
} from "../client";
import { useAuthStore } from "../stores/authStore";
import { isSdkError } from "../api/errors";
import { isEnterprise } from "../env";

type LicenseData = GetLicenseResponse | null;

/**
 * The installed licence, or null when there is none. Run only for an instance admin on enterprise,
 * the one edition with a licence endpoint, and not when the caller passes active false.
 */
export function useAdminLicense({ active = true }: { active?: boolean } = {}) {
  const isAdmin = useAuthStore((s) => s.isAdmin);
  const enabled = active && isAdmin && isEnterprise();

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
