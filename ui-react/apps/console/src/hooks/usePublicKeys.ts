import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  getPublicKeys as getPublicKeysSdk,
  type GetPublicKeysData,
  type PublicKeyResponse,
} from "../client";
import { getPublicKeysQueryKey } from "../client/@tanstack/react-query.gen";
import { paginatedQueryFn, type PaginatedResult } from "../api/pagination";

export function buildPublicKeyFilter(search: string): string {
  const filters = [
    { type: "operator", params: { name: "or" } },
    {
      type: "property",
      params: { name: "name", operator: "contains", value: search },
    },
    { type: "operator", params: { name: "or" } },
    {
      type: "property",
      params: { name: "fingerprint", operator: "contains", value: search },
    },
  ];
  return btoa(JSON.stringify(filters));
}

interface UsePublicKeysParams {
  page?: number;
  perPage?: number;
  search?: string;
}

export function usePublicKeys({
  page = 1,
  perPage = 10,
  search = "",
}: UsePublicKeysParams = {}) {
  const query: GetPublicKeysData["query"] = { page, per_page: perPage };
  if (search) query.filter = buildPublicKeyFilter(search);

  const options = { query };

  const result = useQuery<PaginatedResult<PublicKeyResponse>>({
    queryKey: getPublicKeysQueryKey(options),
    queryFn: paginatedQueryFn(getPublicKeysSdk, options),
  });

  const publicKeys = useMemo(() => result.data?.data ?? [], [result.data]);

  return {
    publicKeys,
    totalCount: result.data?.totalCount ?? 0,
    isLoading: result.isLoading,
    error: result.error,
  };
}
