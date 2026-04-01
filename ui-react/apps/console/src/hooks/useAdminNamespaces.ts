import { useQuery } from "@tanstack/react-query";
import {
  getNamespacesAdmin as getNamespacesAdminSdk,
  type GetNamespacesAdminData,
  type Namespace,
} from "../client";
import {
  getNamespacesAdminQueryKey,
  getNamespaceAdminOptions,
} from "../client/@tanstack/react-query.gen";
import { paginatedQueryFn, type PaginatedResult } from "../api/pagination";
import { useAuthStore } from "../stores/authStore";
import { isSdkError } from "../api/errors";

function buildNameFilter(search: string): string {
  const filter = [
    {
      type: "property",
      params: { name: "name", operator: "contains", value: search },
    },
  ];
  return btoa(JSON.stringify(filter));
}

interface UseAdminNamespacesParams {
  page?: number;
  perPage?: number;
  search?: string;
}

export function useAdminNamespaces({
  page = 1,
  perPage = 10,
  search = "",
}: UseAdminNamespacesParams = {}) {
  const isAdmin = useAuthStore((s) => s.isAdmin);

  const query: GetNamespacesAdminData["query"] = { page, per_page: perPage };
  if (search) query.filter = buildNameFilter(search);
  const options = { query };

  const result = useQuery<PaginatedResult<Namespace>>({
    queryKey: getNamespacesAdminQueryKey(options),
    queryFn: paginatedQueryFn(getNamespacesAdminSdk, options),
    enabled: isAdmin,
    staleTime: 5 * 60 * 1000, // 5 minutes
    retry: (count, err) =>
      isSdkError(err) && err.status === 401 ? false : count < 1,
    refetchOnWindowFocus: false,
  });

  return {
    namespaces: result.data?.data ?? [],
    totalCount: result.data?.totalCount ?? 0,
    isLoading: result.isLoading,
    error: result.error,
    refetch: result.refetch,
  };
}

export function useAdminNamespace(tenantId: string) {
  const isAdmin = useAuthStore((s) => s.isAdmin);

  return useQuery({
    ...getNamespaceAdminOptions({ path: { tenant: tenantId } }),
    enabled: isAdmin && !!tenantId,
    staleTime: 5 * 60 * 1000, // 5 minutes
    retry: (count, err) =>
      isSdkError(err) && err.status === 401 ? false : count < 1,
    refetchOnWindowFocus: false,
  });
}
