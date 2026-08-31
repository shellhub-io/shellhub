import { useQuery } from "@tanstack/react-query";
import { getConfig } from "@/env";
import {
  listAnnouncementsOptions,
  getAnnouncementOptions,
} from "@/client";

/**
 * The most recent announcement, or nothing when announcements are switched off for the instance.
 */
export function useLatestAnnouncement() {
  const enabled = getConfig().announcements;

  const listResult = useQuery({
    ...listAnnouncementsOptions({
      query: { page: 1, per_page: 1, order_by: "desc" },
    }),
    enabled,
    staleTime: 5 * 60 * 1000,
    refetchOnWindowFocus: false,
    retry: false,
  });

  const latestUuid = listResult.data?.[0]?.uuid;

  const detailResult = useQuery({
    ...getAnnouncementOptions({ path: { uuid: latestUuid ?? "" } }),
    enabled: enabled && !!latestUuid,
    staleTime: 5 * 60 * 1000,
    refetchOnWindowFocus: false,
    retry: false,
  });

  return {
    announcement: detailResult.data ?? null,
    isLoading:
      listResult.isLoading || (!!latestUuid && detailResult.isLoading),
  };
}
