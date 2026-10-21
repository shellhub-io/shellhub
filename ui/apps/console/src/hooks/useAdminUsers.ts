import { useQuery } from "@tanstack/react-query";
import {
  getUsers as getUsersSdk,
  getUsersQueryKey,
  getUserOptions,
  type GetUsersData,
  type UserAdminResponse,
} from "../client";
import { paginatedQueryFn, type PaginatedResult } from "../api/pagination";
import { PER_PAGE } from "@/utils/pagination";
import { useAuthStore } from "../stores/authStore";
import { adminQueryDefaults } from "./useAdminRequests";
import { toBase64Json } from "@/utils/encoding";

/**
 * A subset of the user list the admin can narrow to: accounts waiting for an admin's approval, the
 * instance admins themselves, or sign-ups that never confirmed their email. The empty string is
 * every user.
 */
export type AdminUserSubset =
  "" | "awaiting_approval" | "admin" | "not_confirmed";

type FilterProperty = {
  type: "property";
  params: { name: string; operator: string; value: unknown };
};

const SUBSET_FILTERS: Record<Exclude<AdminUserSubset, "">, FilterProperty> = {
  awaiting_approval: {
    type: "property",
    params: { name: "awaiting_approval", operator: "bool", value: true },
  },
  admin: {
    type: "property",
    params: { name: "admin", operator: "bool", value: true },
  },
  not_confirmed: {
    type: "property",
    params: { name: "status", operator: "eq", value: "not-confirmed" },
  },
};

function contains(name: string, value: string): FilterProperty {
  return { type: "property", params: { name, operator: "contains", value } };
}

function usersSearchFields(
  subset: AdminUserSubset,
): ("name" | "username" | "email")[] {
  return subset ? ["username"] : ["name", "username", "email"];
}

/**
 * The fields a search over this subset matches, as a phrase for the search box ("name, username
 * or email"). Every user is searched by all three; a subset narrows the search to username. The
 * phrase and the filter come from the same list, so they cannot disagree.
 */
export function usersSearchScope(subset: AdminUserSubset): string {
  const fields = usersSearchFields(subset);
  return fields.length > 1
    ? `${fields.slice(0, -1).join(", ")} or ${fields[fields.length - 1]}`
    : fields[0];
}

function buildUsersFilter(
  search: string,
  subset: AdminUserSubset,
): string | undefined {
  const matches = search
    ? usersSearchFields(subset).map((field) => contains(field, search))
    : [];
  if (!subset) return matches.length > 0 ? toBase64Json(matches) : undefined;
  if (matches.length === 0) return toBase64Json([SUBSET_FILTERS[subset]]);
  return toBase64Json([
    { type: "operator", params: { name: "and" } },
    ...matches,
    SUBSET_FILTERS[subset],
  ]);
}

interface AdminUsersQueryParams {
  page?: number;
  perPage?: number;
  search?: string;
  subset?: AdminUserSubset;
}

/**
 * Query options for a page of users for the admin list. Nothing is requested unless isAdmin, and a
 * 401 is never retried.
 */
export function adminUsersQueryOptions(
  { page = 1, perPage = PER_PAGE, search = "", subset = "" }: AdminUsersQueryParams,
  isAdmin: boolean,
) {
  const query: GetUsersData["query"] = { page, per_page: perPage };
  const filter = buildUsersFilter(search, subset);
  if (filter) query.filter = filter;
  const options = { query };

  return {
    queryKey: getUsersQueryKey(options),
    queryFn: paginatedQueryFn(getUsersSdk, options),
    enabled: isAdmin,
    ...adminQueryDefaults,
  };
}

/**
 * A page of users for the admin list. Admin-only, so it does not run for anyone else.
 */
export function useAdminUsers(params: AdminUsersQueryParams = {}) {
  const isAdmin = useAuthStore((s) => s.isAdmin);
  const result = useQuery<PaginatedResult<UserAdminResponse>>(
    adminUsersQueryOptions(params, isAdmin),
  );

  return {
    users: result.data?.data ?? [],
    totalCount: result.data?.totalCount ?? 0,
    isLoading: result.isLoading,
    error: result.error,
    isError: result.isError,
  };
}

/**
 * One user by id. An empty id issues no request, which is how a caller skips the lookup.
 */
export function useAdminUser(id: string) {
  const isAdmin = useAuthStore((s) => s.isAdmin);

  return useQuery({
    ...getUserOptions({ path: { id } }),
    enabled: isAdmin && !!id,
    ...adminQueryDefaults,
  });
}
