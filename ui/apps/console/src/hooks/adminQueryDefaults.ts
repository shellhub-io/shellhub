/**
 * What every admin query shares: five minutes before it counts as stale and no refetch when the
 * window regains focus. Retries follow the query client's default, which never retries a 4xx.
 */
export const adminQueryDefaults = {
  staleTime: 5 * 60 * 1000,
  refetchOnWindowFocus: false,
};
