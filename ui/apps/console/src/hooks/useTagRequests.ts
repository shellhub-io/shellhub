import { useQuery } from "@tanstack/react-query";
import { getTags, getTagsQueryKey } from "@/client";
import { paginatedQueryOptions } from "@/api/pagination";

/**
 * The namespace's first hundred tags, for populating filters rather than paging through. Empty
 * while loading and when the request fails.
 */
export function useTags() {
  const { data, isLoading } = useQuery(
    paginatedQueryOptions(getTags, getTagsQueryKey, {
      query: { page: 1, per_page: 100 },
    }),
  );

  return { tags: data?.data ?? [], isLoading };
}
