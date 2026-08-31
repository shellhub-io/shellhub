import { useQuery } from "@tanstack/react-query";
import {
  listWebEndpoints as listWebEndpointsSdk,
  listWebEndpointsQueryKey,
  type ListWebEndpointsData,
  type Webendpoint,
} from "../client";
import { paginatedQueryFn, type PaginatedResult } from "../api/pagination";
import { toBase64Json } from "@/utils/encoding";

interface UseWebEndpointsParams {
  page?: number;
  perPage?: number;
  addressFilter?: string;
}

function encodeAddressFilter(value: string): string {
  const clauses = [
    {
      type: "property",
      params: { name: "address", operator: "contains", value },
    },
  ];
  return toBase64Json(clauses);
}

/**
 * A page of the namespace's web endpoints, optionally filtered by address. A blank filter is
 * dropped rather than sent, so an empty search box does not narrow the list to nothing.
 */
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
