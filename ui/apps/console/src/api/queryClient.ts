import { QueryClient } from "@tanstack/react-query";
import { isSdkError } from "./errors";

/**
 * The single React Query client. Queries go stale after 30 seconds and retry up to three times,
 * except when the server answered with a client error (4xx), which no retry changes; mutations do
 * not retry at all, because a retried write can apply twice.
 */
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: (failureCount, error) =>
        failureCount < 3 && !(isSdkError(error) && error.status < 500),
      staleTime: 30_000,
      refetchOnWindowFocus: true,
    },
    mutations: {
      retry: 0,
    },
  },
});
