import { useQueries, useQuery } from "@tanstack/react-query";
import {
  getNamespacesAdmin as getNamespacesAdminSdk,
  getNamespacesAdminQueryKey,
  getNamespaceAdminOptions,
  type GetNamespacesAdminData,
  type Namespace,
} from "../client";
import { paginatedQueryFn, type PaginatedResult } from "../api/pagination";
import { PER_PAGE } from "@/utils/pagination";
import { useAuthStore } from "../stores/authStore";
import { adminQueryDefaults } from "./useAdminRequests";
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

interface AdminNamespacesQueryParams {
  page?: number;
  perPage?: number;
  search?: string;
}

/**
 * Query options for a page of namespaces for the admin list. Nothing is requested unless isAdmin,
 * and a 401 is never retried.
 */
export function adminNamespacesQueryOptions(
  { page = 1, perPage = PER_PAGE, search = "" }: AdminNamespacesQueryParams,
  isAdmin: boolean,
) {
  const query: GetNamespacesAdminData["query"] = { page, per_page: perPage };
  if (search) query.filter = buildNameFilter(search);
  const options = { query };

  return {
    queryKey: getNamespacesAdminQueryKey(options),
    queryFn: paginatedQueryFn(getNamespacesAdminSdk, options),
    enabled: isAdmin,
    ...adminQueryDefaults,
  };
}

/**
 * A page of namespaces for the admin list. Admin-only, so it does not run for anyone else.
 */
export function useAdminNamespaces(params: AdminNamespacesQueryParams = {}) {
  const isAdmin = useAuthStore((s) => s.isAdmin);
  const result = useQuery<PaginatedResult<Namespace>>(
    adminNamespacesQueryOptions(params, isAdmin),
  );

  return {
    namespaces: result.data?.data ?? [],
    totalCount: result.data?.totalCount ?? 0,
    isLoading: result.isLoading,
    error: result.error,
    isError: result.isError,
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
    ...adminQueryDefaults,
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
      ...adminQueryDefaults,
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
