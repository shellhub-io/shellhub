import { useQueries, useQuery } from "@tanstack/react-query";
import {
  getNamespacesAdmin as getNamespacesAdminSdk,
  getNamespacesAdminQueryKey,
  getNamespaceAdminOptions,
  type GetNamespacesAdminData,
  type Namespace,
} from "../client";
import { paginatedQueryFn, type PaginatedResult } from "../api/pagination";
import { useAuthStore } from "../stores/authStore";
import { isSdkError } from "../api/errors";
import { toBase64Json } from "@/utils/encoding";

function buildNameFilter(search: string): string {
  const filter = [
    {
      type: "property",
      params: { name: "name", operator: "contains", value: search },
    },
  ];
  return toBase64Json(filter);
}

interface UseAdminNamespacesParams {
  page?: number;
  perPage?: number;
  search?: string;
}

/**
 * A page of namespaces for the admin list. Admin-only, so it does not run for anyone else.
 */
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
    isError: result.isError,
    refetch: result.refetch,
  };
}

/**
 * One namespace by tenant id, for the admin detail view.
 */
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

/**
 * The namespaces each user awaiting approval was added to, by user id, read from the details of
 * up to scanLimit namespaces. A user added to a namespace beyond that is missing from the map.
 */
export function useAwaitingMemberNamespaces(scanLimit: number) {
  const isAdmin = useAuthStore((s) => s.isAdmin);
  const { namespaces } = useAdminNamespaces({ perPage: scanLimit });

  return useQueries({
    queries: namespaces.map((namespace) => ({
      ...getNamespaceAdminOptions({ path: { tenant: namespace.tenant_id } }),
      enabled: isAdmin,
      staleTime: 5 * 60 * 1000,
      retry: (count: number, err: unknown) =>
        isSdkError(err) && err.status === 401 ? false : count < 1,
      refetchOnWindowFocus: false,
    })),
    combine: (results) => {
      const joiningByUser = new Map<string, string[]>();
      for (const { data } of results) {
        if (!data) continue;
        for (const member of data.members) {
          if (!member.id || !member.awaiting_approval) continue;
          joiningByUser.set(member.id, [
            ...(joiningByUser.get(member.id) ?? []),
            data.name,
          ]);
        }
      }
      return joiningByUser;
    },
  });
}
