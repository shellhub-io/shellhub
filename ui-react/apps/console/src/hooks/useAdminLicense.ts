import { useQuery } from "@tanstack/react-query";
import {
  getLicenseOptions,
  getLicenseQueryKey,
} from "../client/@tanstack/react-query.gen";
import { useAuthStore } from "../stores/authStore";
import { isSdkError } from "../api/errors";

export { getLicenseQueryKey };

export function useAdminLicense() {
  const isAdmin = useAuthStore((s) => s.isAdmin);

  return useQuery({
    ...getLicenseOptions(),
    enabled: isAdmin,
    staleTime: 5 * 60 * 1000, // 5 minutes
    // 400 means no license stored — deterministic, no point retrying.
    retry: (count, err) => isSdkError(err) && err.status === 400 ? false : count < 1,
    // Prevent refetch when the OS file picker closes and returns focus to the window.
    refetchOnWindowFocus: false,
  });
}
