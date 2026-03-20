import { useQuery } from "@tanstack/react-query";
import {
  getLicenseOptions,
  getLicenseQueryKey,
} from "../client/@tanstack/react-query.gen";
import { useAuthStore } from "../stores/authStore";

export { getLicenseQueryKey };

export function useAdminLicense() {
  const isAdmin = useAuthStore((s) => s.isAdmin);

  return useQuery({
    ...getLicenseOptions(),
    enabled: isAdmin,
    staleTime: 5 * 60 * 1000, // 5 minutes
    retry: 1,
  });
}
