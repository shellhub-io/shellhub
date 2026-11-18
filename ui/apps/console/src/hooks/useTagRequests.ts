import { useQuery } from "@tanstack/react-query";
import { createTag, getTags, getTagsQueryKey } from "@/client";
import { isSdkError } from "@/api/errors";
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

/**
 * Creates the tag unless it already exists, so it can be pushed to a device or container next.
 * Rethrows every other failure.
 */
export async function ensureTag(name: string): Promise<void> {
  try {
    await createTag({ body: { name }, throwOnError: true });
  } catch (e) {
    if (!isSdkError(e) || e.status !== 409) throw e;
  }
}
