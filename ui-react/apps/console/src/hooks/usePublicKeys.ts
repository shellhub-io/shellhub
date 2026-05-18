import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  getPublicKeys as getPublicKeysSdk,
  type GetPublicKeysData,
  type PublicKeyResponse,
} from "../client";
import { getPublicKeysQueryKey } from "../client/@tanstack/react-query.gen";
import { paginatedQueryFn, type PaginatedResult } from "../api/pagination";

interface UsePublicKeysParams {
  page?: number;
  perPage?: number;
}

export function usePublicKeys({
  page = 1,
  perPage = 10,
}: UsePublicKeysParams = {}) {
  const options = { query: { page, per_page: perPage } satisfies GetPublicKeysData["query"] };

  const result = useQuery<PaginatedResult<PublicKeyResponse>>({
    queryKey: getPublicKeysQueryKey(options),
    queryFn: paginatedQueryFn(getPublicKeysSdk, options),
  });

  const publicKeys = useMemo(
    () => result.data?.data ?? [],
    [result.data],
  );

  return {
    publicKeys,
    totalCount: result.data?.totalCount ?? 0,
    isLoading: result.isLoading,
    error: result.error,
  };
}
