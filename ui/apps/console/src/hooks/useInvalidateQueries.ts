import { useQueryClient } from "@tanstack/react-query";

/**
 * Builds an invalidator for whole query families, matched on the first element of the key rather
 * than the whole key. A mutation cannot know which page or filter is cached, so it invalidates
 * by operation id and lets React Query refetch whichever are mounted.
 */
export function useInvalidateByIds(...ids: string[]) {
  const queryClient = useQueryClient();
  const idSet = new Set(ids);
  return () => queryClient.invalidateQueries({
    predicate: (query) => {
      const key = query.queryKey[0];
      return (
        typeof key === "object"
        && key !== null
        && "_id" in key
        && typeof key._id === "string"
        && idSet.has(key._id)
      );
    },
  });
}
