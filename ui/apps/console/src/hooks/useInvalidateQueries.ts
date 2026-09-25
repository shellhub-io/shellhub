import { useQueryClient } from "@tanstack/react-query";

/**
 * The operation id a generated query key was built from, or undefined for a key the generated
 * client did not make.
 */
export function queryOperationId(queryKey: readonly unknown[]): string | undefined {
  const head = queryKey[0];
  return typeof head === "object"
    && head !== null
    && "_id" in head
    && typeof head._id === "string"
    ? head._id
    : undefined;
}

/**
 * Builds an invalidator for whole query families, matched on the first element of the key rather
 * than the whole key. A mutation cannot know which page or filter is cached, so it invalidates
 * by operation id and lets React Query refetch whichever are mounted.
 */
export function useInvalidateByIds(...ids: string[]) {
  const queryClient = useQueryClient();
  const idSet = new Set(ids);
  return () => queryClient.invalidateQueries({
    predicate: (query) => idSet.has(queryOperationId(query.queryKey) ?? ""),
  });
}
