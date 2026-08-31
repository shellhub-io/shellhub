import { QueryClient } from "@tanstack/react-query";

/**
 * The single React Query client. Queries retry and go stale after 30 seconds; mutations do not
 * retry at all, because a retried write can apply twice.
 */
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 3,
      staleTime: 30_000,
      refetchOnWindowFocus: true,
    },
    mutations: {
      retry: 0,
    },
  },
});
