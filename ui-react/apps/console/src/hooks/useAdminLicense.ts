import { useQuery } from "@tanstack/react-query";
import {
  getLicenseOptions,
  getLicenseQueryKey,
} from "../client/@tanstack/react-query.gen";

export { getLicenseQueryKey };

export function useAdminLicense() {
  return useQuery({
    ...getLicenseOptions(),
    staleTime: 5 * 60 * 1000, // 5 minutes
    retry: 1,
  });
}
