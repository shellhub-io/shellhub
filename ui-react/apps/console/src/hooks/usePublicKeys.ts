import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  getPublicKeys as getPublicKeysSdk,
  type GetPublicKeysData,
  type PublicKeyResponse,
} from "../client";
import { getPublicKeysQueryKey } from "../client/@tanstack/react-query.gen";
import { paginatedQueryFn, type PaginatedResult } from "../api/pagination";

export interface PublicKeyFilter {
  hostname?: string;
  tags?: string[];
}

export interface PublicKey {
  data: string;
  fingerprint: string;
  created_at: string;
  tenant_id: string;
  name: string;
  filter: PublicKeyFilter;
  username: string;
}

function normalizePublicKey(pk: PublicKeyResponse): PublicKey {
  let filter: PublicKeyFilter;
  if ("tags" in pk.filter) {
    filter = {
      tags: pk.filter.tags.map((t) =>
        typeof t === "object" && t !== null && "name" in t
          ? t.name
          : String(t),
      ),
    };
  } else {
    filter = { hostname: pk.filter.hostname };
  }

  return {
    ...pk,
    filter,
  };
}

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
    () => result.data?.data.map(normalizePublicKey) ?? [],
    [result.data],
  );

  return {
    publicKeys,
    totalCount: result.data?.totalCount ?? 0,
    isLoading: result.isLoading,
    error: result.error,
  };
}
