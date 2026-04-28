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
  addressFilter?: string;
}

/**
 * Encodes a "contains" filter on the `address` property as the API expects:
 * a base64-encoded JSON array of filter clauses.
 */
function encodeAddressFilter(value: string): string {
  const clauses = [
    {
      type: "property",
      params: { name: "address", operator: "contains", value },
    },
  ];
  return btoa(JSON.stringify(clauses));
}

export function useWebEndpoints({
  page = 1,
  perPage = 10,
  addressFilter,
}: UseWebEndpointsParams = {}) {
  const trimmedFilter = addressFilter?.trim();
  const filter = trimmedFilter ? encodeAddressFilter(trimmedFilter) : undefined;

  const options = {
    query: {
      page,
      per_page: perPage,
      ...(filter ? { filter } : {}),
    },
  } satisfies { query: ListWebEndpointsData["query"] };

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
