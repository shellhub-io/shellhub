import { useQuery } from "@tanstack/react-query";
import {
  listAnnouncementsAdmin as listAnnouncementsAdminSdk,
  type ListAnnouncementsAdminData,
  type AnnouncementShort,
} from "../client";
import {
  listAnnouncementsAdminQueryKey,
  getAnnouncementAdminOptions,
} from "../client/@tanstack/react-query.gen";
import { paginatedQueryFn, type PaginatedResult } from "../api/pagination";
import { useAuthStore } from "../stores/authStore";
import { isSdkError } from "../api/errors";

interface UseAdminAnnouncementsParams {
  page?: number;
  perPage?: number;
}

export function useAdminAnnouncements({
  page = 1,
  perPage = 10,
}: UseAdminAnnouncementsParams = {}) {
  const isAdmin = useAuthStore((s) => s.isAdmin);

  const query: ListAnnouncementsAdminData["query"] = {
    page,
    per_page: perPage,
    order_by: "desc",
  };
  const options = { query };

  const result = useQuery<PaginatedResult<AnnouncementShort>>({
    queryKey: listAnnouncementsAdminQueryKey(options),
    queryFn: paginatedQueryFn(listAnnouncementsAdminSdk, options),
    enabled: isAdmin,
    staleTime: 5 * 60 * 1000,
    retry: (count, err) =>
      isSdkError(err) && err.status === 401 ? false : count < 1,
    refetchOnWindowFocus: false,
  });

  return {
    announcements: result.data?.data ?? [],
    totalCount: result.data?.totalCount ?? 0,
    isLoading: result.isLoading,
    error: result.error,
    refetch: result.refetch,
  };
}

export function useAdminAnnouncement(uuid: string) {
  const isAdmin = useAuthStore((s) => s.isAdmin);

  return useQuery({
    ...getAnnouncementAdminOptions({ path: { uuid } }),
    enabled: isAdmin && !!uuid,
    staleTime: 5 * 60 * 1000,
    retry: (count, err) =>
      isSdkError(err) && err.status === 401 ? false : count < 1,
    refetchOnWindowFocus: false,
  });
}
