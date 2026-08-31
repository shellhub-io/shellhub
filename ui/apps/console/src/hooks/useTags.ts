import { useQuery } from "@tanstack/react-query";
import {
  getTags as getTagsSdk,
  getTagsQueryKey,
  type GetTagsData,
  type Tag,
} from "../client";
import { paginatedQueryFn, type PaginatedResult } from "../api/pagination";

interface UseTagsParams {
  page?: number;
  perPage?: number;
}

/**
 * The namespace's tags. Fetched a hundred at a time, since they are used to populate filters
 * rather than to be paged through.
 */
export function useTags({ page = 1, perPage = 100 }: UseTagsParams = {}) {
  const options = { query: { page, per_page: perPage } } satisfies { query: GetTagsData["query"] };

  const result = useQuery<PaginatedResult<Tag>>({
    queryKey: getTagsQueryKey(options),
    queryFn: paginatedQueryFn(getTagsSdk, options),
  });

  return {
    tags: result.data?.data ?? [],
    totalCount: result.data?.totalCount ?? 0,
    isLoading: result.isLoading,
    error: result.error,
  };
}
