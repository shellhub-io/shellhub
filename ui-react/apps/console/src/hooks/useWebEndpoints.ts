import { useQuery } from "@tanstack/react-query";
import {
  listWebEndpoints as listWebEndpointsSdk,
  type ListWebEndpointsData,
  type Webendpoint,
} from "../client";
import { listWebEndpointsQueryKey } from "../client/@tanstack/react-query.gen";
import { paginatedQueryFn, type PaginatedResult } from "../api/pagination";

interface UseWebEndpointsParams {
  page?: number;
  perPage?: number;
}

export function useWebEndpoints({ page = 1, perPage = 10 }: UseWebEndpointsParams = {}) {
  const options = { query: { page, per_page: perPage } } satisfies { query: ListWebEndpointsData["query"] };

  const result = useQuery<PaginatedResult<Webendpoint>>({
    queryKey: listWebEndpointsQueryKey(options),
    queryFn: paginatedQueryFn(listWebEndpointsSdk, options),
  });

  return {
    webEndpoints: result.data?.data ?? [],
    totalCount: result.data?.totalCount ?? 0,
    isLoading: result.isLoading,
    error: result.error,
  };
}
