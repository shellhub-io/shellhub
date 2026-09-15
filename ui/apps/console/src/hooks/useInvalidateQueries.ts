import { useQueryClient } from "@tanstack/react-query";

/**
 * Builds an invalidator that matches orval-generated query keys by URL path prefix. Orval keys
 * have the endpoint URL as the first array element, so `/api/devices` matches both the list
 * (`/api/devices`) and a detail (`/api/devices/uid-123`).
 */
export function useInvalidateByIds(...pathPrefixes: string[]) {
  const queryClient = useQueryClient();
  return () =>
    queryClient.invalidateQueries({
      predicate: (query) => {
        const head = query.queryKey[0];
        return (
          typeof head === "string" &&
          pathPrefixes.some((p) => head.startsWith(p))
        );
      },
    });
}
